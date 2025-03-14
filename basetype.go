package typutil

import (
	"encoding/json"
	"reflect"
)

// BaseType unwraps values to their underlying primitive types.
//
// It performs the following conversions:
// - For custom types (e.g., `type CustomString string`), returns the underlying primitive (string)
// - For json.RawMessage, attempts to unmarshal the content
// - For reflect.Value, extracts the actual value
// - For pointers and interfaces, dereferences them to their underlying values
// - For numeric types, converts to their Go primitives (int64, uint64, float64, etc.)
//
// Example:
//
//	type MyString string
//	var x MyString = "hello"
//	result := BaseType(x) // returns "hello" as a string, not MyString
//
// This is useful when working with custom types and wanting to normalize them to standard Go types.
func BaseType(v any) any {
	// Handle special types first with direct type assertions
	switch o := v.(type) {
	case json.RawMessage:
		// Try to unmarshal JSON data into a more concrete type
		json.Unmarshal(o, &v)
	case reflect.Value:
		// Extract the actual value from reflect.Value
		v = o.Interface()
	}

	// Use reflection to get the value's kind and convert appropriately
	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Bool:
		// Convert to native bool
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Convert all integer types to int64
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// Convert all unsigned integer types to uint64
		return val.Uint()
	case reflect.Float32, reflect.Float64:
		// Convert all float types to float64
		return val.Float()
	case reflect.Complex64, reflect.Complex128:
		// Convert complex numbers to complex128
		return val.Complex()
	case reflect.Array, reflect.Slice:
		// Keep arrays and slices as is
		// Note: Special case for []byte could be handled here if needed
		return v
	case reflect.Interface, reflect.Pointer:
		// Recursively unwrap interfaces and pointers
		return BaseType(val.Elem().Interface())
	case reflect.String:
		// Convert to native string
		return val.String()
	default:
		// For maps, structs, and other types, return as is
		return v
	}
}
