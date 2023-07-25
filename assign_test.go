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
	B int
	C float64
}

type objB struct {
	A string
	B float64
	C int32
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
}
