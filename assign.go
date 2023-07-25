package typutil

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type assignFunc func(dst, src reflect.Value) error

type assignConvType struct {
	dst reflect.Type
	src reflect.Type
}

var assignFuncCache sync.Map // map[assignConvType]assignFunc

var (
	ErrAssignDestNotPointer = errors.New("assign destination must be a pointer")
	ErrAssignImpossible     = errors.New("the requested assign is not possible")
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

func newAssignFunc(dstt, srct reflect.Type) assignFunc {
	if srct.AssignableTo(dstt) {
		return simpleSet
	}
	if srct.ConvertibleTo(dstt) {
		return convertSet
	}
	if srct.Kind() == reflect.Pointer {
		return ptrReadAndAssign(dstt, srct)
	}

	switch dstt.Kind() {
	case reflect.Pointer:
		return newNewAndAssign(dstt, srct)
	case reflect.Struct:
		switch srct.Kind() {
		case reflect.Struct:
			return makeAssignStructToStruct(dstt, srct)
		}
	}
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

	// TODO process tag

	fieldsIn := make(map[string]*fieldInfo)
	for i := 0; i < srct.NumField(); i++ {
		f := srct.Field(i)
		fieldsIn[f.Name] = &fieldInfo{f, i}
	}
	for i := 0; i < dstt.NumField(); i++ {
		dstf := dstt.Field(i)
		srcf, ok := fieldsIn[dstf.Name]
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

	return func(dst, src reflect.Value) error {
		for _, f := range fields {
			if err := f.set(dst.Field(f.out), src.Field(f.in)); err != nil {
				return err
			}
		}
		return nil
	}
}

func newNewAndAssign(dstt reflect.Type, srct reflect.Type) assignFunc {
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
