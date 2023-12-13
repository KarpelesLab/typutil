package typutil

import (
	"context"
	"fmt"
)

type offsetGetter interface {
	OffsetGet(context.Context, string) (any, error)
}

type valueReader interface {
	ReadValue(ctx context.Context) (any, error)
}

// OffsetGet returns v[offset] dealing with various case of figure. ctx will be passed to some methods handling it
func OffsetGet(ctx context.Context, v any, offset string) (any, error) {
	switch a := v.(type) {
	case offsetGetter:
		return a.OffsetGet(ctx, offset)
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
	case valueReader: // keep this last
		nv, err := a.ReadValue(ctx)
		if err != nil {
			return nil, err
		}
		return OffsetGet(ctx, nv, offset)
	default:
		return nil, fmt.Errorf("unsupported type %T for offset fetching", v)
	}
}
