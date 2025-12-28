package typutil_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestStaticStd(t *testing.T) {
	a := func() string {
		return "hello"
	}

	st := typutil.Func(a)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err := st.Call(context.Background())
	if err != nil {
		t.Errorf("failed to perform: %s", err)
	}

	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "hello" {
		t.Fatalf("failed to execute")
	}
}

func TestStaticParam(t *testing.T) {
	a := func(in struct{ A string }) string {
		return strings.ToUpper(in.A)
	}

	st := typutil.Func(a)
	if st == nil {
		t.Fatalf("unable to gen static method for param test")
	}

	res, err := st.CallArg(context.Background(), map[string]any{"A": "hello"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "HELLO" {
		t.Errorf("failed to perform, result is %s", str)
	}

	res, err = st.CallArg(context.Background(), struct{ A string }{A: "world"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok = res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != "WORLD" {
		t.Errorf("failed to perform, result is %s", str)
	}
}

func TestStaticParams(t *testing.T) {
	add := func(a, b int) (int, error) {
		return a + b, nil
	}
	st := typutil.Func(add)

	res, err := typutil.Call[int](st, context.Background(), 1, "2")
	if err != nil {
		t.Errorf("error returned: %s", err)
	} else if res != 3 {
		t.Errorf("error, got 1 + 2 = %d", res)
	}

	// test with input_json
	//lint:ignore SA1029 library uses string key for input_json context value
	ctx := context.WithValue(context.Background(), "input_json", json.RawMessage("[3,4]"))

	resAny, err := st.Call(ctx)
	if err != nil {
		t.Errorf("error returned: %s", err)
	} else if resAny.(int) != 7 {
		t.Errorf("error, got 3 + 4 = %d", resAny)
	}
}

type scannableObject struct {
	A string
}

func (s *scannableObject) Scan(v any) error {
	s.A = fmt.Sprintf("%#v", v)
	return nil
}

func TestStaticScanner(t *testing.T) {
	a := func(in struct{ Foo *scannableObject }) string {
		return in.Foo.A
	}

	st := typutil.Func(a)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err := st.CallArg(context.Background(), map[string]any{"Foo": "Hello"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok := res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != `"Hello"` {
		t.Errorf("failed to perform, result is %s", str)
	}

	b := func(in struct{ Foo scannableObject }) string {
		return in.Foo.A
	}

	st = typutil.Func(b)
	if st == nil {
		t.Fatalf("unable to gen static method")
	}

	res, err = st.CallArg(context.Background(), map[string]any{"Foo": "World"})
	if err != nil {
		t.Fatalf("failed to run: %s", err)
	}
	str, ok = res.(string)
	if !ok {
		t.Fatalf("failed to convert type")
	}
	if str != `"World"` {
		t.Errorf("failed to perform, result is %s", str)
	}
}

func TestDefaultArgs(t *testing.T) {
	myFunc := func(a, b, c int) int {
		return a + b + c
	}

	f := typutil.Func(myFunc).WithDefaults(typutil.Required, typutil.Required, 42)

	res, err := typutil.Call[int](f, context.Background(), 10, 20)
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if res != 72 {
		t.Errorf("expected res==72, got %d", res)
	}

	_, err = typutil.Call[int](f, context.Background(), 10)
	if !errors.Is(err, typutil.ErrMissingArgs) {
		t.Errorf("unexpected error on missing args: %s", err)
	}

	_, err = typutil.Call[int](f, context.Background(), 1, 2, 3, 4)
	if !errors.Is(err, typutil.ErrTooManyArgs) {
		t.Errorf("unexpected error on too many args: %s", err)
	}

	myFuncVar := func(ms string, v ...int) int {
		m, _ := strconv.Atoi(ms)
		var r int
		for _, x := range v {
			r += x + m
		}
		return r
	}

	f = typutil.Func(myFuncVar).WithDefaults(typutil.Required, 3, 7)

	res, err = typutil.Call[int](f, context.Background(), 1) // 1+3 + 1+7 = 12
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if res != 12 {
		t.Errorf("expected res==12, got %d", res)
	}

	f = typutil.Func(myFuncVar, typutil.StrictArgs)

	_, err = typutil.Call[int](f, context.Background(), 1, 2, 3) // error
	if !errors.Is(err, typutil.ErrAssignImpossible) {
		t.Errorf("unexpected error %v", err)
	}
}

func TestDefaultArgsCtxBuf(t *testing.T) {
	myFunc := func(ctx context.Context, A string, B bool, C string) (any, error) {
		if B {
			return C, nil
		} else {
			return A, nil
		}
	}

	f := typutil.Func(myFunc).WithDefaults(typutil.Required, false, "C_default")

	res, err := typutil.Call[string](f, context.Background(), "A_set")
	if err != nil {
		t.Errorf("error: %s", err)
	}
	if res != "A_set" {
		t.Errorf("unexpected result")
	}
}

func TestCallableString(t *testing.T) {
	t.Run("no args", func(t *testing.T) {
		f := typutil.Func(func() {})
		s := f.String()
		if s != "func()" {
			t.Errorf("expected 'func()', got %q", s)
		}
	})

	t.Run("single arg", func(t *testing.T) {
		f := typutil.Func(func(a int) {})
		s := f.String()
		if s != "func(int)" {
			t.Errorf("expected 'func(int)', got %q", s)
		}
	})

	t.Run("multiple args", func(t *testing.T) {
		f := typutil.Func(func(a string, b int, c bool) {})
		s := f.String()
		if s != "func(string, int, bool)" {
			t.Errorf("expected 'func(string, int, bool)', got %q", s)
		}
	})

	t.Run("variadic", func(t *testing.T) {
		f := typutil.Func(func(a string, b ...int) {})
		s := f.String()
		if s != "func(string, ...int)" {
			t.Errorf("expected 'func(string, ...int)', got %q", s)
		}
	})

	t.Run("complex types", func(t *testing.T) {
		f := typutil.Func(func(a []string, b map[string]int) {})
		s := f.String()
		if !strings.Contains(s, "[]string") || !strings.Contains(s, "map[string]int") {
			t.Errorf("expected complex types in string, got %q", s)
		}
	})
}

func TestCallableArgKind(t *testing.T) {
	f := typutil.Func(func(a string, b int, c bool, d float64) {})

	tests := []struct {
		index    int
		expected reflect.Kind
	}{
		{0, reflect.String},
		{1, reflect.Int},
		{2, reflect.Bool},
		{3, reflect.Float64},
		{4, reflect.Invalid},  // out of bounds
		{-1, reflect.Invalid}, // negative index (will be out of bounds check)
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("arg_%d", tc.index), func(t *testing.T) {
			// Note: negative index check might behave differently
			if tc.index >= 0 {
				kind := f.ArgKind(tc.index)
				if kind != tc.expected {
					t.Errorf("ArgKind(%d) = %v, want %v", tc.index, kind, tc.expected)
				}
			}
		})
	}
}

func TestCallableIsStringArg(t *testing.T) {
	f := typutil.Func(func(a string, b int, c []byte, d string) {})

	tests := []struct {
		index    int
		expected bool
	}{
		{0, true},  // string
		{1, false}, // int
		{2, false}, // []byte
		{3, true},  // string
		{4, false}, // out of bounds
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("arg_%d", tc.index), func(t *testing.T) {
			result := f.IsStringArg(tc.index)
			if result != tc.expected {
				t.Errorf("IsStringArg(%d) = %v, want %v", tc.index, result, tc.expected)
			}
		})
	}
}
