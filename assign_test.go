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
	var arg any
	arg = map[string]any{"A": "from a map", "X": 99, "C": 123.456}
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
