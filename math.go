package typutil

import "math"

// op represents a mathematical operation that can be performed on different numeric types.
// Each operation needs three implementations: for floating point, unsigned integers, and signed integers.
type op struct {
	opf func(float64, float64) float64 // Operation on floating point numbers
	opu func(uint64, uint64) uint64    // Operation on unsigned integers
	opi func(int64, int64) int64       // Operation on signed integers
}

// mathOps maps operation symbols to their implementations.
// Supported operations: +, -, *, /, ^, %, &, |
var mathOps = map[string]op{
	"+": op{
		opf: func(a, b float64) float64 { return a + b },
		opu: func(a, b uint64) uint64 { return a + b },
		opi: func(a, b int64) int64 { return a + b },
	},
	"-": op{
		opf: func(a, b float64) float64 { return a - b },
		opu: func(a, b uint64) uint64 { return a - b },
		opi: func(a, b int64) int64 { return a - b },
	},
	"/": op{
		opf: func(a, b float64) float64 { return a / b },
		opu: func(a, b uint64) uint64 { return a / b },
		opi: func(a, b int64) int64 { return a / b },
	},
	"*": op{
		opf: func(a, b float64) float64 { return a * b },
		opu: func(a, b uint64) uint64 { return a * b },
		opi: func(a, b int64) int64 { return a * b },
	},
	"^": op{
		opf: func(a, b float64) float64 { return math.NaN() },
		opu: func(a, b uint64) uint64 { return a ^ b },
		opi: func(a, b int64) int64 { return a ^ b },
	},
	"%": op{
		opf: func(a, b float64) float64 { return math.NaN() },
		opu: func(a, b uint64) uint64 { return a % b },
		opi: func(a, b int64) int64 { return a % b },
	},
	"&": op{
		opf: func(a, b float64) float64 { return math.NaN() },
		opu: func(a, b uint64) uint64 { return a & b },
		opi: func(a, b int64) int64 { return a & b },
	},
	"|": op{
		opf: func(a, b float64) float64 { return math.NaN() },
		opu: func(a, b uint64) uint64 { return a | b },
		opi: func(a, b int64) int64 { return a | b },
	},
}

// Math performs a mathematical operation on two values of any type and returns the result.
//
// The function works by:
// 1. Converting both inputs to numeric types using AsNumber()
// 2. Determining the correct type of operation based on the numeric types
// 3. Applying the appropriate operation and returning the result
//
// Parameters:
//   - mathop: The operation to perform as a string. Supported operations: "+", "-", "*", "/", "^", "%", "&", "|"
//   - a, b: The operands for the operation. Can be of any type that can be converted to a number
//
// Returns:
//   - The result of the operation as int64, uint64, or float64 depending on the inputs
//   - A boolean indicating success (true) or failure (false)
//
// Examples:
//
//	result, ok := Math("+", 40, 2)         // result = int64(42), ok = true
//	result, ok := Math("+", 40.5, 1.5)     // result = float64(42.0), ok = true
//	result, ok := Math("/", 84, 2)         // result = int64(42), ok = true
//	result, ok := Math("+", "40", "2")     // result = int64(42), ok = true
//	result, ok := Math("invalid", 1, 2)    // result = 0, ok = false
//
// Notes:
//   - If either input has a float type, the result will be a float64
//   - Division by zero will cause a panic - it's recommended to check for zero divisors before calling
//   - Bitwise operations (^, %, &, |) return NaN when operating on floats
func Math(mathop string, a, b any) (any, bool) {
	// Look up the requested operation
	op, ok := mathOps[mathop]
	if !ok {
		// Return failure if operation is not supported
		return 0, false
	}

	// Convert both operands to numeric types
	na, oka := AsNumber(a)
	nb, okb := AsNumber(b)
	ok = oka && okb // Both conversions must succeed

	// Apply the operation based on the specific numeric types
	// The logic ensures that:
	// 1. We use the correct operation for the numeric types
	// 2. We convert types appropriately when mixing different types
	// 3. We handle sign-sensitive operations carefully

	switch ta := na.(type) {
	case uint64:
		// First operand is unsigned
		switch tb := nb.(type) {
		case uint64:
			// Both operands are unsigned, use unsigned operation
			return op.opu(ta, tb), ok
		case int64:
			if tb > 0 {
				// Positive signed can be safely converted to unsigned
				return op.opu(ta, uint64(tb)), ok
			} else {
				// With negative second operand, convert first to signed
				return op.opi(int64(ta), tb), ok
			}
		case float64:
			// If either operand is float, result is float
			return op.opf(float64(ta), tb), ok
		default:
			return 0, false
		}
	case int64:
		// First operand is signed
		switch tb := nb.(type) {
		case int64:
			// Both operands are signed, use signed operation
			return op.opi(ta, tb), ok
		case uint64:
			if ta > 0 {
				// Positive signed can be safely converted to unsigned
				return op.opu(uint64(ta), tb), ok
			} else {
				// With negative first operand, convert second to signed
				return op.opi(ta, int64(tb)), ok
			}
		case float64:
			// If either operand is float, result is float
			return op.opf(float64(ta), tb), ok
		default:
			return 0, false
		}
	case float64:
		// First operand is float, convert second operand to float
		switch tb := nb.(type) {
		case int64:
			return op.opf(ta, float64(tb)), ok
		case uint64:
			return op.opf(ta, float64(tb)), ok
		case float64:
			return op.opf(ta, tb), ok
		default:
			return 0, false
		}
	default:
		// Unsupported type combination
		return 0, false
	}
}
