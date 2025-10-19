package typutil_test

import (
	"testing"

	"github.com/KarpelesLab/typutil"
)

// Test slice to slice conversions
func TestSliceToSlice(t *testing.T) {
	// Test int slice to string slice
	intSlice := []int{1, 2, 3, 4, 5}
	var strSlice []string
	err := typutil.Assign(&strSlice, intSlice)
	if err != nil {
		t.Errorf("int slice to string slice failed: %s", err)
	}
	if len(strSlice) != 5 {
		t.Errorf("expected length 5, got %d", len(strSlice))
	}
	if strSlice[0] != "1" || strSlice[4] != "5" {
		t.Errorf("unexpected values: %v", strSlice)
	}

	// Test string slice to int slice
	strInput := []string{"10", "20", "30"}
	var intOutput []int
	err = typutil.Assign(&intOutput, strInput)
	if err != nil {
		t.Errorf("string slice to int slice failed: %s", err)
	}
	if len(intOutput) != 3 {
		t.Errorf("expected length 3, got %d", len(intOutput))
	}
	if intOutput[0] != 10 || intOutput[2] != 30 {
		t.Errorf("unexpected values: %v", intOutput)
	}

	// Test float slice to int slice
	floatSlice := []float64{1.5, 2.5, 3.5}
	var intSlice2 []int
	err = typutil.Assign(&intSlice2, floatSlice)
	if err != nil {
		t.Errorf("float slice to int slice failed: %s", err)
	}
	if len(intSlice2) != 3 {
		t.Errorf("expected length 3, got %d", len(intSlice2))
	}
	// Rounding should occur (1.5->2, 2.5->2 or 3, 3.5->4)
	if intSlice2[0] != 2 || intSlice2[2] != 4 {
		t.Errorf("unexpected values (expected rounding): %v", intSlice2)
	}
}

func TestSliceOfStructs(t *testing.T) {
	type Source struct {
		Name string
		Age  int
	}

	type Dest struct {
		Name string
		Age  string
	}

	source := []Source{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	var dest []Dest
	err := typutil.Assign(&dest, source)
	if err != nil {
		t.Errorf("slice of structs conversion failed: %s", err)
	}

	if len(dest) != 2 {
		t.Errorf("expected length 2, got %d", len(dest))
	}

	if dest[0].Name != "Alice" || dest[0].Age != "30" {
		t.Errorf("unexpected dest[0]: %+v", dest[0])
	}

	if dest[1].Name != "Bob" || dest[1].Age != "25" {
		t.Errorf("unexpected dest[1]: %+v", dest[1])
	}
}

func TestEmptySlice(t *testing.T) {
	var emptyInt []int
	var emptyStr []string
	err := typutil.Assign(&emptyStr, emptyInt)
	if err != nil {
		t.Errorf("empty slice conversion failed: %s", err)
	}
	if len(emptyStr) != 0 {
		t.Errorf("expected empty slice, got length %d", len(emptyStr))
	}
}

func TestAsSlice(t *testing.T) {
	// Test As with slices
	intSlice := []int{1, 2, 3}
	strSlice, err := typutil.As[[]string](intSlice)
	if err != nil {
		t.Errorf("As[[]string] failed: %s", err)
	}
	if len(strSlice) != 3 {
		t.Errorf("expected length 3, got %d", len(strSlice))
	}
	if strSlice[0] != "1" || strSlice[2] != "3" {
		t.Errorf("unexpected values: %v", strSlice)
	}
}

func TestByteSliceConversions(t *testing.T) {
	// Test string to []byte (base64 decoding)
	base64Str := "SGVsbG8gV29ybGQ=" // "Hello World" in base64
	var byteSlice []byte
	err := typutil.Assign(&byteSlice, base64Str)
	if err != nil {
		t.Errorf("string to []byte failed: %s", err)
	}
	if string(byteSlice) != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", string(byteSlice))
	}

	// Test invalid base64
	invalidBase64 := "not valid base64!!!"
	var byteSlice2 []byte
	err = typutil.Assign(&byteSlice2, invalidBase64)
	if err == nil {
		t.Errorf("expected error for invalid base64, got nil")
	}

	// Test []byte to string (base64 encoding)
	bytes := []byte("Test Data")
	var str string
	err = typutil.Assign(&str, bytes)
	if err != nil {
		t.Errorf("[]byte to string failed: %s", err)
	}
	// Should be base64 encoded
	if str != "VGVzdCBEYXRh" {
		t.Errorf("expected base64 'VGVzdCBEYXRh', got %q", str)
	}
}

func TestSliceCapacityHandling(t *testing.T) {
	// Test that Assign properly handles slice capacity
	source := []int{1, 2, 3, 4, 5}

	// Start with nil slice (should allocate as needed)
	var dest []string

	err := typutil.Assign(&dest, source)
	if err != nil {
		t.Errorf("slice assignment failed: %s", err)
	}

	if len(dest) != 5 {
		t.Errorf("expected length 5, got %d", len(dest))
	}

	if dest[0] != "1" || dest[4] != "5" {
		t.Errorf("unexpected values: %v", dest)
	}
}

func TestNestedSlices(t *testing.T) {
	// Test slice of slices
	source := [][]int{{1, 2}, {3, 4}, {5, 6}}
	var dest [][]string

	err := typutil.Assign(&dest, source)
	if err != nil {
		t.Errorf("nested slice conversion failed: %s", err)
	}

	if len(dest) != 3 {
		t.Errorf("expected length 3, got %d", len(dest))
	}

	if len(dest[0]) != 2 || dest[0][0] != "1" || dest[0][1] != "2" {
		t.Errorf("unexpected dest[0]: %v", dest[0])
	}

	if len(dest[2]) != 2 || dest[2][0] != "5" || dest[2][1] != "6" {
		t.Errorf("unexpected dest[2]: %v", dest[2])
	}
}
