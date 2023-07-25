package typutil_test

import (
	"strings"
	"testing"

	"github.com/KarpelesLab/typutil"
)

type valA struct {
	A string `validator:"notempty"`
}

func TestValidator(t *testing.T) {
	var a *valA

	err := typutil.Assign(&a, map[string]any{"A": ""})
	if err == nil || err != typutil.ErrEmptyValue {
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
}
