package typutil

import (
	"reflect"
	"unsafe"
)

type deepCloneContext struct {
	cache map[reflect.Type]map[uintptr]reflect.Value
}

func (c *deepCloneContext) get(t reflect.Type, p uintptr) (reflect.Value, bool) {
	if c.cache == nil {
		return reflect.Value{}, false
	}
	if _, ok := c.cache[t]; !ok {
		return reflect.Value{}, false
	}
	if _, ok := c.cache[t][p]; !ok {
		return reflect.Value{}, false
	}
	return c.cache[t][p], true
}

func (c *deepCloneContext) set(t reflect.Type, p uintptr, v reflect.Value) {
	if c.cache == nil {
		c.cache = make(map[reflect.Type]map[uintptr]reflect.Value)
	}
	if _, ok := c.cache[t]; !ok {
		c.cache[t] = make(map[uintptr]reflect.Value)
	}
	c.cache[t][p] = v
}

// DeepClone performs a deep duplication of the provided argument, and returns the newly created object
func DeepClone[T any](v T) T {
	return DeepCloneReflect(reflect.ValueOf(v)).Interface().(T)
}

// DeepCloneReflect performs a deep duplication of the provided reflect.Value
func DeepCloneReflect(src reflect.Value) reflect.Value {
	ptrs := &deepCloneContext{}
	return deepCloneReflect(src, ptrs)
}

func deepCloneReflect(src reflect.Value, ptrs *deepCloneContext) reflect.Value {
	if !src.IsValid() {
		// invalid value → return as is
		return src
	}

	switch src.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Func:
		// can be used as is
		return src
	case reflect.String:
		// strings are not writable, no need to duplicate
		return src
	case reflect.Slice:
		if src.IsNil() {
			return reflect.New(src.Type()).Elem()
		}
		// TODO in case of slice, multiple slices may point to the same data but have different len
		ptr := src.Pointer()
		if r, ok := ptrs.get(src.Type(), ptr); ok {
			return r
		}
		// duplicate the value
		size := src.Len()
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		for i := 0; i < size; i++ {
			dst.Index(i).Set(deepCloneReflect(src.Index(i), ptrs))
		}
		ptrs.set(src.Type(), ptr, dst)
		return dst
	case reflect.Array:
		size := src.Len()
		dst := reflect.New(src.Type()).Elem()
		for i := 0; i < size; i++ {
			dst.Index(i).Set(deepCloneReflect(src.Index(i), ptrs))
		}
		return dst
	case reflect.Map:
		if src.IsNil() {
			return reflect.New(src.Type()).Elem()
		}
		ptr := src.Pointer()
		if r, ok := ptrs.get(src.Type(), ptr); ok {
			return r
		}
		dst := reflect.MakeMap(src.Type())
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(deepCloneReflect(iter.Key(), ptrs), deepCloneReflect(iter.Value(), ptrs))
		}
		ptrs.set(src.Type(), ptr, dst)
		return dst
	case reflect.Ptr:
		newPtr := reflect.New(src.Type()).Elem()
		if !src.IsNil() {
			ptr := src.Pointer()
			if r, ok := ptrs.get(src.Type(), ptr); ok {
				return r
			}
			// generate a new target for value
			newV := reflect.New(src.Type().Elem())
			newV.Elem().Set(deepCloneReflect(src.Elem(), ptrs))
			newPtr.Set(newV)
			ptrs.set(src.Type(), ptr, newPtr)
		}
		return newPtr
	case reflect.Interface:
		newPtr := reflect.New(src.Type()).Elem()
		if !src.IsNil() {
			// generate a new target for value
			newV := reflect.New(src.Elem().Type())
			newV.Elem().Set(deepCloneReflect(src.Elem(), ptrs))
			newPtr.Set(newV)
		}
		return newPtr
	case reflect.Struct:
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src) // first, make a shallow copy so we have a guaranteed addressable version of the struct
		n := src.NumField()
		for i := 0; i < n; i += 1 {
			if !src.Type().Field(i).IsExported() {
				// accessing unexported fields normally will cause panic, so we do it the not normal way
				field := dst.Field(i)
				val := deepCloneReflect(reflect.NewAt(field.Type(), unsafe.Pointer(dst.Field(i).UnsafeAddr())).Elem(), ptrs)
				reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(val)
				continue
			}
			dst.Field(i).Set(deepCloneReflect(dst.Field(i), ptrs))
		}
		return dst
	case reflect.UnsafePointer:
		fallthrough
	default:
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src)
		return dst
	}
}
