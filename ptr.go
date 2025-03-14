package typutil

import "reflect"

// IsNil recursively checks if a value is nil, following pointers and interfaces to their underlying values.
//
// This function provides a more robust nil check than the builtin `== nil` comparison, which
// only works for direct nil values. IsNil can detect "deeply nested" nil values, such as:
// - A nil pointer or interface
// - A pointer to a nil pointer
// - An interface containing a nil pointer
// - A pointer to an interface containing a nil pointer, etc.
//
// It also correctly identifies nil channels, maps, slices, and functions.
//
// Example:
//
//	var x *string            // x is nil
//	IsNil(x)                 // returns true
//
//	var y **string           // y points to a nil *string
//	IsNil(y)                 // returns true
//
//	var z interface{} = x    // z is an interface containing a nil pointer
//	IsNil(z)                 // returns true
//
// For non-nillable types (like int, string, etc.), it always returns false.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	return isNilReflect(reflect.ValueOf(v))
}

// isNilReflect is an internal helper function that uses reflection to recursively check
// if a reflect.Value is nil or contains a nil value.
func isNilReflect(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Slice:
		// These types have a built-in IsNil method that checks if they are nil
		return v.IsNil()
	case reflect.Ptr, reflect.Interface:
		// For pointers and interfaces, first check if they are nil themselves
		if v.IsNil() {
			return true
		}
		// If not nil, recursively check what they point to
		return isNilReflect(v.Elem())
	default:
		// Other types (int, string, struct, etc.) can never be nil
		return false
	}
}

// Flatten unwraps a value by removing all layers of pointers and interfaces,
// returning the underlying value.
//
// This is particularly useful when working with values that might be wrapped in multiple
// layers of pointers or interfaces, and you want to get to the actual value.
//
// Examples:
//
//	s := "hello"
//	ptr := &s
//	ptrptr := &ptr
//	Flatten(ptrptr)      // returns "hello" as a string
//
//	var nilPtr *string = nil
//	Flatten(nilPtr)      // returns nil
//
//	var iface interface{} = &s
//	Flatten(iface)       // returns "hello" as a string
//
// The function preserves the nil status of values - if any pointer in the chain is nil,
// nil will be returned.
func Flatten(a any) any {
	if a == nil {
		return a
	}
	return flattenReflect(reflect.ValueOf(a))
}

// flattenReflect is an internal helper function that uses reflection to recursively
// unwrap reflect.Value that might be a pointer or interface, getting to the underlying value.
func flattenReflect(a reflect.Value) any {
	switch a.Kind() {
	case reflect.Ptr, reflect.Interface:
		// For pointers and interfaces, check if they are nil first
		if a.IsNil() {
			return nil
		}
		// If not nil, recursively unwrap what they contain
		return flattenReflect(a.Elem())
	default:
		// For other types, just return the actual value
		return a.Interface()
	}
}
