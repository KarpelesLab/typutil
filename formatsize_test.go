package typutil_test

import (
	"testing"

	"github.com/KarpelesLab/typutil"
)

type fmtSizeTestV struct {
	in  uint64
	out string
}

func TestFormatSize(t *testing.T) {
	testV := []*fmtSizeTestV{
		&fmtSizeTestV{1, "1 B"},
		&fmtSizeTestV{0, "0 B"},
		&fmtSizeTestV{1000, "1000 B"},
		&fmtSizeTestV{1025, "1.00 KiB"},
		&fmtSizeTestV{2000, "1.95 KiB"},
		&fmtSizeTestV{2047, "1.99 KiB"},
		&fmtSizeTestV{2048, "2.00 KiB"},
		&fmtSizeTestV{1000000, "976.56 KiB"},
		&fmtSizeTestV{123456789123456789, "109.65 PiB"},
	}

	for _, test := range testV {
		res := typutil.FormatSize(test.in)
		if res != test.out {
			t.Errorf("test failed for %d: got %s instead of %s", test.in, res, test.out)
		}
	}
}
