package typutil_test

import (
	"testing"

	"github.com/KarpelesLab/typutil"
)

// Test validators with 0% coverage
func TestMaxLengthValidator(t *testing.T) {
	type TestStruct struct {
		Name string `validator:"maxlength=10"`
	}

	// Valid case - within limit
	m := map[string]any{"Name": "Short"}
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("maxlength validation should pass: %s", err)
	}
	if result.Name != "Short" {
		t.Errorf("expected 'Short', got %q", result.Name)
	}

	// Valid case - exactly at limit
	m2 := map[string]any{"Name": "TenLetters"}
	result2, err := typutil.As[TestStruct](m2)
	if err != nil {
		t.Errorf("maxlength validation should pass for exact length: %s", err)
	}
	if result2.Name != "TenLetters" {
		t.Errorf("expected 'TenLetters', got %q", result2.Name)
	}

	// Invalid case - exceeds limit
	m3 := map[string]any{"Name": "ThisIsTooLongForTheLimit"}
	_, err = typutil.As[TestStruct](m3)
	if err == nil {
		t.Errorf("maxlength validation should fail for long string")
	}
}

func TestIpAddressValidator(t *testing.T) {
	type TestStruct struct {
		IP string `validator:"ip_address"`
	}

	// Valid IPv4
	m := map[string]any{"IP": "192.168.1.1"}
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("ip_address validation should pass for valid IPv4: %s", err)
	}
	if result.IP != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got %q", result.IP)
	}

	// Valid IPv6
	m2 := map[string]any{"IP": "2001:0db8:85a3:0000:0000:8a2e:0370:7334"}
	result2, err := typutil.As[TestStruct](m2)
	if err != nil {
		t.Errorf("ip_address validation should pass for valid IPv6: %s", err)
	}
	if result2.IP != "2001:0db8:85a3:0000:0000:8a2e:0370:7334" {
		t.Errorf("unexpected IP: %q", result2.IP)
	}

	// Valid short IPv6
	m3 := map[string]any{"IP": "::1"}
	result3, err := typutil.As[TestStruct](m3)
	if err != nil {
		t.Errorf("ip_address validation should pass for ::1: %s", err)
	}
	if result3.IP != "::1" {
		t.Errorf("expected '::1', got %q", result3.IP)
	}

	// Invalid IP
	m4 := map[string]any{"IP": "not an ip"}
	_, err = typutil.As[TestStruct](m4)
	if err == nil {
		t.Errorf("ip_address validation should fail for invalid IP")
	}

	// Invalid IP - out of range
	m5 := map[string]any{"IP": "256.256.256.256"}
	_, err = typutil.As[TestStruct](m5)
	if err == nil {
		t.Errorf("ip_address validation should fail for out of range IP")
	}
}

func TestHex6ColorValidator(t *testing.T) {
	type TestStruct struct {
		Color string `validator:"hex6color"`
	}

	// Valid hex color with #  (# will be stripped)
	m := map[string]any{"Color": "#FF5733"}
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("hex6color validation should pass: %s", err)
	}
	if result.Color != "FF5733" {
		t.Errorf("expected 'FF5733' (# stripped), got %q", result.Color)
	}

	// Valid lowercase with # (# will be stripped)
	m2 := map[string]any{"Color": "#abc123"}
	result2, err := typutil.As[TestStruct](m2)
	if err != nil {
		t.Errorf("hex6color validation should pass for lowercase: %s", err)
	}
	if result2.Color != "abc123" {
		t.Errorf("expected 'abc123' (# stripped), got %q", result2.Color)
	}

	// Valid - no hash
	m3 := map[string]any{"Color": "FF5733"}
	result3, err := typutil.As[TestStruct](m3)
	if err != nil {
		t.Errorf("hex6color validation should pass without #: %s", err)
	}
	if result3.Color != "FF5733" {
		t.Errorf("expected 'FF5733', got %q", result3.Color)
	}

	// Invalid - too short
	m4 := map[string]any{"Color": "#FFF"}
	_, err = typutil.As[TestStruct](m4)
	if err == nil {
		t.Errorf("hex6color validation should fail for short color")
	}

	// Invalid - too long
	m5 := map[string]any{"Color": "#FF5733AA"}
	_, err = typutil.As[TestStruct](m5)
	if err == nil {
		t.Errorf("hex6color validation should fail for long color")
	}

	// Invalid - non-hex characters
	m6 := map[string]any{"Color": "#GGGGGG"}
	_, err = typutil.As[TestStruct](m6)
	if err == nil {
		t.Errorf("hex6color validation should fail for non-hex characters")
	}
}

func TestHex64Validator(t *testing.T) {
	type TestStruct struct {
		Hash string `validator:"hex64"`
	}

	// Valid 64-char hex string (SHA256 hash)
	validHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	m := map[string]any{"Hash": validHash}
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("hex64 validation should pass: %s", err)
	}
	if result.Hash != validHash {
		t.Errorf("expected %q, got %q", validHash, result.Hash)
	}

	// Valid uppercase
	validHashUpper := "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855"
	m2 := map[string]any{"Hash": validHashUpper}
	result2, err := typutil.As[TestStruct](m2)
	if err != nil {
		t.Errorf("hex64 validation should pass for uppercase: %s", err)
	}
	if result2.Hash != validHashUpper {
		t.Errorf("expected %q, got %q", validHashUpper, result2.Hash)
	}

	// Invalid - too short
	m3 := map[string]any{"Hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b85"}
	_, err = typutil.As[TestStruct](m3)
	if err == nil {
		t.Errorf("hex64 validation should fail for 63 chars")
	}

	// Invalid - too long
	m4 := map[string]any{"Hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b8555"}
	_, err = typutil.As[TestStruct](m4)
	if err == nil {
		t.Errorf("hex64 validation should fail for 65 chars")
	}

	// Invalid - non-hex characters
	m5 := map[string]any{"Hash": "g3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}
	_, err = typutil.As[TestStruct](m5)
	if err == nil {
		t.Errorf("hex64 validation should fail for non-hex characters")
	}

	// Empty string is allowed (use not_empty to prevent it)
	m6 := map[string]any{"Hash": ""}
	result6, err := typutil.As[TestStruct](m6)
	if err != nil {
		t.Errorf("hex64 validation should pass for empty string: %s", err)
	}
	if result6.Hash != "" {
		t.Errorf("expected empty string, got %q", result6.Hash)
	}
}

func TestMultipleValidatorsWithNewOnes(t *testing.T) {
	type TestStruct struct {
		Email    string `validator:"not_empty,maxlength=50"`
		Password string `validator:"minlength=8,maxlength=128"`
		IP       string `validator:"not_empty,ip_address"`
		Color    string `validator:"hex6color"`
	}

	m := map[string]any{
		"Email":    "user@example.com",
		"Password": "securepass123",
		"IP":       "10.0.0.1",
		"Color":    "#123456",
	}

	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("multiple validators should pass: %s", err)
	}

	if result.Email != "user@example.com" {
		t.Errorf("unexpected email: %q", result.Email)
	}
	if result.Password != "securepass123" {
		t.Errorf("unexpected password: %q", result.Password)
	}
	if result.IP != "10.0.0.1" {
		t.Errorf("unexpected IP: %q", result.IP)
	}
	if result.Color != "123456" {
		t.Errorf("unexpected color: %q (# should be stripped)", result.Color)
	}

	// Test failure cases
	m2 := map[string]any{
		"Email":    "verylongemailaddressthatexceedsthefiftycharlimit@example.com",
		"Password": "short",
		"IP":       "invalid",
		"Color":    "notacolor",
	}

	_, err = typutil.As[TestStruct](m2)
	if err == nil {
		t.Errorf("validators should fail for invalid data")
	}
}

func TestNotEmptyValidatorEdgeCases(t *testing.T) {
	type TestStruct struct {
		Field string `validator:"not_empty"`
	}

	// Empty string should fail
	m := map[string]any{"Field": ""}
	_, err := typutil.As[TestStruct](m)
	if err == nil {
		t.Errorf("not_empty should fail for empty string")
	}

	// Whitespace should pass (it's not empty)
	m2 := map[string]any{"Field": "   "}
	result, err := typutil.As[TestStruct](m2)
	if err != nil {
		t.Errorf("not_empty should pass for whitespace: %s", err)
	}
	if result.Field != "   " {
		t.Errorf("expected whitespace, got %q", result.Field)
	}

	// Non-empty should pass
	m3 := map[string]any{"Field": "value"}
	result2, err := typutil.As[TestStruct](m3)
	if err != nil {
		t.Errorf("not_empty should pass: %s", err)
	}
	if result2.Field != "value" {
		t.Errorf("expected 'value', got %q", result2.Field)
	}
}
