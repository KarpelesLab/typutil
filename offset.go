package typutil

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
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
	case map[string]string:
		return a[offset], nil
	case url.Values:
		res := a[offset]
		if len(res) == 0 {
			return nil, nil
		} else {
			return res[0], nil
		}
	case []any:
		// convert offset to int, ensure it is in range
		n, ok := AsUint(offset)
		if !ok {
			return nil, fmt.Errorf("%w: %T", ErrBadOffset, offset)
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
		vr := reflect.ValueOf(v)
		switch vr.Kind() {
		case reflect.Map:
			switch vr.Type().Key().Kind() {
			case reflect.String:
				// this we can handle
				v := vr.MapIndex(reflect.ValueOf(offset))
				if v.IsZero() {
					return nil, nil
				} else {
					return v.Interface(), nil
				}
			}
		}
		return nil, fmt.Errorf("unsupported type %T for offset fetching", v)
	}
}
