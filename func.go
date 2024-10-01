package typutil

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

type requiredArg int

const Required requiredArg = 1

type Callable struct {
	fn       reflect.Value
	cnt      int             // number of actual args
	ctxPos   int             // pos of ctx argument, or -1
	arg      []reflect.Type  // type used for the argument to the method
	def      []reflect.Value // default values
	variadic bool            // is the func's last argument a ...
	vartyp   reflect.Type
}

var (
	ctxTyp = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Func returns a [Callable] object for a func that accepts a context.Context and/or any
// number of arguments
func Func(method any) *Callable {
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		panic("static method not a method")
	}

	typ := v.Type()
	res := &Callable{fn: v, ctxPos: -1, cnt: typ.NumIn()}

	ni := res.cnt

	for i := 0; i < ni; i += 1 {
		in := typ.In(i)
		if in.Implements(ctxTyp) {
			if res.ctxPos != -1 {
				panic("method taking multiple ctx arguments")
			}
			res.ctxPos = i
			res.cnt -= 1
			continue
		}
		res.arg = append(res.arg, in)
	}

	if typ.IsVariadic() {
		res.variadic = true
		ln := len(res.arg)
		res.vartyp = res.arg[ln-1].Elem() // last argument is an array []...
		res.arg = res.arg[:ln-1]
	}

	return res
}

// WithDefaults sets default arguments for the given callable that will be used when a call doesn't have
// enough arguments to call the method. It will panic if there aren't enough arguments provided or if a
// value is not compatible with the function's expected argument type.
//
// Example:
//
//	func myFunc(a, b, c int) { ... }
//	f := Func(myFunc).WithDefaults(typutil.Required, typutil.Required, 42) // c defaults to 42
//	f.CallArg(context.Background(), 10, 20) // equivalent to myFunc(10, 20, 42)
func (s *Callable) WithDefaults(args ...any) *Callable {
	// build def
	if len(args) < s.cnt {
		panic("WithDefaults requires at least the same number of arguments as the function")
	}
	if len(args) > s.cnt && !s.variadic {
		panic(ErrTooManyArgs)
	}

	def := make([]reflect.Value, len(args))

	for argN, arg := range args {
		if arg == Required {
			// keep this as an invalid value
			continue
		}
		var argV reflect.Value
		if argN >= len(s.arg) {
			argV = reflect.New(s.vartyp)
		} else {
			argV = reflect.New(s.arg[argN])
		}
		err := AssignReflect(argV, reflect.ValueOf(arg))
		if err != nil {
			panic(err)
		}
		def[argN] = argV.Elem()
	}

	res := &Callable{}
	*res = *s
	res.def = def
	return res
}

// Call invokes the func without any argument. If the func expects some kind of argument, Call will attempt
// to get input_json from the context, and if obtained it will be parsed and passed as argument to the method.
func (s *Callable) Call(ctx context.Context) (any, error) {
	// call this function, typically fetching request body from the context via input_json
	if s.cnt > 0 {
		// grab input json, call json.Unmarshal on argV
		input, ok := ctx.Value("input_json").(json.RawMessage)
		if ok {
			if s.cnt > 1 {
				// if the method take multiple arguments, the json value must be an array. By using RawMessage we
				// ensure we only parse the array part here, and not the contents, so it can be parsed for each
				// relevant argument type directly
				var args []RawJsonMessage
				err := json.Unmarshal(input, &args)
				if err != nil {
					return nil, err
				}
				anyArgs := make([]any, len(args))
				for n, v := range args {
					anyArgs[n] = v
				}
				return s.CallArg(ctx, anyArgs...)
			}
			return s.CallArg(ctx, RawJsonMessage(input))
		}
	}

	return s.CallArg(ctx)
}

// CallArg calls the method with the specified arguments. If less arguments are provided than required, an error will be raised.
func (s *Callable) CallArg(ctx context.Context, arg ...any) (any, error) {
	if s.cnt == 0 {
		// no args, ignore any input
		var args []reflect.Value
		if s.ctxPos == 0 {
			args = append(args, reflect.ValueOf(ctx))
		}
		return s.parseResult(s.fn.Call(args))
	}
	if len(arg) < s.cnt && s.def == nil {
		// not enough arguments to cover cnt
		return nil, ErrMissingArgs
	}
	if len(arg) > s.cnt && !s.variadic {
		return nil, ErrTooManyArgs
	}
	// call this function but pass arg values
	var args []reflect.Value
	var ctxPos int

	if s.ctxPos != -1 {
		args = make([]reflect.Value, len(arg)+1)
		args[s.ctxPos] = reflect.ValueOf(ctx)
		ctxPos = s.ctxPos
	} else {
		args = make([]reflect.Value, len(arg))
		ctxPos = len(arg) + 1
	}

	for argN, v := range arg {
		var argV reflect.Value
		if argN >= len(s.arg) {
			argV = reflect.New(s.vartyp)
		} else {
			argV = reflect.New(s.arg[argN])
		}
		err := AssignReflect(argV, reflect.ValueOf(v))
		if err != nil {
			return nil, err
		}

		if argN >= ctxPos {
			args[argN+1] = argV.Elem()
		} else {
			args[argN] = argV.Elem()
		}
	}
	if len(args) < len(s.def) {
		add := s.def[len(args):]
		for _, v := range add {
			if !v.IsValid() {
				return nil, ErrMissingArgs
			}
		}
		args = append(args, add...)
	}

	return s.parseResult(s.fn.Call(args))
}

// IsStringArg returns true if the nth argument of the callable is a string, or a type related to string
func (s *Callable) IsStringArg(n int) bool {
	return s.ArgKind(n) == reflect.String
}

// ArgKind returns the kind for the nth argument. reflect.Invalid will be returned if there is no such argument
func (s *Callable) ArgKind(n int) reflect.Kind {
	if n >= len(s.arg) {
		return reflect.Invalid
	}
	return s.arg[n].Kind()
}

var errTyp = reflect.TypeOf((*error)(nil)).Elem()

func (s *Callable) parseResult(res []reflect.Value) (output any, err error) {
	// for each value in res, try to find which one is an error and which one is a result
	for _, v := range res {
		if v.Type().Implements(errTyp) {
			err, _ = v.Interface().(error)
			continue
		}
		output = v.Interface()
	}
	return
}

// Call will call the callable with the provided arguments, and cast the return type to the specified
// type automatically. If the return type is not of the correct type, an error will be returned, unless
// there was already an error.
func Call[T any](s *Callable, ctx context.Context, arg ...any) (T, error) {
	res, err := s.CallArg(ctx, arg...)
	if v, ok := res.(T); ok {
		return v, err
	} else if err == nil {
		err = fmt.Errorf("%w: %T", ErrDifferentType, res)
	}
	return reflect.New(reflect.TypeFor[T]()).Elem().Interface().(T), err
}
