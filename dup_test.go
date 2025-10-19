package typutil_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/KarpelesLab/typutil"
)

type dupTestSruct struct {
	A  []byte
	B  string
	C  *int
	d  *int
	e  *int
	E  map[string]string
	F  []string
	X  any
	Y  any
	z  string
	t1 time.Time
	t2 time.Time
}

func TestDup(t *testing.T) {
	v := 42
	w := 1337

	loc := time.FixedZone("UTC-8", -8*60*60)
	a := &dupTestSruct{
		A:  []byte("hello"),
		B:  "world",
		C:  &v,
		d:  &w,
		e:  &v,
		E:  map[string]string{"foo": "bar"},
		X:  w,
		z:  "are you here?",
		t1: time.Now().In(loc),
		t2: time.Now().In(loc),
	}

	b := typutil.DeepClone(*a)

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
	if b.C != b.e {
		t.Errorf("b.C should equal b.e (same pointer)")
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
