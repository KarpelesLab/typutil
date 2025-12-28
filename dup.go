package typutil

import (
	"reflect"
	"unsafe"
)

// deepCloneContext tracks already-cloned pointers to handle circular references
// and preserve pointer identity (two pointers to the same value remain pointing
// to the same cloned value). The cache is keyed by (Type, pointer address) to
// correctly handle pointers of different types pointing to the same memory.
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

// DeepClone performs a deep duplication of the provided argument, returning a
// newly created independent copy. All nested pointers, slices, maps, and structs
// are recursively cloned.
//
// Circular references are handled correctly - if a structure contains pointers
// that form a cycle, the cloned structure will have equivalent cycles pointing
// to the cloned values.
//
// Struct fields tagged with `clone:"-"` are skipped during deep cloning and
// retain their shallow-copied values. This is useful for fields like database
// connections, mutexes, or other resources that should not be duplicated:
//
//	type MyStruct struct {
//	    Data    []byte
//	    Conn    *sql.DB `clone:"-"` // will not be deep cloned
//	}
func DeepClone[T any](v T) T {
	return DeepCloneReflect(reflect.ValueOf(v)).Interface().(T)
}

// DeepCloneReflect performs a deep duplication of the provided reflect.Value.
// See DeepClone for details on behavior.
func DeepCloneReflect(src reflect.Value) reflect.Value {
	ptrs := &deepCloneContext{}
	return deepCloneReflect(src, ptrs)
}

// deepCloneReflect is the internal recursive implementation of deep cloning.
// It handles each reflect.Kind appropriately and uses the ptrs cache to track
// already-cloned references for cycle detection and pointer identity preservation.
func deepCloneReflect(src reflect.Value, ptrs *deepCloneContext) reflect.Value {
	if !src.IsValid() {
		return src
	}

	switch src.Kind() {
	// Primitive types are immutable or passed by value - return as-is
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128, reflect.Func:
		return src

	// Strings are immutable in Go - no need to duplicate
	case reflect.String:
		return src

	// Slices: create new backing array and deep clone each element
	case reflect.Slice:
		if src.IsNil() {
			return reflect.New(src.Type()).Elem()
		}
		// Check cache first - same backing array should produce same clone
		// NOTE: multiple slices may share backing array with different len/cap
		ptr := src.Pointer()
		if r, ok := ptrs.get(src.Type(), ptr); ok {
			return r
		}
		size := src.Len()
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
		// Cache before recursing to handle self-referential structures
		ptrs.set(src.Type(), ptr, dst)
		for i := 0; i < size; i++ {
			dst.Index(i).Set(deepCloneReflect(src.Index(i), ptrs))
		}
		return dst

	// Arrays: fixed-size, create new array and deep clone each element
	case reflect.Array:
		size := src.Len()
		dst := reflect.New(src.Type()).Elem()
		for i := 0; i < size; i++ {
			dst.Index(i).Set(deepCloneReflect(src.Index(i), ptrs))
		}
		return dst

	// Maps: create new map and deep clone all keys and values
	case reflect.Map:
		if src.IsNil() {
			return reflect.New(src.Type()).Elem()
		}
		ptr := src.Pointer()
		if r, ok := ptrs.get(src.Type(), ptr); ok {
			return r
		}
		dst := reflect.MakeMap(src.Type())
		// Cache before iterating to handle maps containing themselves
		ptrs.set(src.Type(), ptr, dst)
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(deepCloneReflect(iter.Key(), ptrs), deepCloneReflect(iter.Value(), ptrs))
		}
		return dst

	// Pointers: create new pointer and deep clone the pointed-to value
	case reflect.Ptr:
		newPtr := reflect.New(src.Type()).Elem()
		if !src.IsNil() {
			ptr := src.Pointer()
			if r, ok := ptrs.get(src.Type(), ptr); ok {
				return r
			}
			newV := reflect.New(src.Type().Elem())
			newPtr.Set(newV)
			// Cache before recursing to handle circular references (e.g., linked lists)
			ptrs.set(src.Type(), ptr, newPtr)
			newV.Elem().Set(deepCloneReflect(src.Elem(), ptrs))
		}
		return newPtr

	// Interfaces: deep clone the underlying concrete value, preserving its type
	case reflect.Interface:
		newPtr := reflect.New(src.Type()).Elem()
		if !src.IsNil() {
			newPtr.Set(deepCloneReflect(src.Elem(), ptrs))
		}
		return newPtr

	// Structs: shallow copy first, then deep clone each field
	case reflect.Struct:
		structType := src.Type()
		dst := reflect.New(structType).Elem()
		// Shallow copy provides base values; we then selectively deep clone fields
		dst.Set(src)
		n := src.NumField()
		for i := 0; i < n; i++ {
			field := structType.Field(i)

			// Fields with `clone:"-"` tag retain their shallow-copied value
			if tag := field.Tag.Get("clone"); tag == "-" {
				continue
			}

			if !field.IsExported() {
				// Unexported fields require unsafe access to read/write
				dstField := dst.Field(i)
				val := deepCloneReflect(reflect.NewAt(dstField.Type(), unsafe.Pointer(dstField.UnsafeAddr())).Elem(), ptrs)
				reflect.NewAt(dstField.Type(), unsafe.Pointer(dstField.UnsafeAddr())).Elem().Set(val)
				continue
			}
			dst.Field(i).Set(deepCloneReflect(dst.Field(i), ptrs))
		}
		return dst

	// UnsafePointer and any unhandled types: shallow copy only
	case reflect.UnsafePointer:
		fallthrough
	default:
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src)
		return dst
	}
}
