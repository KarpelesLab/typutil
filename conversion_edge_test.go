package typutil_test

import (
	"encoding/json"
	"testing"

	"github.com/KarpelesLab/typutil"
)

// Test edge cases for type conversion functions to improve coverage

func TestAsIntEdgeCases(t *testing.T) {
	// Test with json.Number
	jsonNum := json.Number("123")
	result, ok := typutil.AsInt(jsonNum)
	if !ok || result != 123 {
		t.Errorf("AsInt with json.Number failed: got %d, %v", result, ok)
	}

	// Test with very large uint64 (high bit set)
	largeUint := uint64(1 << 63)
	result2, ok := typutil.AsInt(largeUint)
	if ok {
		t.Errorf("AsInt should return false for uint64 with high bit set")
	}
	_ = result2

	// Test uintptr
	uintptrVal := uintptr(100)
	result3, ok := typutil.AsInt(uintptrVal)
	if !ok || result3 != 100 {
		t.Errorf("AsInt with uintptr failed: got %d, %v", result3, ok)
	}
}

func TestAsUintEdgeCases(t *testing.T) {
	// Test negative int8
	negInt8 := int8(-5)
	result, ok := typutil.AsUint(negInt8)
	if ok {
		t.Errorf("AsUint should return false for negative int8")
	}
	_ = result

	// Test negative int16
	negInt16 := int16(-100)
	result2, ok := typutil.AsUint(negInt16)
	if ok {
		t.Errorf("AsUint should return false for negative int16")
	}
	_ = result2

	// Test negative int32
	negInt32 := int32(-1000)
	result3, ok := typutil.AsUint(negInt32)
	if ok {
		t.Errorf("AsUint should return false for negative int32")
	}
	_ = result3

	// Test negative int64
	negInt64 := int64(-10000)
	result4, ok := typutil.AsUint(negInt64)
	if ok {
		t.Errorf("AsUint should return false for negative int64")
	}
	_ = result4

	// Test negative int
	negInt := -50
	result5, ok := typutil.AsUint(negInt)
	if ok {
		t.Errorf("AsUint should return false for negative int")
	}
	_ = result5

	// Test with json.Number
	jsonNum := json.Number("456")
	result6, ok := typutil.AsUint(jsonNum)
	if !ok || result6 != 456 {
		t.Errorf("AsUint with json.Number failed: got %d, %v", result6, ok)
	}
}

func TestAsFloatEdgeCases(t *testing.T) {
	// Test with uintptr
	uintptrVal := uintptr(42)
	result, ok := typutil.AsFloat(uintptrVal)
	if !ok || result != 42.0 {
		t.Errorf("AsFloat with uintptr failed: got %f, %v", result, ok)
	}

	// Test fallback to AsInt
	testVal := true
	result2, ok := typutil.AsFloat(testVal)
	if !ok || result2 != 1.0 {
		t.Errorf("AsFloat fallback for bool failed: got %f, %v", result2, ok)
	}
}

func TestAsNumberEdgeCases(t *testing.T) {
	// Test with uintptr
	uintptrVal := uintptr(42)
	result, ok := typutil.AsNumber(uintptrVal)
	if !ok {
		t.Errorf("AsNumber should succeed for uintptr")
	}
	if result != uint64(42) {
		t.Errorf("expected uint64(42), got %v (%T)", result, result)
	}

	// Test string parsing priority (int first)
	result2, ok := typutil.AsNumber("123")
	if !ok {
		t.Errorf("AsNumber('123') should succeed")
	}
	if result2 != int64(123) {
		t.Errorf("expected int64(123), got %v (%T)", result2, result2)
	}

	// Test string parsing (uint when int doesn't fit)
	result3, ok := typutil.AsNumber("18446744073709551615") // max uint64
	if !ok {
		t.Errorf("AsNumber(max uint64 string) should succeed")
	}
	if _, isUint := result3.(uint64); !isUint {
		t.Errorf("expected uint64, got %T", result3)
	}

	// Test string parsing (float when neither int nor uint fit)
	result4, ok := typutil.AsNumber("3.14159")
	if !ok {
		t.Errorf("AsNumber('3.14159') should succeed")
	}
	if _, isFloat := result4.(float64); !isFloat {
		t.Errorf("expected float64, got %T", result4)
	}

	// Test invalid string (should still return a value, but ok=false)
	result5, ok := typutil.AsNumber("notanumber")
	if ok {
		t.Errorf("AsNumber('notanumber') should return ok=false")
	}
	_ = result5
}

func TestAsByteArrayEdgeCases(t *testing.T) {
	// Test with string
	str := "test"
	result, ok := typutil.AsByteArray(str)
	if !ok {
		t.Errorf("AsByteArray with string should succeed")
	}
	if string(result) != "test" {
		t.Errorf("expected 'test', got %q", string(result))
	}

	// Test with int type on 32-bit and 64-bit systems
	intVal := int(42)
	result2, ok := typutil.AsByteArray(intVal)
	if !ok {
		t.Errorf("AsByteArray with int should succeed")
	}
	if len(result2) != 4 && len(result2) != 8 {
		t.Errorf("expected 4 or 8 bytes for int, got %d", len(result2))
	}

	// Test with uint
	uintVal := uint(42)
	result3, ok := typutil.AsByteArray(uintVal)
	if !ok {
		t.Errorf("AsByteArray with uint should succeed")
	}
	if len(result3) != 4 && len(result3) != 8 {
		t.Errorf("expected 4 or 8 bytes for uint, got %d", len(result3))
	}

	// Test with float32
	float32Val := float32(3.14)
	result4, ok := typutil.AsByteArray(float32Val)
	if !ok {
		t.Errorf("AsByteArray with float32 should succeed")
	}
	if len(result4) == 0 {
		t.Errorf("expected non-empty byte array for float32")
	}

	// Test with complex64
	complex64Val := complex64(1 + 2i)
	result5, ok := typutil.AsByteArray(complex64Val)
	if !ok {
		t.Errorf("AsByteArray with complex64 should succeed")
	}
	if len(result5) == 0 {
		t.Errorf("expected non-empty byte array for complex64")
	}

	// Test with complex128
	complex128Val := complex128(1 + 2i)
	result6, ok := typutil.AsByteArray(complex128Val)
	if !ok {
		t.Errorf("AsByteArray with complex128 should succeed")
	}
	if len(result6) == 0 {
		t.Errorf("expected non-empty byte array for complex128")
	}

	// Test with struct (should use fmt.Sprintf fallback)
	type TestStruct struct {
		Value int
	}
	structVal := TestStruct{Value: 42}
	result7, ok := typutil.AsByteArray(structVal)
	if ok {
		t.Errorf("AsByteArray with struct should return ok=false")
	}
	if len(result7) == 0 {
		t.Errorf("expected fallback string representation")
	}
}

func TestAsStringEdgeCases(t *testing.T) {
	// Test with byte slice
	byteSlice := []byte("test string")
	result, ok := typutil.AsString(byteSlice)
	if !ok || result != "test string" {
		t.Errorf("AsString with byte slice failed: got %q, %v", result, ok)
	}

	// Test all integer types
	int16Val := int16(123)
	result2, ok := typutil.AsString(int16Val)
	if !ok || result2 != "123" {
		t.Errorf("AsString with int16 failed: got %q, %v", result2, ok)
	}

	int8Val := int8(45)
	result3, ok := typutil.AsString(int8Val)
	if !ok || result3 != "45" {
		t.Errorf("AsString with int8 failed: got %q, %v", result3, ok)
	}

	uint32Val := uint32(999)
	result4, ok := typutil.AsString(uint32Val)
	if !ok || result4 != "999" {
		t.Errorf("AsString with uint32 failed: got %q, %v", result4, ok)
	}

	uint16Val := uint16(555)
	result5, ok := typutil.AsString(uint16Val)
	if !ok || result5 != "555" {
		t.Errorf("AsString with uint16 failed: got %q, %v", result5, ok)
	}

	uint8Val := uint8(200)
	result6, ok := typutil.AsString(uint8Val)
	if !ok || result6 != "200" {
		t.Errorf("AsString with uint8 failed: got %q, %v", result6, ok)
	}

	// Test with struct (fallback to fmt.Sprintf)
	type TestStruct struct {
		Value int
	}
	structVal := TestStruct{Value: 42}
	result7, ok := typutil.AsString(structVal)
	if ok {
		t.Errorf("AsString with struct should return ok=false")
	}
	if result7 == "" {
		t.Errorf("expected fallback string representation, got empty")
	}
}

func TestMapToMapConversion(t *testing.T) {
	// Test map with different key/value types
	sourceMap := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	var destMap map[string]string
	err := typutil.Assign(&destMap, sourceMap)
	if err != nil {
		t.Errorf("map conversion failed: %s", err)
	}

	if destMap["one"] != "1" || destMap["two"] != "2" || destMap["three"] != "3" {
		t.Errorf("unexpected map values: %v", destMap)
	}
}

func TestStructToMapConversion(t *testing.T) {
	type Source struct {
		Name  string
		Age   int
		Email string `json:"email_address"`
	}

	source := Source{
		Name:  "Alice",
		Age:   30,
		Email: "alice@example.com",
	}

	var dest map[string]any
	err := typutil.Assign(&dest, source)
	if err != nil {
		t.Errorf("struct to map conversion failed: %s", err)
	}

	if dest["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %v", dest["Name"])
	}
	if dest["Age"] != 30 {
		t.Errorf("expected Age=30, got %v", dest["Age"])
	}
	if dest["email_address"] != "alice@example.com" {
		t.Errorf("expected email_address='alice@example.com', got %v", dest["email_address"])
	}
}

func TestPointerConversions(t *testing.T) {
	// Test pointer unwrapping with type conversion
	str := "99"
	var resultInt int
	err := typutil.Assign(&resultInt, &str)
	if err != nil {
		t.Errorf("pointer with conversion failed: %s", err)
	}
	if resultInt != 99 {
		t.Errorf("expected 99, got %d", resultInt)
	}
}

func TestBoolConversions(t *testing.T) {
	// Test various bool conversions
	tests := []struct {
		input    any
		expected bool
	}{
		{true, true},
		{false, false},
		{1, true},
		{0, false},
		{int64(1), true},
		{int64(0), false},
		{uint64(1), true},
		{uint64(0), false},
		{float64(1.0), true},
		{float64(0.0), false},
		{"yes", true},
		{"", false},
		{"0", false},
		{"1", true},
		{[]byte("x"), true},
		{[]byte("0"), false},
		{[]byte{}, false},
	}

	for _, tt := range tests {
		var result bool
		err := typutil.Assign(&result, tt.input)
		if err != nil {
			t.Errorf("bool conversion failed for %v (%T): %s", tt.input, tt.input, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("bool conversion for %v (%T): expected %v, got %v", tt.input, tt.input, tt.expected, result)
		}
	}
}
