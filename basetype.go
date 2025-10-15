// Package typutil provides flexible type conversion and function calling utilities for Go.
//
// This library offers several key features:
//
// 1. Type Conversion:
//   - Convert between different Go types with automatic handling of common conversions
//   - Support for primitive types, structs, slices, maps, and custom types
//   - Convert between structs with different field types but matching names
//
// 2. Function Wrapping:
//   - Wrap Go functions to support flexible argument handling
//   - Automatic context detection and passing
//   - Default parameter values
//   - Type conversion for function arguments
//   - JSON input handling
//
// 3. Validation:
//   - Validate struct field values using tag-based validators
//   - Register custom validators for specific types
//   - Support for complex validation rules with parameters
//
// 4. Pointer and Interface Utilities:
//   - Check for nil values in interfaces and pointers
//   - Flatten nested pointers
//   - Create pointers to values
//
// The package is designed to make working with dynamic types and function calls
// in Go easier and more flexible, while maintaining type safety where possible.
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
		if val.IsNil() {
			return nil
		}
		return BaseType(val.Elem().Interface())
	case reflect.String:
		// Convert to native string
		return val.String()
	default:
		// For maps, structs, and other types, return as is
		return v
	}
}
