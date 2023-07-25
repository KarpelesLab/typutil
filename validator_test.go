package typutil_test

import (
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
		return
	}
}
