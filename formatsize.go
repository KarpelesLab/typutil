package typutil

import (
	"bytes"
	"math/bits"
	"strconv"
)

var units = []byte{0, 'K', 'M', 'G', 'T', 'P', 'E'}

// FormatSize is a simple method that will format a given integer value into something easier to read as a human
func FormatSize(x uint64) string {
	if x == 0 {
		return "0 B"
	}

	bitsLen := bits.Len64(x)
	index := (bitsLen - 1) / 10

	if index >= len(units) {
		index = len(units) - 1
	}
	if index == 0 {
		// if byte value, do not add a decimal part and return as is
		return strconv.FormatUint(x, 10) + " B"
	}

	// compute integer & fraction parts
	e := index * 10
	unit := units[index]
	integer_part := x >> e
	fraction_numerator := (x-(integer_part<<e))*100 + 50
	fraction := fraction_numerator / (1 << e)

	// generate text buffer
	buf := &bytes.Buffer{}
	buf.WriteString(strconv.FormatUint(integer_part, 10))
	buf.WriteByte('.')

	fraction_str := strconv.FormatUint(fraction, 10)
	if len(fraction_str) == 1 {
		buf.WriteByte('0')
	}
	buf.WriteString(fraction_str)

	//log.Printf("x=%d e=%d integer_part=%d fraction_numerator=%d fraction=%d", x, e, integer_part, fraction_numerator, fraction)
	return string(append(buf.Bytes(), ' ', unit, 'i', 'B'))
}
