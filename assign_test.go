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
