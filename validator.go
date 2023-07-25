package typutil

import (
	"fmt"
	"reflect"
	"strings"
)

type validatorObject struct {
	fnc reflect.Value
	arg reflect.Type
}

var validators = map[string]*validatorObject{}

func init() {
	SetValidator("notempty", validateNotEmpty)
}

// A validator func is a function that takes one argument (the value being validated) and returns either nil or an error
// If the function accepts a modifiable value (a pointer for example) it might be possible to modify the value during validation

// SetValidator sets the given function as validator with the given name
func SetValidator[T any](validator string, fnc func(T) error) {
	vfnc := reflect.ValueOf(fnc)
	argt := reflect.TypeOf((*T)(nil)).Elem()

	validators[validator] = &validatorObject{fnc: vfnc, arg: argt}
}

func getValidators(s string) ([]*validatorObject, error) {
	if s == "" {
		return nil, nil
	}
	a := strings.Split(s, ",")
	res := make([]*validatorObject, 0, len(a))

	for _, v := range a {
		o, ok := validators[v]
		if !ok {
			return res, fmt.Errorf("validator not found: %s", a)
		}
		res = append(res, o)
	}

	return res, nil
}

func (v *validatorObject) run(val any) error {
	valT := reflect.New(v.arg)
	err := assignReflectValues(valT, reflect.ValueOf(val))
	if err != nil {
		return err
	}

	res := v.fnc.Call([]reflect.Value{valT})
	if res[0].IsNil() {
		return nil
	}
	return res[0].Interface().(error)
}

func (v *validatorObject) runReflectValue(val reflect.Value) error {
	valT := reflect.New(v.arg)
	err := assignReflectValues(valT, val)
	if err != nil {
		return err
	}

	res := v.fnc.Call([]reflect.Value{valT})
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
		// AsBool will return true if value is non zero, non empty
		if AsBool(v) {
			return nil
		}
		return ErrEmptyValue
	}
}
