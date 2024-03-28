package typutil

import "reflect"

// IsNil recursively checks if v is nil, even if it is a pointer of an interface of a ...
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	return isNilReflect(reflect.ValueOf(v))
}

func isNilReflect(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return true
		}
		return isNilReflect(v.Elem())
	default:
		return false
	}
}

// Flatten transforms a into a simple interface, even if it is contains multiple levels. *string becomes string (or nil if it's nil), etc
func Flatten(a any) any {
	if a == nil {
		return a
	}
	return flattenReflect(reflect.ValueOf(a))
}

func flattenReflect(a reflect.Value) any {
	switch a.Kind() {
	case reflect.Ptr, reflect.Interface:
		if a.IsNil() {
			return nil
		}
		return flattenReflect(a.Elem())
	default:
		return a.Interface()
	}
}
