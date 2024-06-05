package typutil_test

import (
	"bytes"
	"testing"

	"github.com/KarpelesLab/typutil"
)

type dupTestSruct struct {
	A []byte
	B string
	C *int
	d *int
	E map[string]string
	F []string
}

func TestDup(t *testing.T) {
	v := 42
	w := 1337

	a := &dupTestSruct{
		A: []byte("hello"),
		B: "world",
		C: &v,
		d: &w,
		E: map[string]string{"foo": "bar"},
	}

	b := typutil.DeepClone(a)

	if !bytes.Equal(a.A, b.A) {
		t.Errorf("b should be equal a")
	}

	if b.B != "world" {
		t.Errorf("b.B should equal world")
	}

	b.A[0] = 'H'

	if bytes.Equal(a.A, b.A) {
		t.Errorf("b should not be equal a")
	}

	if a.C == b.C {
		t.Errorf("a.C should not equal b.C")
	}

	*b.C = 43

	if v != 42 {
		t.Errorf("b.C should not affect v")
	}

	if b.d == nil {
		t.Errorf("b.d should not be nil")
	} else if a.d == b.d {
		t.Errorf("a.d should not equal b.d")
	} else if *b.d != 1337 {
		t.Errorf("b.d should equal 1337")
	}
}
