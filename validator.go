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

// A validator func is a function that takes one argument (the value being validated) and returns either nil or an error
// If the function accepts a modifiable value (a pointer for example) it might be possible to modify the value during validation

// SetValidator sets the given function as validator with the given name. This should be typically called in init()
func SetValidator[T any](validator string, fnc func(T) error) {
	vfnc := reflect.ValueOf(fnc)
	argt := reflect.TypeOf((*T)(nil)).Elem()

	validatorsLk.Lock()
	defer validatorsLk.Unlock()

	validators[validator] = &validatorObject{fnc: vfnc, arg: argt}
}

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

// Validate accept any struct as argument and returns if the struct is valid. The parameter should be a pointer
// to the struct so validators can edit values.
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
