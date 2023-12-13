package typutil_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/KarpelesLab/typutil"
)

type valA struct {
	A string `validator:"not_empty"`
}

func TestValidator(t *testing.T) {
	var a *valA

	err := typutil.Assign(&a, map[string]any{"A": ""})
	if err == nil || !errors.Is(err, typutil.ErrEmptyValue) {
		t.Errorf("struct to struct assign failed: %s", err)
	}

	err = typutil.Assign(&a, map[string]any{"A": "hello"})
	if err != nil {
		t.Errorf("struct to struct assign failed: %s", err)
		return
	}

	if a.A != "hello" {
		t.Errorf("assign failed")
	}

	err = typutil.Validate(a)
	if err != nil {
		t.Errorf("validate failed: %s", err)
	}

	a.A = ""
	err = typutil.Validate(a)
	if err == nil || !errors.Is(err, typutil.ErrEmptyValue) {
		t.Errorf("validate error failed: %v", err)
	}
}

type valB struct {
	X string `validator:"valb_test"`
}

func TestPtrValidator(t *testing.T) {
	var b *valB

	typutil.SetValidator("valb_test", func(s *string) error {
		*s = strings.ToUpper(*s)
		return nil
	})

	err := typutil.Assign(&b, map[string]any{"X": "Value"})
	if err != nil {
		t.Errorf("struct to struct assign failed: %s", err)
		return
	}

	if b.X != "VALUE" {
		t.Errorf("validator alter failed, got %s", b.X)
	}

	b.X = "another value"
	err = typutil.Validate(b)
	if err != nil {
		t.Errorf("failed to validate struct: %s", err)
	}
	if b.X != "ANOTHER VALUE" {
		t.Errorf("unexpected X value: %s", b.X)
	}
}

type valC struct {
	A string `validator:"minlength=3"`
}

func TestArgValidator(t *testing.T) {
	var c *valC

	err := typutil.Assign(&c, map[string]any{"A": "b"})
	if err == nil {
		t.Errorf("should have had an error")
	} else if err.Error() != "on field A: string must be at least 3 characters" {
		t.Errorf("unexpected error %s", err)
	}

	err = typutil.Assign(&c, map[string]any{"A": "boo"})
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
}
