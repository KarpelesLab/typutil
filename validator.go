package typutil

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type validatorObject struct {
	fnc reflect.Value
	arg reflect.Type
}

var (
	validators   = map[string]*validatorObject{}
	validatorsLk sync.RWMutex
)

// A validator function takes one argument (the value being validated) and returns either nil or an error.
// If the function accepts a modifiable value (like a pointer), it can potentially modify the value during validation.
//
// Validators are registered by name and can be used in struct field tags with the "validator" key.
// For example:
//
//	type User struct {
//	    Name string `validator:"required"`
//	    Age  int    `validator:"min=18"`
//	}
//
// Multiple validators can be specified with commas:
//
//	Email string `validator:"required,email"`
//
// Validators can accept arguments after an equals sign:
//
//	Password string `validator:"minlength=8,maxlength=64"`

// SetValidator registers a typed validation function with the given name.
//
// This function provides a type-safe way to register validators. The validator
// function must accept a single argument of type T and return an error.
//
// Parameters:
//   - validator: The name of the validator (used in struct tags)
//   - fnc: The validation function that checks values of type T
//
// This function should typically be called in init() to register validators
// before they are used.
//
// Example:
//
//	func init() {
//	    // Register a validator that ensures strings are not empty
//	    SetValidator("required", func(s string) error {
//	        if s == "" {
//	            return errors.New("value is required")
//	        }
//	        return nil
//	    })
//
//	    // Register a validator that ensures integers are positive
//	    SetValidator("positive", func(i int) error {
//	        if i <= 0 {
//	            return errors.New("value must be positive")
//	        }
//	        return nil
//	    })
//	}
func SetValidator[T any](validator string, fnc func(T) error) {
	vfnc := reflect.ValueOf(fnc)
	argt := reflect.TypeOf((*T)(nil)).Elem()

	validatorsLk.Lock()
	defer validatorsLk.Unlock()

	validators[validator] = &validatorObject{fnc: vfnc, arg: argt}
}

// SetValidatorArgs registers a validation function that may accept additional arguments.
//
// Unlike SetValidator, this function accepts any function type as long as it takes
// at least one argument (the value to validate) and returns an error. Additional
// arguments can be specified in the validator tag after an equals sign.
//
// Parameters:
//   - validator: The name of the validator (used in struct tags)
//   - fnc: The validation function, which must:
//   - Take at least one argument (the value to validate)
//   - Return an error (or nil if validation passes)
//
// Example:
//
//	func init() {
//	    // Register a validator that ensures strings have a minimum length
//	    SetValidatorArgs("minlength", func(s string, minLen int) error {
//	        if len(s) < minLen {
//	            return fmt.Errorf("must be at least %d characters long", minLen)
//	        }
//	        return nil
//	    })
//	}
//
//	// Then use it in a struct tag:
//	type User struct {
//	    Password string `validator:"minlength=8"` // Password must be at least 8 characters
//	}
//
// Panics if:
//   - fnc is not a function
//   - fnc does not accept at least one argument
func SetValidatorArgs(validator string, fnc any) {
	vfnc := reflect.ValueOf(fnc)
	if vfnc.Kind() != reflect.Func {
		panic("not a function")
	}
	t := vfnc.Type()
	if t.NumIn() < 1 {
		panic("validator function must accept at least one argument")
	}
	argt := t.In(0)

	validators[validator] = &validatorObject{fnc: vfnc, arg: argt}
}

// getValidators returns the validator objects for a given validator tag value. Multiple validators can be defined
func getValidators(s string) ([]*validatorObject, [][]reflect.Value, error) {
	if s == "" {
		return nil, nil, nil
	}
	a := strings.Split(s, ",")
	res := make([]*validatorObject, 0, len(a))
	res2 := make([][]reflect.Value, 0, len(a))

	validatorsLk.RLock()
	defer validatorsLk.RUnlock()

	for _, v := range a {
		p := strings.IndexByte(v, '=')
		a := ""
		// allow arguments after =, such as maxlength=2
		if p != -1 {
			a = v[p+1:]
			v = v[:p]
		}
		o, ok := validators[v]
		if !ok {
			return res, res2, fmt.Errorf("validator not found: %s", a)
		}
		res = append(res, o)
		res2 = append(res2, o.convertArgs(a))
	}

	return res, res2, nil
}

type fieldValidator struct {
	fld  int    // field index
	name string // field name
	vals []*validatorObject
	args [][]reflect.Value // extra validator param, if any
}

type structValidator []*fieldValidator

var (
	validatorCache   = make(map[reflect.Type]structValidator)
	validatorCacheLk sync.Mutex
)

func getValidatorForType(t reflect.Type) structValidator {
	validatorCacheLk.Lock()
	defer validatorCacheLk.Unlock()

	if t.Kind() != reflect.Struct {
		return nil
	}

	val, ok := validatorCache[t]
	if ok {
		return val
	}

	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		vals, args, err := getValidators(f.Tag.Get("validator"))
		if err != nil {
			// skip
			continue
		}
		if len(vals) == 0 {
			continue
		}
		val = append(val, &fieldValidator{fld: i, name: f.Name, vals: vals, args: args})
	}
	validatorCache[t] = val
	return val
}

func (sv structValidator) validate(val reflect.Value) error {
	var err error
	for _, vd := range sv {
		f := val.Field(vd.fld).Addr()
		for n, sub := range vd.vals {
			err = sub.runReflectValue(f, vd.args[n])
			if err != nil {
				return fmt.Errorf("on field %s: %w", vd.name, err)
			}
		}
	}
	return nil
}

// Validate checks if a struct meets all validation rules defined in its field tags.
//
// This function processes all fields in a struct that have "validator" tags and
// runs the appropriate validation functions on them. If any validation fails,
// an error is returned with details about which field failed and why.
//
// Parameters:
//   - obj: A pointer to the struct to validate. Using a pointer is required so that
//     validators can potentially modify values during validation.
//
// Returns:
//   - nil if all validations pass
//   - An error if any validation fails, formatted as "on field X: error details"
//   - ErrStructPtrRequired if obj is not a pointer to a struct
//
// Example:
//
//	type User struct {
//	    Name     string `validator:"required"`
//	    Email    string `validator:"required,email"`
//	    Password string `validator:"minlength=8"`
//	}
//
//	user := &User{Name: "Alice", Email: "invalid", Password: "123"}
//	err := Validate(user) // Returns error: "on field Email: invalid email format"
//
// Validations are applied in the order they appear in the tag, from left to right.
// If a field has no validator tag or the tag is empty, it is not validated.
func Validate(obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		return ErrStructPtrRequired
	}
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ErrStructPtrRequired
	}

	return getValidatorForType(v.Type()).validate(v)
}

func (v *validatorObject) runReflectValue(val reflect.Value, args []reflect.Value) error {
	valT := reflect.New(v.arg)
	err := AssignReflect(valT, val)
	if err != nil {
		return err
	}

	res := v.fnc.Call(append([]reflect.Value{valT.Elem()}, args...))
	if res[0].IsNil() {
		return nil
	}
	return res[0].Interface().(error)
}

func (v *validatorObject) convertArgs(args string) []reflect.Value {
	t := v.fnc.Type()
	if t.NumIn() <= 1 {
		// 0 shouldn't happen, 1 means there are no extra args to take into account
		return nil
	}
	argsArray := strings.Split(args, ",")
	extraCnt := t.NumIn() - 1
	if len(argsArray) < extraCnt {
		// not enough args
		extraCnt = len(argsArray)
	}
	res := make([]reflect.Value, 0, extraCnt)

	for i := 0; i < extraCnt; i++ {
		argt := t.In(i + 1)
		v := reflect.New(argt).Elem()
		AssignReflect(v, reflect.ValueOf(argsArray[i]))
		res = append(res, v)
	}
	return res
}
