package typutil

import (
	"encoding"
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type assignFunc func(dst, src reflect.Value) error

type assignConvType struct {
	dst reflect.Type
	src reflect.Type
}

var assignFuncCache sync.Map // map[assignConvType]assignFunc

type valueScanner interface {
	Scan(any) error
}

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	valueScannerType    = reflect.TypeOf((*valueScanner)(nil)).Elem()
)

func getAssignFunc(dstt reflect.Type, srct reflect.Type) assignFunc {
	if dstt == srct {
		return simpleSet
	}

	act := assignConvType{dstt, srct}
	if fi, ok := assignFuncCache.Load(act); ok {
		return fi.(assignFunc)
	}

	// deal with recursive type the same way encoding/json does
	var (
		wg sync.WaitGroup
		f  assignFunc
	)
	wg.Add(1)

	fi, loaded := assignFuncCache.LoadOrStore(act, assignFunc(func(dst, src reflect.Value) error {
		wg.Wait()
		return f(dst, src)
	}))
	if loaded {
		return fi.(assignFunc)
	}

	// compute real func
	f = newAssignFunc(dstt, srct)
	wg.Done()
	assignFuncCache.Store(act, f)
	return f
}

func Assign(dst, src any) error {
	// grab dst value
	vdst := reflect.ValueOf(dst)
	// check if pointer (required)
	if vdst.Kind() != reflect.Pointer || vdst.IsNil() {
		return ErrAssignDestNotPointer
	}
	// grab source
	vsrc := reflect.ValueOf(src)
	if vsrc.Kind() == reflect.Interface {
		vsrc = vsrc.Elem()
	}

	// do the thing
	f := getAssignFunc(vdst.Type(), vsrc.Type())
	if f == nil {
		return fmt.Errorf("%w: %T to %T", ErrAssignImpossible, src, dst)
	}
	return f(vdst, vsrc)
}

func assignReflectValues(vdst, vsrc reflect.Value) error {
	if vsrc.Kind() == reflect.Interface {
		vsrc = vsrc.Elem()
	}
	if vdst.Kind() == reflect.Interface {
		vdst = vdst.Elem()
	}

	f := getAssignFunc(vdst.Type(), vsrc.Type())
	if f == nil {
		return fmt.Errorf("%w: %s to %s", ErrAssignImpossible, vsrc.Type(), vdst.Type())
	}
	return f(vdst, vsrc)
}

func ptrCount(t reflect.Type) int {
	n := 0
	for t.Kind() == reflect.Pointer {
		n += 1
		t = t.Elem()
	}
	return n
}

func newAssignFunc(dstt, srct reflect.Type) assignFunc {
	//log.Printf("assign func lookup %s → %s", srct, dstt)
	if srct.AssignableTo(dstt) {
		return simpleSet
	}
	if srct.ConvertibleTo(dstt) {
		return convertSet
	}

	srcptrct := ptrCount(srct)
	dstptrct := ptrCount(dstt)
	//log.Printf("assign func lookup %s → %s (%d → %d)", srct, dstt, srcptrct, dstptrct)

	// with this we try to adjust src & dst to have the same number of pointer elements so we may have a chance to assign values directly
	if srcptrct > dstptrct {
		return ptrReadAndAssign(dstt, srct)
	} else if dstptrct > 0 {
		return newNewAndAssign(dstt, srct)
	}

	// check for interfaces/etc
	if reflect.PointerTo(dstt).Implements(valueScannerType) {
		return assignViaScanInterface
	}

	switch dstt.Kind() {
	case reflect.String:
		return makeAssignToString(dstt, srct)
	case reflect.Bool:
		return makeAssignToBool(dstt, srct)
	case reflect.Float32, reflect.Float64:
		return makeAssignToFloat(dstt, srct)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return makeAssignToInt(dstt, srct)
	case reflect.Slice:
		return makeAssignToSlice(dstt, srct)
	case reflect.Map:
		return makeAssignToMap(dstt, srct)
	case reflect.Struct:
		switch srct.Kind() {
		case reflect.Struct:
			return makeAssignStructToStruct(dstt, srct)
		case reflect.Map:
			return makeAssignMapToStruct(dstt, srct)
		}
	}

	//log.Printf("[assign] failed to generate function to convert from %s to %s", srct, dstt)
	return nil
}

func simpleSet(dst, src reflect.Value) error {
	dst.Set(src)
	return nil
}

func convertSet(dst, src reflect.Value) error {
	v := src.Convert(dst.Type())
	dst.Set(v)
	return nil
}

type assignStructInOut struct {
	in, out int
	set     assignFunc
}

type fieldInfo struct {
	reflect.StructField
	idx int
}

func makeAssignStructToStruct(dstt, srct reflect.Type) assignFunc {
	var fields []*assignStructInOut

	fieldsIn := make(map[string]*fieldInfo)
	for i := 0; i < srct.NumField(); i++ {
		f := srct.Field(i)
		name := f.Name
		if jsonTag := f.Tag.Get("json"); jsonTag != "" {
			// check if json tag renames field
			if jsonTag[0] == '-' {
				continue
			}
			if jsonTag[0] != ',' {
				jsonA := strings.Split(jsonTag, ",")
				name = jsonA[0]
			}
		}
		fieldsIn[name] = &fieldInfo{f, i}
	}
	for i := 0; i < dstt.NumField(); i++ {
		dstf := dstt.Field(i)
		name := dstf.Name
		if jsonTag := dstf.Tag.Get("json"); jsonTag != "" {
			// check if json tag renames field
			if jsonTag[0] == '-' {
				continue
			}
			if jsonTag[0] != ',' {
				jsonA := strings.Split(jsonTag, ",")
				name = jsonA[0]
			}
		}
		srcf, ok := fieldsIn[name]
		if !ok {
			continue
		}

		fnc := newAssignFunc(dstf.Type, srcf.StructField.Type)
		if fnc == nil {
			return nil
		}

		fields = append(fields, &assignStructInOut{
			in:  srcf.idx,
			out: i,
			set: fnc,
		})
	}

	fieldsIn = nil

	validator := getValidatorForType(dstt)

	return func(dst, src reflect.Value) error {
		for _, f := range fields {
			dstf := dst.Field(f.out)
			if err := f.set(dstf, src.Field(f.in)); err != nil {
				return err
			}
		}
		if err := validator.validate(dst); err != nil {
			return err
		}
		return nil
	}
}

func makeAssignMapToStruct(dstt, srct reflect.Type) assignFunc {
	// srct is a map
	switch srct.Key().Kind() {
	case reflect.String:
		// we index dstt's fields by string
		fields := make(map[string]*assignStructInOut)
		mapvtype := srct.Elem()

		for i := 0; i < dstt.NumField(); i++ {
			f := dstt.Field(i)
			fnc := newAssignFunc(f.Type, mapvtype)
			if fnc == nil {
				return nil
			}
			name := f.Name
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				// check if json tag renames field
				if jsonTag[0] == '-' {
					continue
				}
				if jsonTag[0] != ',' {
					jsonA := strings.Split(jsonTag, ",")
					name = jsonA[0]
				}
			}
			fields[name] = &assignStructInOut{out: i, set: fnc}
		}

		validator := getValidatorForType(dstt)

		return func(dst, src reflect.Value) error {
			iter := src.MapRange()
			for iter.Next() {
				f, ok := fields[iter.Key().String()]
				if !ok {
					continue
				}
				dstf := dst.Field(f.out)
				if err := f.set(dstf, iter.Value()); err != nil {
					return err
				}
			}
			if err := validator.validate(dst); err != nil {
				return err
			}
			return nil
		}
	default:
		// unsupported map type
		return nil
	}
}

func newNewAndAssign(dstt, srct reflect.Type) assignFunc {
	subt := dstt.Elem()
	subf := newAssignFunc(subt, srct)
	if subf == nil {
		return nil
	}

	return func(dst, src reflect.Value) error {
		if dst.IsNil() {
			dst.Set(reflect.New(subt))
		}
		return subf(dst.Elem(), src)
	}
}

func ptrReadAndAssign(dstt, srct reflect.Type) assignFunc {
	subt := srct.Elem()
	subf := newAssignFunc(dstt, subt)
	if subf == nil {
		return nil
	}

	return func(dst, src reflect.Value) error {
		if src.IsNil() {
			return ErrNilPointerRead
		}
		return subf(dst, src.Elem())
	}
}

func makeAssignToString(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.String:
		return func(dst, src reflect.Value) error {
			dst.Set(src)
			return nil
		}
	case reflect.Slice:
		if srct.Elem().Kind() == reflect.Uint8 {
			// encode to base64
			return func(dst, src reflect.Value) error {
				dst.SetString(base64.StdEncoding.EncodeToString(src.Bytes()))
				return nil
			}
		}
		fallthrough
	default:
		// perform runtime conversion
		return func(dst, src reflect.Value) error {
			str, ok := AsString(src.Interface())
			if !ok {
				return fmt.Errorf("failed to convert %s to string", src.Type())
			}
			dst.SetString(str)
			return nil
		}
	}
}

func makeAssignToSlice(dstt, srct reflect.Type) assignFunc {
	if dstt.Elem().Kind() == reflect.Uint8 {
		// []byte = possibly a string
		return makeAssignToByteSlice(dstt, srct)
	}

	switch srct.Kind() {
	case reflect.Slice:
		// slice→slice
		f := getAssignFunc(dstt.Elem(), srct.Elem())
		if f == nil {
			return nil
		}

		return func(dst, src reflect.Value) error {
			ln := src.Len()
			if dst.Cap() < ln {
				dst.Grow(ln - dst.Cap())
			}
			dst.SetLen(ln)
			//dst.Set(reflect.MakeSlice(dstt.Elem(), ln, ln))
			for i := 0; i < ln; i++ {
				if err := f(dst.Index(i), src.Index(i)); err != nil {
					return err
				}
			}
			return nil
		}
	default:
		return nil // ???
	}
}

func makeAssignToByteSlice(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.String:
		// assume base64 encoded
		return func(dst, src reflect.Value) error {
			dec, err := base64.StdEncoding.DecodeString(src.String())
			if err != nil {
				return err
			}
			dst.SetBytes(dec)
			return nil
		}
	default:
		// ???
		return nil
	}
}

func makeAssignToFloat(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.Float32, reflect.Float64:
		return func(dst, src reflect.Value) error {
			dst.Set(src)
			return nil
		}
	default:
		// perform runtime conversion
		return func(dst, src reflect.Value) error {
			v, ok := AsFloat(src.Interface())
			if !ok {
				return fmt.Errorf("failed to convert %s to float", src.Type())
			}
			dst.SetFloat(v)
			return nil
		}
	}
}

func makeAssignToInt(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(dst, src reflect.Value) error {
			dst.Set(src)
			return nil
		}
	default:
		// perform runtime conversion
		return func(dst, src reflect.Value) error {
			v, ok := AsInt(src.Interface())
			if !ok {
				return fmt.Errorf("failed to convert %s to int", src.Type())
			}
			dst.SetInt(v)
			return nil
		}
	}
}

func makeAssignToBool(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.Bool:
		return func(dst, src reflect.Value) error {
			dst.Set(src)
			return nil
		}
	default:
		// perform runtime conversion
		return func(dst, src reflect.Value) error {
			dst.SetBool(AsBool(src.Interface()))
			return nil
		}
	}
}

func makeAssignToMap(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.Map:
		kf := getAssignFunc(dstt.Key(), srct.Key())
		vf := getAssignFunc(dstt.Elem(), srct.Elem())

		if kf == nil || vf == nil {
			return nil
		}

		return func(dst, src reflect.Value) error {
			dst.Set(reflect.MakeMap(dstt))
			iter := src.MapRange()
			for iter.Next() {
				dk := reflect.New(dstt.Key()).Elem()
				dv := reflect.New(dstt.Elem()).Elem()
				if err := kf(dk, iter.Key()); err != nil {
					return err
				}
				if err := vf(dv, iter.Value()); err != nil {
					return err
				}
				dst.SetMapIndex(dk, dv)
			}
			return nil
		}
	case reflect.Struct:
		if dstt.Key().Kind() != reflect.String {
			// we require map converted from struct to have a string key
			return nil
		}
		subt := dstt.Elem()

		fieldsIn := make(map[string]*assignStructInOut)
		for i := 0; i < srct.NumField(); i++ {
			f := srct.Field(i)
			name := f.Name
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				// check if json tag renames field
				if jsonTag[0] == '-' {
					continue
				}
				if jsonTag[0] != ',' {
					jsonA := strings.Split(jsonTag, ",")
					name = jsonA[0]
				}
			}
			fnc := getAssignFunc(subt, f.Type)
			if fnc == nil {
				return nil
			}
			fieldsIn[name] = &assignStructInOut{in: i, set: fnc}
		}

		return func(dst, src reflect.Value) error {
			dst.Set(reflect.MakeMap(dstt))
			for s, f := range fieldsIn {
				dv := reflect.New(dstt.Elem()).Elem()
				if err := f.set(dv, src.Field(f.in)); err != nil {
					return err
				}
				dst.SetMapIndex(reflect.ValueOf(s), dv)
			}
			return nil
		}
	default:
		return nil
	}
}

func assignViaScanInterface(dst, src reflect.Value) error {
	if !dst.CanAddr() {
		return ErrDestinationNotAddressable
	}
	return dst.Addr().Interface().(valueScanner).Scan(src.Interface())
}
