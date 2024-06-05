package typutil

import (
	"reflect"
	"unsafe"
)

// DeepClone performs a deep duplication of the provided argument, and returns the newly created object
func DeepClone[T any](v T) T {
	return DeepCloneReflect(reflect.ValueOf(v)).Interface().(T)
}

// DeepCloneReflect performs a deep duplication of the provided reflect.Value
func DeepCloneReflect(src reflect.Value) reflect.Value {
	ptrs := make(map[uintptr]reflect.Value)
	return deepCloneReflect(src, ptrs)
}

func deepCloneReflect(src reflect.Value, ptrs map[uintptr]reflect.Value) reflect.Value {
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
		ptr := src.Pointer()
		if r, ok := ptrs[ptr]; ok {
			return r
		}
		// duplicate the value
		size := src.Len()
		dst := reflect.MakeSlice(src.Type(), size, size)
		for i := 0; i < size; i++ {
			dst.Index(i).Set(deepCloneReflect(src.Index(i), ptrs))
		}
		ptrs[ptr] = dst
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
		if r, ok := ptrs[ptr]; ok {
			return r
		}
		dst := reflect.MakeMap(src.Type())
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(deepCloneReflect(iter.Key(), ptrs), deepCloneReflect(iter.Value(), ptrs))
		}
		ptrs[ptr] = dst
		return dst
	case reflect.Ptr:
		newPtr := reflect.New(src.Type()).Elem()
		if !src.IsNil() {
			ptr := src.Pointer()
			if r, ok := ptrs[ptr]; ok {
				return r
			}
			// generate a new target for value
			newV := reflect.New(src.Type().Elem())
			newV.Elem().Set(deepCloneReflect(src.Elem(), ptrs))
			newPtr.Set(newV)
			ptrs[ptr] = newPtr
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
		n := src.NumField()
		for i := 0; i < n; i += 1 {
			if !src.Type().Field(i).IsExported() {
				// accessing unexported fields will cause panic
				//log.Printf("type = %s", dst.Field(i).Type())
				field := dst.Field(i)
				val := deepCloneReflect(reflect.NewAt(field.Type(), unsafe.Pointer(src.Field(i).UnsafeAddr())).Elem(), ptrs)
				reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(val)
				continue
			}
			dst.Field(i).Set(deepCloneReflect(src.Field(i), ptrs))
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
