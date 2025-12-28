package typutil_test

import (
	"testing"

	"github.com/KarpelesLab/typutil"
)

type equalTestVector struct {
	a, b any
	res  bool
}

func TestEqual(t *testing.T) {
	v := []*equalTestVector{
		{1, 1, true},
		{1, 0, false},
		{42, "42", true},
		{5, 5.0, true},
		{5, 5.5, false},
		{5.5, 5, false},
		{"42e-5", 0.00042, true},
	}

	for _, sub := range v {
		res := typutil.Equal(sub.a, sub.b)
		if res != sub.res {
			t.Errorf("equal failed: %v == %v should return %v, got %v", sub.a, sub.b, sub.res, res)
		}
	}
}

func TestEqualNil(t *testing.T) {
	t.Run("nil nil", func(t *testing.T) {
		if got := typutil.Equal(nil, nil); !got {
			t.Errorf("Equal(nil, nil) = %v, want true", got)
		}
	})

	t.Run("nil string", func(t *testing.T) {
		if got := typutil.Equal(nil, ""); got {
			t.Errorf("Equal(nil, \"\") = %v, want false", got)
		}
	})

	t.Run("string nil", func(t *testing.T) {
		if got := typutil.Equal("", nil); got {
			t.Errorf("Equal(\"\", nil) = %v, want false", got)
		}
	})
}

func TestEqualBytes(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{"equal bytes", []byte("hello"), []byte("hello"), true},
		{"unequal bytes", []byte("hello"), []byte("world"), false},
		{"empty bytes", []byte{}, []byte{}, true},
		{"bytes vs string", []byte("hello"), "hello", true},
		{"string vs bytes", "hello", []byte("hello"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typutil.Equal(tt.a, tt.b); got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEqualNumericTypes(t *testing.T) {
	tests := []struct {
		name string
		a, b any
		want bool
	}{
		{"int8 equal", int8(42), int8(42), true},
		{"int16 equal", int16(42), int16(42), true},
		{"int32 equal", int32(42), int32(42), true},
		{"int64 equal", int64(42), int64(42), true},
		{"uint8 equal", uint8(42), uint8(42), true},
		{"uint16 equal", uint16(42), uint16(42), true},
		{"uint32 equal", uint32(42), uint32(42), true},
		{"uint64 equal", uint64(42), uint64(42), true},
		{"float32 equal", float32(3.14), float32(3.14), true},
		{"float64 equal", float64(3.14), float64(3.14), true},
		{"int vs float", 42, 42.0, true},
		{"float vs int", 42.0, 42, true},
		{"int vs string", 42, "42", true},
		{"string vs int", "42", 42, true},
		{"bool true vs int", true, 1, true},
		{"bool false vs int", false, 0, true},
		{"int vs bool true", 1, true, true},
		{"int vs bool false", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typutil.Equal(tt.a, tt.b); got != tt.want {
				t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEqualStringsAndBytes(t *testing.T) {
	t.Run("string vs string equal", func(t *testing.T) {
		if got := typutil.Equal("hello", "hello"); !got {
			t.Errorf("Equal(\"hello\", \"hello\") = %v, want true", got)
		}
	})

	t.Run("string vs string unequal", func(t *testing.T) {
		if got := typutil.Equal("hello", "world"); got {
			t.Errorf("Equal(\"hello\", \"world\") = %v, want false", got)
		}
	})
}

func TestEqualUnknownTypes(t *testing.T) {
	type customType struct {
		value int
	}

	a := customType{value: 42}
	b := customType{value: 42}
	c := customType{value: 0}

	if !typutil.Equal(a, b) {
		t.Errorf("Equal with same custom type values should be true")
	}
	if typutil.Equal(a, c) {
		t.Errorf("Equal with different custom type values should be false")
	}
}
