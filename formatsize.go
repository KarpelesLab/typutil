package typutil

import (
	"bytes"
	"math/bits"
	"strconv"
)

var units = []byte{0, 'K', 'M', 'G', 'T', 'P', 'E'}

// FormatSize formats a byte size as a human-readable string with appropriate units.
//
// This function converts a raw byte count into a formatted string using binary prefixes
// (KiB, MiB, GiB, etc.) according to IEC standards (powers of 1024).
//
// The formatted string includes:
// - The integer part
// - A decimal point
// - Two decimal places
// - The appropriate binary unit (B, KiB, MiB, GiB, TiB, PiB, EiB)
//
// For values less than 1024 bytes, the function returns a simple byte count without decimals.
//
// Parameters:
//   - x: The size in bytes to format
//
// Returns:
//   - A human-readable string representing the size
//
// Examples:
//   - FormatSize(0) → "0 B"
//   - FormatSize(1023) → "1023 B"
//   - FormatSize(1024) → "1.00 KiB"
//   - FormatSize(1536) → "1.50 KiB"
//   - FormatSize(1048576) → "1.00 MiB"
//   - FormatSize(1073741824) → "1.00 GiB"
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
