package typutil

import "reflect"

func init() {
	SetValidator("notempty", validateNotEmpty)
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
