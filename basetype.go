package typutil

import (
	"encoding/json"
	"reflect"

	"github.com/KarpelesLab/pjson"
)

// BaseType attempts to convert v into its base type, that is if v is a type
// that is defined as `type foo string`, a simple string will be returned.
func BaseType(v any) any {
	switch o := v.(type) {
	case json.RawMessage:
		json.Unmarshal(o, &v)
	case pjson.RawMessage:
		pjson.Unmarshal(o, &v)
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint()
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Complex64, reflect.Complex128:
		return val.Complex()
	case reflect.Array, reflect.Slice:
		// []byte ?
		return val.Slice(0, val.Len())
	case reflect.Interface, reflect.Pointer:
		return BaseType(val.Elem())
	case reflect.String:
		return val.String()
	default:
		// ??
		return v
	}
}
