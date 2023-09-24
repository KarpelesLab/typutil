package typutil

import "math"

type op struct {
	opf func(float64, float64) float64
	opu func(uint64, uint64) uint64
	opi func(int64, int64) int64
}

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

// Math performs a mathematical operation on two variables of any type and returns a numeric type
// and true if everything went fine.
//
// For example Math("+", 40.000, "2") will return float64(42)
//
// For now if passed any float value, the function will always return a float value in the end. This
// may change in the future.
func Math(mathop string, a, b any) (any, bool) {
	op, ok := mathOps[mathop]
	if !ok {
		// invalid math op
		return 0, false
	}
	na, oka := AsNumber(a)
	nb, okb := AsNumber(b)
	ok = oka && okb

	switch ta := na.(type) {
	case uint64:
		switch tb := nb.(type) {
		case uint64:
			return op.opu(ta, tb), ok
		case int64:
			if tb > 0 {
				return op.opu(ta, uint64(tb)), ok
			} else {
				return op.opi(int64(ta), tb), ok
			}
		case float64:
			return op.opf(float64(ta), tb), ok
		default:
			return 0, false
		}
	case int64:
		switch tb := nb.(type) {
		case int64:
			return op.opi(ta, tb), ok
		case uint64:
			if ta > 0 {
				return op.opu(uint64(ta), tb), ok
			} else {
				return op.opi(ta, int64(tb)), ok
			}
		case float64:
			return op.opf(float64(ta), tb), ok
		default:
			return 0, false
		}
	case float64:
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
		return 0, false
	}
}
