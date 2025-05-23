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

// AssignableTo is an interface that can be implemented by objects able to assign themselves
// to values. Do not use Assign() inside AssignTo or you risk generating an infinite loop.
//
// This looks a bit like sql's Valuer, except instead of a returning a any interface, the
// parameter is a pointer to the target type. Typically, Unmarshal will be used in here.
type AssignableTo interface {
	AssignTo(any) error
}

var (
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	valueScannerType    = reflect.TypeOf((*valueScanner)(nil)).Elem()
	valueAssignerType   = reflect.TypeOf((*AssignableTo)(nil)).Elem()
)

func getAssignFunc(dstt reflect.Type, srct reflect.Type) (assignFunc, error) {
	if dstt == srct {
		return simpleSet, nil
	}

	act := assignConvType{dstt, srct}
	if fi, ok := assignFuncCache.Load(act); ok {
		return fi.(assignFunc), nil
	}

	// deal with recursive type the same way encoding/json does
	var (
		wg  sync.WaitGroup
		f   assignFunc
		err error
	)
	wg.Add(1)
	defer wg.Done()

	fi, loaded := assignFuncCache.LoadOrStore(act, assignFunc(func(dst, src reflect.Value) error {
		wg.Wait()
		if err != nil {
			return err
		}
		return f(dst, src)
	}))
	if loaded {
		return fi.(assignFunc), nil
	}

	// compute real func
	f, err = newAssignFunc(dstt, srct)
	if err != nil {
		assignFuncCache.Delete(act)
		return nil, err
	}
	assignFuncCache.Store(act, f)
	return f, nil
}

// Assign sets dst to the value of src, performing type conversion as needed.
//
// This is the main entry point for the type conversion system in typutil. It handles
// automatic conversion between various Go types, including:
// - Primitive types (string, int, float, bool)
// - Pointers and interfaces
// - Slices and maps
// - Structs (using field names or JSON tags for matching)
// - Custom types that implement valueScanner or AssignableTo interfaces
//
// For container types (slices, maps, structs), a shallow copy is performed.
//
// Parameters:
//   - dst: A pointer to the destination value (must be a non-nil pointer)
//   - src: The source value to assign from
//
// Returns:
//   - An error if the assignment cannot be performed, such as when:
//   - The destination is not a pointer
//   - The types are incompatible and cannot be converted
//   - The source value is invalid
//
// Example:
//
//	// Convert between compatible types
//	var i int
//	err := Assign(&i, "42")  // i becomes 42
//
//	// Convert between struct types with matching fields
//	type Person struct {
//	    Name string
//	    Age int
//	}
//	type User struct {
//	    Name string
//	    Age string
//	}
//	p := Person{Name: "Alice", Age: 30}
//	var u User
//	err := Assign(&u, p)  // u becomes User{Name: "Alice", Age: "30"}
//
// Note that unlike json.Unmarshal or similar functions, Assign requires a pointer
// to the destination value, not the destination value itself.
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
	f, err := getAssignFunc(vdst.Type(), vsrc.Type())
	if err != nil {
		return fmt.Errorf("%w (assigning %T to %T)", err, src, dst)
	}
	return f(vdst, vsrc)
}

// AssignReflect assigns a value from one reflect.Value to another, with type conversion.
//
// This is the reflection-based version of Assign that works directly with reflect.Value
// objects. It's used internally by the library and is also available for advanced use cases
// where you're already working with reflection.
//
// Parameters:
//   - vdst: The destination reflect.Value
//   - vsrc: The source reflect.Value
//
// Returns:
//   - An error if the assignment cannot be performed
//
// This function handles unwrapping interface values, dealing with pointers,
// and finding the appropriate conversion function for the types involved.
func AssignReflect(vdst, vsrc reflect.Value) error {
	if vsrc.Kind() == reflect.Interface {
		vsrc = vsrc.Elem()
	}
	if vdst.Kind() == reflect.Interface {
		vdst = vdst.Elem()
	}
	if !vdst.CanAddr() && vdst.Kind() == reflect.Ptr {
		vdst = vdst.Elem()
	}

	if !vsrc.IsValid() {
		return ErrInvalidSource
	}

	f, err := getAssignFunc(vdst.Type(), vsrc.Type())
	if err != nil {
		return fmt.Errorf("%w (assigning %s to %s)", err, vsrc.Type(), vdst.Type())
	}
	return f(vdst, vsrc)
}

// As converts a value to the specified type T, with type conversion as needed.
//
// This generic function provides a type-safe way to convert values between
// different types. It leverages the type conversion capabilities of Assign
// but returns the result as the requested type rather than modifying an
// existing variable.
//
// Type Parameters:
//   - T: The target type to convert to
//
// Parameters:
//   - v: The value to convert
//
// Returns:
//   - The converted value as type T
//   - An error if the conversion cannot be performed
//
// Example:
//
//	// Convert string to int
//	i, err := As[int]("42")  // i is 42
//
//	// Convert between struct types
//	type Person struct {Name string; Age int}
//	type User struct {Name string; Age string}
//
//	p := Person{Name: "Alice", Age: 30}
//	u, err := As[User](p)  // u is User{Name: "Alice", Age: "30"}
func As[T any](v any) (T, error) {
	// convert any type to T
	typ := reflect.TypeOf((*T)(nil)).Elem()
	obj := reflect.New(typ) // it's a pointer

	err := AssignReflect(obj, reflect.ValueOf(v))

	return obj.Elem().Interface().(T), err
}

func ptrCount(t reflect.Type) int {
	n := 0
	for t.Kind() == reflect.Pointer {
		n += 1
		t = t.Elem()
	}
	return n
}

func newAssignFunc(dstt, srct reflect.Type) (assignFunc, error) {
	//log.Printf("assign func lookup %s → %s", srct, dstt)
	if srct.AssignableTo(dstt) {
		return simpleSet, nil
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
		return makeAssignScanIntf(dstt, srct)
	}
	if reflect.PointerTo(srct).Implements(valueAssignerType) {
		return makeAssignToIntf(dstt, srct)
	}

	switch dstt.Kind() {
	case reflect.String:
		return makeAssignToString(dstt, srct), nil
	case reflect.Bool:
		return makeAssignToBool(dstt, srct), nil
	case reflect.Float32, reflect.Float64:
		return makeAssignToFloat(dstt, srct), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return makeAssignToInt(dstt, srct), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return makeAssignToUint(dstt, srct), nil
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
		case reflect.Interface:
			return makeAssignAnyToRuntime(dstt, srct), nil
		}
	}

	//log.Printf("[assign] failed to generate function to convert from %s to %s", srct, dstt)
	return nil, fmt.Errorf("%w: invalid conversion from %s to %s", ErrAssignImpossible, srct, dstt)
}

func simpleSet(dst, src reflect.Value) error {
	dst.Set(src)
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

func makeAssignStructToStruct(dstt, srct reflect.Type) (assignFunc, error) {
	var fields []*assignStructInOut

	fieldsIn := make(map[string]*fieldInfo)
	for i, m := 0, srct.NumField(); i < m; i++ {
		f := srct.Field(i)
		if !f.IsExported() {
			// skip non-exported fields
			continue
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
		fieldsIn[name] = &fieldInfo{f, i}
	}
	for i, m := 0, dstt.NumField(); i < m; i++ {
		dstf := dstt.Field(i)
		if !dstf.IsExported() {
			// skip non-exported fields
			continue
		}
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

		fnc, err := newAssignFunc(dstf.Type, srcf.StructField.Type)
		if fnc == nil {
			return nil, err
		}

		fields = append(fields, &assignStructInOut{
			in:  srcf.idx,
			out: i,
			set: fnc,
		})
	}

	fieldsIn = nil

	validator := getValidatorForType(dstt)

	f := func(dst, src reflect.Value) error {
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
	return f, nil
}

func makeAssignMapToStruct(dstt, srct reflect.Type) (assignFunc, error) {
	// srct is a map
	switch srct.Key().Kind() {
	case reflect.String:
		// we index dstt's fields by string
		fields := make(map[string]*assignStructInOut)
		mapvtype := srct.Elem()

		for i := 0; i < dstt.NumField(); i++ {
			f := dstt.Field(i)
			if !f.IsExported() {
				// skip non-exported fields
				continue
			}
			fnc, err := newAssignFunc(f.Type, mapvtype)
			if err != nil {
				return nil, err
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

		f := func(dst, src reflect.Value) error {
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
		return f, nil
	default:
		// unsupported map type
		return nil, fmt.Errorf("%w: invalid src type %s", ErrAssignImpossible, srct)
	}
}

func makeAssignAnyToRuntime(dstt, srct reflect.Type) assignFunc {
	return func(dst, src reflect.Value) error {
		return AssignReflect(dst, src)
	}
}

func newNewAndAssign(dstt, srct reflect.Type) (assignFunc, error) {
	subt := dstt.Elem()
	subf, err := newAssignFunc(subt, srct)
	if err != nil {
		return nil, err
	}

	f := func(dst, src reflect.Value) error {
		if dst.IsNil() {
			dst.Set(reflect.New(subt))
		}
		return subf(dst.Elem(), src)
	}
	return f, nil
}

func ptrReadAndAssign(dstt, srct reflect.Type) (assignFunc, error) {
	subt := srct.Elem()
	subf, err := newAssignFunc(dstt, subt)
	if err != nil {
		return nil, err
	}

	f := func(dst, src reflect.Value) error {
		if src.IsNil() {
			return ErrNilPointerRead
		}
		return subf(dst, src.Elem())
	}
	return f, nil
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

func makeAssignToSlice(dstt, srct reflect.Type) (assignFunc, error) {
	if dstt.Elem().Kind() == reflect.Uint8 {
		// []byte = possibly a string
		return makeAssignToByteSlice(dstt, srct)
	}

	switch srct.Kind() {
	case reflect.Slice:
		// slice→slice
		convfunc, err := getAssignFunc(dstt.Elem(), srct.Elem())
		if err != nil {
			return nil, err
		}

		f := func(dst, src reflect.Value) error {
			ln := src.Len()
			if dst.Cap() < ln {
				dst.Grow(ln - dst.Cap())
			}
			dst.SetLen(ln)
			//dst.Set(reflect.MakeSlice(dstt.Elem(), ln, ln))
			for i := 0; i < ln; i++ {
				if err := convfunc(dst.Index(i), src.Index(i)); err != nil {
					return err
				}
			}
			return nil
		}
		return f, nil
	case reflect.Interface:
		// perform this runtime
		return makeAssignAnyToRuntime(dstt, srct), nil
	default:
		return nil, fmt.Errorf("%w: invalid source %s", ErrAssignImpossible, srct.Kind())
	}
}

func makeAssignToByteSlice(dstt, srct reflect.Type) (assignFunc, error) {
	switch srct.Kind() {
	case reflect.String:
		// assume base64 encoded
		f := func(dst, src reflect.Value) error {
			dec, err := base64.StdEncoding.DecodeString(src.String())
			if err != nil {
				return err
			}
			dst.SetBytes(dec)
			return nil
		}
		return f, nil
	case reflect.Interface:
		// perform this runtime
		return makeAssignAnyToRuntime(dstt, srct), nil
	default:
		return nil, fmt.Errorf("%w: unsupported type %s to byte slice", ErrAssignImpossible, srct)
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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
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

func makeAssignToUint(dstt, srct reflect.Type) assignFunc {
	switch srct.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(dst, src reflect.Value) error {
			dst.Set(src)
			return nil
		}
	default:
		// perform runtime conversion
		return func(dst, src reflect.Value) error {
			v, ok := AsUint(src.Interface())
			if !ok {
				return fmt.Errorf("failed to convert %s to int", src.Type())
			}
			dst.SetUint(v)
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

func makeAssignToMap(dstt, srct reflect.Type) (assignFunc, error) {
	switch srct.Kind() {
	case reflect.Map:
		kf, err := getAssignFunc(dstt.Key(), srct.Key())
		if err != nil {
			return nil, err
		}
		vf, err := getAssignFunc(dstt.Elem(), srct.Elem())
		if err != nil {
			return nil, err
		}

		f := func(dst, src reflect.Value) error {
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
		return f, nil
	case reflect.Struct:
		if dstt.Key().Kind() != reflect.String {
			// we require map converted from struct to have a string key
			return nil, fmt.Errorf("%w: map key is not of string type", ErrAssignImpossible)
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
			fnc, err := getAssignFunc(subt, f.Type)
			if err != nil {
				return nil, err
			}
			fieldsIn[name] = &assignStructInOut{in: i, set: fnc}
		}

		f := func(dst, src reflect.Value) error {
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
		return f, nil
	default:
		return nil, fmt.Errorf("%w: unsupported type %s", ErrAssignImpossible, srct)
	}
}

func makeAssignScanIntf(dstt, srct reflect.Type) (assignFunc, error) {
	validator := getValidatorForType(dstt)

	f := func(dst, src reflect.Value) error {
		if !dst.CanAddr() {
			return ErrDestinationNotAddressable
		}
		err := dst.Addr().Interface().(valueScanner).Scan(src.Interface())
		if err != nil {
			return err
		}

		return validator.validate(dst)
	}
	return f, nil
}

func makeAssignToIntf(dstt, srct reflect.Type) (assignFunc, error) {
	validator := getValidatorForType(dstt)

	f := func(dst, src reflect.Value) error {
		if !dst.CanAddr() {
			return ErrDestinationNotAddressable
		}
		srcptr := reflect.New(srct)
		srcptr.Elem().Set(src)
		err := srcptr.Interface().(AssignableTo).AssignTo(dst.Addr().Interface())
		if err != nil {
			return err
		}

		return validator.validate(dst)
	}
	return f, nil
}
