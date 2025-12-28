package typutil_test

import (
	"errors"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestAssignSimple(t *testing.T) {
	var a string
	err := typutil.Assign(&a, "hello world")

	if err != nil {
		t.Errorf("simple assign failed: %s", err)
	} else if a != "hello world" {
		t.Errorf("unexpected value %v", a)
	}

	var b *string
	err = typutil.Assign(&b, "hello too")

	if err != nil {
		t.Errorf("simple assign failed: %s", err)
	} else if *b != "hello too" {
		t.Errorf("unexpected value %v", a)
	}
}

type objA struct {
	A string
	B int `json:"X"`
	C float64
	D float64
}

type objB struct {
	A string
	B float64 `json:"X"`
	C int32
	D uint16
}

func TestAssignAs(t *testing.T) {
	a, err := typutil.As[int]("42")

	if err != nil {
		t.Errorf("got error = %s", err)
	}

	if a != 42 {
		t.Errorf("expected 42, got %v", a)
	}

	b, err := typutil.As[int](nil)
	if !errors.Is(err, typutil.ErrInvalidSource) {
		t.Errorf("got unexpected error = %s", err)
	}

	if b != 0 {
		t.Errorf("expected 0, got %v", b)
	}

	c, err := typutil.As[string](42)
	if err != nil {
		t.Errorf("got error = %s", err)
	}

	if c != "42" {
		t.Errorf("expected 42, got %v", c)
	}
}

func TestAssignAssignTo(t *testing.T) {
	a, err := typutil.As[objA](typutil.RawJsonMessage(`{"A":"value A","X":42,"C":123.456,"D":1e-19}`))
	if err != nil {
		t.Errorf("got error = %s", err)
		return
	}

	if a.A != "value A" {
		t.Errorf("unexpected value for a.A: %s", a.A)
	}
	if a.B != 42 {
		t.Errorf("unexpected value for a.B: %v", a.B)
	}
}

func TestAssignStruct(t *testing.T) {
	a := &objA{A: "hello world", B: 42, C: 123456.789321, D: 555.22}
	var b *objB

	err := typutil.Assign(&b, a)
	if err != nil {
		t.Errorf("struct to struct assign failed: %s", err)
		return
	}

	if b.A != "hello world" {
		t.Errorf("struct to struct string missing")
	}
	if b.B != 42 {
		t.Errorf("struct to struct B missing")
	}
	if b.C != 123457 {
		t.Errorf("struct to struct C invalid value: %v", b.C)
	}
	if b.D != 555 {
		t.Errorf("struct to struct D invalid value: %v", b.D)
	}

	// test assign from map
	b = nil
	arg := any(map[string]any{"A": "from a map", "X": 99, "C": 123.456})
	err = typutil.Assign(&b, &arg)
	if err != nil {
		t.Errorf("map to struct assign failed: %s", err)
		return
	}

	if b.A != "from a map" {
		t.Errorf("map to struct string invalid")
	}
	if b.B != 99 {
		t.Errorf("map to struct B invalid")
	}
	if b.C != 123 {
		t.Errorf("map to struct C invalid")
	}

	var m map[string]any

	err = typutil.Assign(&m, b)
	if err != nil {
		t.Errorf("struct to map assign failed: %s", err)
		return
	}

	if m["A"].(string) != "from a map" {
		t.Errorf("struct to map A failed")
	}
}

func TestAsMapToStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Age     int
		Active  bool
		Score   float64
		Tag     string `json:"customTag"`
		Missing string // not in map, should be zero value
	}

	m := map[string]any{
		"Name":      "Alice",
		"Age":       "30", // string that should convert to int
		"Active":    true,
		"Score":     95.5,
		"customTag": "tagged",  // should match via json tag
		"Extra":     "ignored", // should be ignored
	}

	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("As[TestStruct](map[string]any) failed: %s", err)
		return
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name='Alice', got %q", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected Age=30, got %d", result.Age)
	}
	if !result.Active {
		t.Errorf("expected Active=true, got %v", result.Active)
	}
	if result.Score != 95.5 {
		t.Errorf("expected Score=95.5, got %f", result.Score)
	}
	if result.Tag != "tagged" {
		t.Errorf("expected Tag='tagged' (via json tag), got %q", result.Tag)
	}
	if result.Missing != "" {
		t.Errorf("expected Missing='' (zero value), got %q", result.Missing)
	}
}

func TestAsMapToStructPointer(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	m := map[string]any{"Value": "test"}

	// Test that As returns the struct value, not a pointer
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("As[TestStruct](map) failed: %s", err)
		return
	}

	if result.Value != "test" {
		t.Errorf("expected Value='test', got %q", result.Value)
	}

	// Test with pointer type
	ptrResult, err := typutil.As[*TestStruct](m)
	if err != nil {
		t.Errorf("As[*TestStruct](map) failed: %s", err)
		return
	}

	if ptrResult == nil {
		t.Errorf("expected non-nil pointer")
		return
	}

	if ptrResult.Value != "test" {
		t.Errorf("expected Value='test', got %q", ptrResult.Value)
	}
}

func TestAssignErrors(t *testing.T) {
	t.Run("non-pointer destination", func(t *testing.T) {
		var a string
		err := typutil.Assign(a, "hello")
		if !errors.Is(err, typutil.ErrAssignDestNotPointer) {
			t.Errorf("expected ErrAssignDestNotPointer, got %v", err)
		}
	})

	t.Run("nil pointer destination", func(t *testing.T) {
		var a *string
		err := typutil.Assign(a, "hello")
		if !errors.Is(err, typutil.ErrAssignDestNotPointer) {
			t.Errorf("expected ErrAssignDestNotPointer, got %v", err)
		}
	})
}

func TestAssignToString(t *testing.T) {
	t.Run("byte slice to string (base64)", func(t *testing.T) {
		var s string
		err := typutil.Assign(&s, []byte("hello"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Should be base64 encoded
		if s != "aGVsbG8=" {
			t.Errorf("expected base64 'aGVsbG8=', got %q", s)
		}
	})

	t.Run("number to string", func(t *testing.T) {
		var s string
		err := typutil.Assign(&s, 42)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s != "42" {
			t.Errorf("expected '42', got %q", s)
		}
	})

	t.Run("bool to string", func(t *testing.T) {
		var s string
		err := typutil.Assign(&s, true)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if s != "1" {
			t.Errorf("expected '1', got %q", s)
		}
	})
}

func TestAssignToBool(t *testing.T) {
	t.Run("string to bool", func(t *testing.T) {
		var b bool
		err := typutil.Assign(&b, "true")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !b {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("number to bool", func(t *testing.T) {
		var b bool
		err := typutil.Assign(&b, 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !b {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("zero to bool", func(t *testing.T) {
		var b bool
		err := typutil.Assign(&b, 0)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if b {
			t.Errorf("expected false, got true")
		}
	})
}

func TestAssignToFloat(t *testing.T) {
	t.Run("string to float", func(t *testing.T) {
		var f float64
		err := typutil.Assign(&f, "3.14")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if f != 3.14 {
			t.Errorf("expected 3.14, got %f", f)
		}
	})

	t.Run("int to float", func(t *testing.T) {
		var f float64
		err := typutil.Assign(&f, 42)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if f != 42.0 {
			t.Errorf("expected 42.0, got %f", f)
		}
	})

	t.Run("invalid string to float", func(t *testing.T) {
		var f float64
		err := typutil.Assign(&f, "not a number")
		if err == nil {
			t.Errorf("expected error for invalid conversion")
		}
	})
}

func TestAssignToInt(t *testing.T) {
	t.Run("string to int", func(t *testing.T) {
		var i int
		err := typutil.Assign(&i, "42")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if i != 42 {
			t.Errorf("expected 42, got %d", i)
		}
	})

	t.Run("float to int", func(t *testing.T) {
		var i int
		err := typutil.Assign(&i, 3.7)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if i != 4 { // rounded
			t.Errorf("expected 4, got %d", i)
		}
	})

	t.Run("invalid string to int", func(t *testing.T) {
		var i int
		err := typutil.Assign(&i, "not a number")
		if err == nil {
			t.Errorf("expected error for invalid conversion")
		}
	})
}

func TestAssignToUint(t *testing.T) {
	t.Run("string to uint", func(t *testing.T) {
		var u uint
		err := typutil.Assign(&u, "42")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if u != 42 {
			t.Errorf("expected 42, got %d", u)
		}
	})

	t.Run("int to uint", func(t *testing.T) {
		var u uint
		err := typutil.Assign(&u, 42)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if u != 42 {
			t.Errorf("expected 42, got %d", u)
		}
	})

	t.Run("invalid string to uint", func(t *testing.T) {
		var u uint
		err := typutil.Assign(&u, "not a number")
		if err == nil {
			t.Errorf("expected error for invalid conversion")
		}
	})
}

func TestAssignToByteSlice(t *testing.T) {
	t.Run("valid base64 to byte slice", func(t *testing.T) {
		var b []byte
		err := typutil.Assign(&b, "aGVsbG8=") // "hello" in base64
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if string(b) != "hello" {
			t.Errorf("expected 'hello', got %q", string(b))
		}
	})

	t.Run("invalid base64 to byte slice", func(t *testing.T) {
		var b []byte
		err := typutil.Assign(&b, "not valid base64!!!")
		if err == nil {
			t.Errorf("expected error for invalid base64")
		}
	})
}

func TestAssignMapConversions(t *testing.T) {
	t.Run("map key conversion", func(t *testing.T) {
		src := map[string]int{"one": 1, "two": 2}
		var dst map[string]float64
		err := typutil.Assign(&dst, src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dst["one"] != 1.0 || dst["two"] != 2.0 {
			t.Errorf("unexpected values: %v", dst)
		}
	})

	t.Run("map value conversion", func(t *testing.T) {
		src := map[string]string{"num": "42"}
		var dst map[string]int
		err := typutil.Assign(&dst, src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dst["num"] != 42 {
			t.Errorf("expected 42, got %d", dst["num"])
		}
	})
}

func TestAssignSliceConversions(t *testing.T) {
	t.Run("int slice to float slice", func(t *testing.T) {
		src := []int{1, 2, 3}
		var dst []float64
		err := typutil.Assign(&dst, src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(dst) != 3 || dst[0] != 1.0 || dst[1] != 2.0 || dst[2] != 3.0 {
			t.Errorf("unexpected values: %v", dst)
		}
	})

	t.Run("string slice to int slice", func(t *testing.T) {
		src := []string{"1", "2", "3"}
		var dst []int
		err := typutil.Assign(&dst, src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(dst) != 3 || dst[0] != 1 || dst[1] != 2 || dst[2] != 3 {
			t.Errorf("unexpected values: %v", dst)
		}
	})
}

func TestAssignStructUnexportedFields(t *testing.T) {
	type srcStruct struct {
		Public   string
		internal string
	}
	type dstStruct struct {
		Public   string
		internal string
	}

	src := srcStruct{Public: "visible", internal: "hidden"}
	var dst dstStruct
	err := typutil.Assign(&dst, src)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if dst.Public != "visible" {
		t.Errorf("expected 'visible', got %q", dst.Public)
	}
	// internal field should not be copied
	if dst.internal != "" {
		t.Errorf("unexported field should not be copied")
	}
}

func TestAssignStructWithJsonMinusTag(t *testing.T) {
	type srcStruct struct {
		Name    string
		Ignored string `json:"-"`
	}
	type dstStruct struct {
		Name    string
		Ignored string `json:"-"`
	}

	src := srcStruct{Name: "test", Ignored: "should be ignored"}
	var dst dstStruct
	err := typutil.Assign(&dst, src)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if dst.Name != "test" {
		t.Errorf("expected 'test', got %q", dst.Name)
	}
	if dst.Ignored != "" {
		t.Errorf("ignored field should not be copied, got %q", dst.Ignored)
	}
}

func TestAssignInterfaceValue(t *testing.T) {
	t.Run("interface to struct via runtime", func(t *testing.T) {
		type TestStruct struct {
			Value int
		}
		var src interface{} = map[string]any{"Value": 42}
		var dst TestStruct
		err := typutil.Assign(&dst, src)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if dst.Value != 42 {
			t.Errorf("expected 42, got %d", dst.Value)
		}
	})
}
