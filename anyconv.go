package typutil

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"reflect"
	"strconv"
)

// AsBool converts any value to a boolean using an intuitive conversion strategy.
//
// Conversion rules:
// - bool: used directly
// - numbers: true if non-zero, false if zero
// - strings: true if non-empty and not "0", false otherwise
// - bytes/buffer: true if length > 1 or not "0", false otherwise
// - maps/slices/collections: true if not empty, false otherwise
// - nil: false
//
// This is useful when working with user inputs, configuration values,
// or any scenario where values of different types need to be interpreted as booleans.
func AsBool(v any) bool {
	v = BaseType(v)
	switch r := v.(type) {
	case bool:
		return r
	case int:
		return r != 0
	case int64:
		return r != 0
	case uint64:
		return r != 0
	case float64:
		return r != 0
	case *bytes.Buffer:
		if r.Len() > 1 {
			return true
		}
		if r.Len() == 0 || r.String() == "0" {
			return false
		}
		return true
	case string:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r == "0" {
			return false
		}
		return true
	case []byte:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r[0] == '0' {
			return false
		}
		return true
	case map[string]any:
		if len(r) > 0 {
			return true
		}
		return false
	case []any:
		if len(r) > 0 {
			return true
		}
		return false
	case url.Values:
		return len(r) > 0
	default:
		return false
	}
}

// AsInt converts any value to an int64 using flexible type conversion rules.
//
// It returns the converted value and a boolean indicating success (true) or failure (false).
//
// Conversion rules:
// - Integer types: directly converted to int64
// - Unsigned integers: converted to int64 (returns false if too large for int64)
// - Booleans: true → 1, false → 0
// - Floating point: rounded to nearest integer (returns false if not a whole number)
// - Strings: parsed as integers (returns false if not a valid integer)
// - Byte slices: converted to string and parsed
// - Byte buffers: contents parsed as integers
// - nil: returns 0
//
// This is useful for normalizing input data from various sources into consistent integer values.
func AsInt(v any) (int64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n&(1<<63) != 0 {
			return int64(n), false
		}
		return int64(n), true
	case uint:
		return int64(n), true
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case float32:
		x := math.Round(float64(n))
		y := int64(x)
		return y, float64(y) == x
	case float64:
		x := math.Round(n)
		y := int64(x)
		return y, float64(y) == x
	case string:
		res, err := strconv.ParseInt(n, 0, 64)
		return res, err == nil
	case []byte:
		res, err := strconv.ParseInt(string(n), 0, 64)
		return res, err == nil
	case *bytes.Buffer:
		return AsInt(n.String())
	case json.Number:
		return AsInt(string(n))
	case nil:
		return 0, true
	default:
		log.Printf("[number] failed to parse type %T", n)
	}

	return 0, false
}

// AsUint converts any value to a uint64 using flexible type conversion rules.
//
// It returns the converted value and a boolean indicating success (true) or failure (false).
//
// Conversion rules:
// - Integer types: converted to uint64 (returns false if negative)
// - Unsigned integers: directly converted to uint64
// - Booleans: true → 1, false → 0
// - Floating point: rounded to nearest integer (returns false if negative or not a whole number)
// - Strings: parsed as unsigned integers (returns false if not a valid unsigned integer)
// - nil: returns 0
//
// This is useful for normalizing input data from various sources into consistent unsigned integer values.
func AsUint(v any) (uint64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return uint64(n), n >= 0
	case int16:
		return uint64(n), n >= 0
	case int32:
		return uint64(n), n >= 0
	case int64:
		return uint64(n), n >= 0
	case int:
		return uint64(n), n >= 0
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case uint:
		return uint64(n), true
	case float32:
		if n < 0 {
			return 0, false
		}
		x := math.Round(float64(n))
		y := uint64(x)
		return y, float64(y) == x
	case float64:
		if n < 0 {
			return 0, false
		}
		x := math.Round(n)
		y := uint64(x)
		return y, float64(y) == x
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		res, err := strconv.ParseUint(n, 0, 64)
		return res, err == nil
	case json.Number:
		return AsUint(string(n))
	case nil:
		return 0, true
	}

	return 0, false
}

// AsFloat converts any value to a float64 using flexible type conversion rules.
//
// It returns the converted value and a boolean indicating success (true) or failure (false).
//
// Conversion rules:
// - Float types: directly converted to float64
// - Integer types: converted to equivalent float64
// - Unsigned integers: converted to equivalent float64
// - Strings: parsed as floating point numbers (returns false if not a valid number)
// - nil: returns 0.0
// - Other types: attempts conversion via AsInt as a fallback
//
// This is useful for normalizing input data from various sources into consistent floating point values.
func AsFloat(v any) (float64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uintptr:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		res, err := strconv.ParseFloat(n, 64)
		return res, err == nil
	case nil:
		return 0, true
	}

	res, ok := AsInt(v)
	return float64(res), ok
}

// AsNumber converts any value to the most appropriate numeric type (int64, uint64, or float64).
//
// It intelligently chooses the numeric type that best represents the input value:
// - Most integers are represented as int64
// - Large unsigned integers (that don't fit in int64) are represented as uint64
// - Decimal numbers are represented as float64
// - String representations of numbers are parsed to the appropriate type
//
// It returns the converted value (as interface{}) and a boolean indicating success (true) or failure (false).
//
// This is particularly useful when you need to convert a value to a number, but don't know
// exactly which numeric type would be most appropriate.
func AsNumber(v any) (any, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return uint64(n), true
	case uintptr:
		return uint64(n), true
	case uint:
		return int64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case nil:
		return int64(0), true
	case bool:
		if n {
			return int64(1), true
		} else {
			return int64(0), true
		}
	case string:
		if res, err := strconv.ParseInt(n, 0, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseUint(n, 0, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseFloat(n, 64); err == nil {
			return res, true
		}
		v, _ := AsNumber(AsBool(n))
		return v, false
	case *bytes.Buffer:
		if n.Len() > 100 {
			return nil, false
		}
		return AsNumber(n.String())
	default:
		// reflect for values that do not match directly
		rv := reflect.ValueOf(n)
		switch rv.Kind() {
		case reflect.Bool:
			if rv.Bool() {
				return int64(1), true
			} else {
				return int64(0), true
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return rv.Int(), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return rv.Uint(), true
		case reflect.Float32, reflect.Float64:
			return rv.Float(), true
		case reflect.String:
			return AsNumber(rv.String())
		}
	}

	return nil, false
}

// AsString converts any value to a string using flexible conversion rules.
//
// It returns the converted string and a boolean indicating success (true) or failure (false).
//
// Conversion rules:
// - String types: used directly
// - Byte slices and buffers: converted to strings
// - Numeric types: formatted as base-10 strings
// - Booleans: true → "1", false → "0"
// - Other types: uses fmt.Sprintf("%v", value) but returns false to indicate non-direct conversion
//
// This is useful when you need to display or serialize values of various types as strings.
func AsString(v any) (string, bool) {
	v = BaseType(v)
	switch s := v.(type) {
	case string:
		return s, true
	case []byte:
		return string(s), true
	case *bytes.Buffer:
		return s.String(), true
	case int64:
		return strconv.FormatInt(s, 10), true
	case int:
		return strconv.FormatInt(int64(s), 10), true
	case int32:
		return strconv.FormatInt(int64(s), 10), true
	case int16:
		return strconv.FormatInt(int64(s), 10), true
	case int8:
		return strconv.FormatInt(int64(s), 10), true
	case uint64:
		return strconv.FormatUint(s, 10), true
	case uint:
		return strconv.FormatUint(uint64(s), 10), true
	case uint32:
		return strconv.FormatUint(uint64(s), 10), true
	case uint16:
		return strconv.FormatUint(uint64(s), 10), true
	case uint8:
		return strconv.FormatUint(uint64(s), 10), true
	case bool:
		if s {
			return "1", true
		} else {
			return "0", true
		}
	default:
		return fmt.Sprintf("%v", v), false
	}
}

// AsByteArray converts any value to a byte slice ([]byte) using flexible conversion rules.
//
// It returns the converted byte slice and a boolean indicating success (true) or failure (false).
//
// Conversion rules:
// - Strings: converted to UTF-8 byte representation
// - Byte slices: returned directly
// - Buffer types: contents extracted as bytes
// - Numeric types: converted to their binary representation (big-endian)
// - Booleans: true → [1], false → [0]
// - nil: returns nil
// - Complex/Float types: binary representation using encoding/binary
// - Other types: string representation as bytes, but marked as non-direct conversion (false)
//
// This is useful for serialization, hashing, or when working with binary protocols.
func AsByteArray(v any) ([]byte, bool) {
	v = BaseType(v)
	switch s := v.(type) {
	case string:
		return []byte(s), true
	case []byte:
		return s, true
	case *bytes.Buffer:
		return s.Bytes(), true
	case interface{ Bytes() []byte }:
		return s.Bytes(), true
	case int64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(s))
		return buf, true
	case uint64:
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, s)
		return buf, true
	case int32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(s))
		return buf, true
	case uint32:
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, s)
		return buf, true
	case int16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(s))
		return buf, true
	case uint16:
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, s)
		return buf, true
	case int8:
		return []byte{byte(s)}, true
	case uint8:
		return []byte{byte(s)}, true
	case int:
		if math.MaxUint == math.MaxUint32 {
			// 32 bits int
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(s))
			return buf, true
		} else {
			// 64 bits int
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(s))
			return buf, true
		}
	case uint:
		if math.MaxUint == math.MaxUint32 {
			// 32 bits int
			buf := make([]byte, 4)
			binary.BigEndian.PutUint32(buf, uint32(s))
			return buf, true
		} else {
			// 64 bits int
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, uint64(s))
			return buf, true
		}
	case bool:
		if s {
			return []byte{1}, true
		} else {
			return []byte{0}, true
		}
	case nil:
		return nil, true
	case float32, float64, complex64, complex128:
		buf := &bytes.Buffer{}
		binary.Write(buf, binary.BigEndian, s)
		return buf.Bytes(), true
	default:
		return []byte(fmt.Sprintf("%v", v)), false
	}
}

// ToType converts a value to the same type as a reference value.
//
// It examines the type of the reference value (ref) and attempts to convert
// the input value (v) to that same type. This is useful when you need to ensure
// type compatibility between values.
//
// Parameters:
//   - ref: The reference value whose type will be used as the target type
//   - v: The value to convert to the target type
//
// Returns:
//   - The converted value with the same type as ref
//   - A boolean indicating success (true) or failure (false)
//
// Deprecated: Use the generic As[T](v) function instead, which provides type safety at compile time.
func ToType(ref, v any) (any, bool) {
	switch ref.(type) {
	case bool:
		return AsBool(v), true
	case int:
		return toTypeInt[int](v)
	case int8:
		return toTypeInt[int8](v)
	case int16:
		return toTypeInt[int16](v)
	case int32:
		return toTypeInt[int32](v)
	case int64:
		return toTypeInt[int64](v)
	case uint:
		return toTypeUint[uint](v)
	case uint8:
		return toTypeUint[uint8](v)
	case uint16:
		return toTypeUint[uint16](v)
	case uint32:
		return toTypeUint[uint32](v)
	case uint64:
		return toTypeUint[uint64](v)
	case uintptr:
		return toTypeUint[uintptr](v)
	case float32:
		return toTypeFloat[float32](v)
	case float64:
		return toTypeFloat[float64](v)
	case []byte:
		return AsByteArray(v)
	case string:
		return AsString(v)
	default:
		t := reflect.TypeOf(ref)
		switch t.Kind() {
		case reflect.Bool:
			return AsBool(v), true
		case reflect.Int:
			return toTypeInt[int](v)
		case reflect.Int8:
			return toTypeInt[int8](v)
		case reflect.Int16:
			return toTypeInt[int16](v)
		case reflect.Int32:
			return toTypeInt[int32](v)
		case reflect.Int64:
			return toTypeInt[int64](v)
		case reflect.Uint:
			return toTypeUint[uint](v)
		case reflect.Uint8:
			return toTypeUint[uint8](v)
		case reflect.Uint16:
			return toTypeUint[uint16](v)
		case reflect.Uint32:
			return toTypeUint[uint32](v)
		case reflect.Uint64:
			return toTypeUint[uint64](v)
		case reflect.Uintptr:
			return toTypeUint[uintptr](v)
		case reflect.Float32:
			return toTypeFloat[float32](v)
		case reflect.Float64:
			return toTypeFloat[float64](v)
		case reflect.String:
			return AsString(v)
		}

		v := reflect.ValueOf(v)
		if v.CanConvert(t) {
			return v.Convert(t).Interface(), true
		}

		return nil, false
	}
}

// toTypeInt is a generic helper function that converts any value to a signed integer type.
// It supports all signed integer types (int, int8, int16, int32, int64).
func toTypeInt[T Signed](v any) (T, bool) {
	// First convert to a numeric type using AsNumber
	n, ok := AsNumber(v)

	// Then convert to the specific signed integer type based on the numeric type
	switch xn := n.(type) {
	case int64:
		// Direct conversion from int64 to the target type
		return T(xn), ok
	case uint64:
		// Converting from uint64 to signed type (potential overflow for large values)
		return T(xn), ok
	case float64:
		// Converting from float64 to signed type (potential loss of precision)
		return T(xn), ok
	default:
		// Fallback for unsupported types
		return 0, false
	}
}

// toTypeUint is a generic helper function that converts any value to an unsigned integer type.
// It supports all unsigned integer types (uint, uint8, uint16, uint32, uint64, uintptr).
func toTypeUint[T Unsigned](v any) (T, bool) {
	// First convert to a numeric type using AsNumber
	n, ok := AsNumber(v)

	// Then convert to the specific unsigned integer type based on the numeric type
	switch xn := n.(type) {
	case int64:
		// Converting from int64 to unsigned type (negative values will wrap)
		return T(xn), ok
	case uint64:
		// Direct conversion from uint64 to the target type
		return T(xn), ok
	case float64:
		// Converting from float64 to unsigned type (potential loss of precision)
		return T(xn), ok
	default:
		// Fallback for unsupported types
		return 0, false
	}
}

// toTypeFloat is a generic helper function that converts any value to a floating-point type.
// It supports both float32 and float64 types.
func toTypeFloat[T ~float32 | ~float64](v any) (T, bool) {
	// First convert to a numeric type using AsNumber
	n, ok := AsNumber(v)

	// Then convert to the specific float type based on the numeric type
	switch xn := n.(type) {
	case int64:
		// Converting from int64 to float (exact for small integers)
		return T(xn), ok
	case uint64:
		// Converting from uint64 to float (potential precision loss for large values)
		return T(xn), ok
	case float64:
		// Converting from float64 to the target float type
		return T(xn), ok
	default:
		// Fallback for unsupported types
		return 0, false
	}
}
