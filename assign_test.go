package typutil_test

import (
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
}

type objB struct {
	A string
	B float64 `json:"X"`
	C int32
}

func TestAssignAs(t *testing.T) {
	a, err := typutil.As[int]("42")

	if err != nil {
		t.Errorf("got error = %s", err)
	}

	if a != 42 {
		t.Errorf("expected 42, got %v", a)
	}
}

func TestAssignStruct(t *testing.T) {
	a := &objA{A: "hello world", B: 42, C: 123456.789321}
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
	if b.C != 123456 {
		t.Errorf("struct to struct C invalid value: %v", b.C)
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
