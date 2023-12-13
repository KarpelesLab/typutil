package typutil

import "fmt"

// GetOffset returns v[offset] dealing with various case of figure
func GetOffset(v any, offset string) (any, error) {
	switch a := v.(type) {
	case map[string]any:
		return a[offset], nil
	case []any:
		// convert offset to int, ensure it is in range
		n, ok := AsUint(offset)
		if !ok {
			return nil, ErrBadOffset
		}
		if n < 0 || n >= uint64(len(a)) {
			// silent error
			return nil, nil
		}
		return a[n], nil
	default:
		return nil, fmt.Errorf("unsupported type %T for offset fetching", v)
	}
}
