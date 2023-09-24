package typutil

import "bytes"

// Equal returns true if a and b are somewhat equal
func Equal(a, b any) bool {
	// this is an approximate equal method, a==b can be true even if both aren't exactly the same type
	if a == nil {
		if b == nil {
			return true
		}
		// reverse things so a is not nil
		a, b = b, a
	}

	if typePriority(a) < typePriority(b) {
		// if a has lower priority, reverse
		a, b = b, a
	}

	b, _ = ToType(a, b)

	switch av := a.(type) {
	case []byte:
		return bytes.Compare(av, b.([]byte)) == 0
	default:
		// hope this works, lol
		return a == b
	}
}

// typePriority returns a numeric value defining which type will have priority on the other
func typePriority(v any) int {
	switch v.(type) {
	case float32, float64:
		return 4
	case bool:
		return 3
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		return 3
	case string, []byte, *bytes.Buffer:
		return 2
	case nil:
		return -1
	default:
		return 1 // ??
	}
}
