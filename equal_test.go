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
		&equalTestVector{1, 1, true},
		&equalTestVector{1, 0, false},
		&equalTestVector{42, "42", true},
		&equalTestVector{5, 5.0, true},
		&equalTestVector{5, 5.5, false},
		&equalTestVector{5.5, 5, false},
	}

	for _, sub := range v {
		res := typutil.Equal(sub.a, sub.b)
		if res != sub.res {
			t.Errorf("equal failed: %v == %v should return %v, got %v", sub.a, sub.b, sub.res, res)
		}
	}
}
