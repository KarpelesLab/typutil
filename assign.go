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

func newAssignFunc(dstt reflect.Type, srct reflect.Type) assignFunc {
	if srct.AssignableTo(dstt) {
		return simpleSet
	}
	if srct.ConvertibleTo(dstt) {
		return convertSet
	}
	switch dstt.Kind() {
	case reflect.Pointer:
		return newNewAndAssign(dstt, srct)
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
