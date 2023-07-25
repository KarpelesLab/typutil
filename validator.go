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

func init() {
	SetValidator("notempty", validateNotEmpty)
}

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

type fieldValidator struct {
	fld  int // field index
	vals []*validatorObject
}

type structValidator []*fieldValidator

var (
	validatorCache   = make(map[reflect.Type]structValidator)
	validatorCacheLk sync.Mutex
)

func getValidatorForType(t reflect.Type) structValidator {
	validatorCacheLk.Lock()
	defer validatorCacheLk.Unlock()

	val, ok := validatorCache[t]
	if ok {
		return val
	}

	n := t.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		vals, err := getValidators(f.Tag.Get("validator"))
		if err != nil {
			// skip
			continue
		}
		if len(vals) == 0 {
			continue
		}
		val = append(val, &fieldValidator{fld: i, vals: vals})
	}
	validatorCache[t] = val
	return val
}

func (sv structValidator) validate(val reflect.Value) error {
	var err error
	for _, vd := range sv {
		f := val.Field(vd.fld).Addr()
		for _, sub := range vd.vals {
			err = sub.runReflectValue(f)
			if err != nil {
				return err
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

// getValidators returns the validator objects for a given validator tag value. Multiple validators can be defined
func getValidators(s string) ([]*validatorObject, error) {
	if s == "" {
		return nil, nil
	}
	a := strings.Split(s, ",")
	res := make([]*validatorObject, 0, len(a))

	validatorsLk.RLock()
	defer validatorsLk.RUnlock()

	for _, v := range a {
		o, ok := validators[v]
		if !ok {
			return res, fmt.Errorf("validator not found: %s", a)
		}
		res = append(res, o)
	}

	return res, nil
}

func (v *validatorObject) runReflectValue(val reflect.Value) error {
	valT := reflect.New(v.arg)
	err := assignReflectValues(valT, val)
	if err != nil {
		return err
	}

	res := v.fnc.Call([]reflect.Value{valT.Elem()})
	if res[0].IsNil() {
		return nil
	}
	return res[0].Interface().(error)
}

func validateNotEmpty(v any) error {
	switch t := v.(type) {
	case string:
		if len(t) == 0 {
			return ErrEmptyValue
		}
		return nil
	default:
		s := reflect.ValueOf(v)
		if s.Kind() == reflect.Pointer {
			return validateNotEmpty(s.Elem().Interface())
		}
		// AsBool will return true if value is non zero, non empty
		if AsBool(v) {
			return nil
		}
		return ErrEmptyValue
	}
}
