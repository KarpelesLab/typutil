package typutil_test

import (
	"testing"

	"github.com/KarpelesLab/typutil"
)

// Test examples from README
func TestReadmeExamples(t *testing.T) {
	// Example 1: Basic As conversion
	result, err := typutil.As[int]("42")
	if err != nil {
		t.Errorf("As[int] failed: %s", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}

	// Example 2: Map to struct
	type User struct {
		Name string
		Age  int
	}

	m := map[string]any{"Name": "Alice", "Age": 30}
	user, err := typutil.As[User](m)
	if err != nil {
		t.Errorf("As[User] failed: %s", err)
	}
	if user.Name != "Alice" || user.Age != 30 {
		t.Errorf("unexpected user: %+v", user)
	}

	// Example 3: Assign
	var age int
	err = typutil.Assign(&age, "42")
	if err != nil {
		t.Errorf("Assign failed: %s", err)
	}
	if age != 42 {
		t.Errorf("expected age=42, got %d", age)
	}

	// Example 4: Basic type conversions
	i, _ := typutil.As[int]("123")
	if i != 123 {
		t.Errorf("expected 123, got %d", i)
	}

	f, _ := typutil.As[float64]("3.14")
	if f != 3.14 {
		t.Errorf("expected 3.14, got %f", f)
	}

	s, _ := typutil.As[string](42)
	if s != "42" {
		t.Errorf("expected '42', got %q", s)
	}

	b, _ := typutil.As[bool](1)
	if !b {
		t.Errorf("expected true, got %v", b)
	}
}

func TestReadmeStructConversion(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	type Employee struct {
		Name string
		Age  string
	}

	p := Person{Name: "Alice", Age: 30}
	e, err := typutil.As[Employee](p)
	if err != nil {
		t.Errorf("struct conversion failed: %s", err)
	}
	if e.Name != "Alice" || e.Age != "30" {
		t.Errorf("unexpected employee: %+v", e)
	}
}

func TestReadmeMapToStructWithTags(t *testing.T) {
	type Config struct {
		Host     string
		Port     int
		Timeout  float64
		Username string `json:"user"`
		Password string `json:"pass"`
	}

	m := map[string]any{
		"Host":    "localhost",
		"Port":    "8080", // String converted to int
		"Timeout": 30.5,
		"user":    "admin",  // Matches via json tag
		"pass":    "secret", // Matches via json tag
		"Extra":   "ignored",
	}

	config, err := typutil.As[Config](m)
	if err != nil {
		t.Errorf("config conversion failed: %s", err)
	}
	if config.Host != "localhost" {
		t.Errorf("expected Host='localhost', got %q", config.Host)
	}
	if config.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", config.Port)
	}
	if config.Timeout != 30.5 {
		t.Errorf("expected Timeout=30.5, got %f", config.Timeout)
	}
	if config.Username != "admin" {
		t.Errorf("expected Username='admin', got %q", config.Username)
	}
	if config.Password != "secret" {
		t.Errorf("expected Password='secret', got %q", config.Password)
	}
}

func TestReadmeConversionHelpers(t *testing.T) {
	// AsString
	str, ok := typutil.AsString(42)
	if !ok || str != "42" {
		t.Errorf("AsString(42) failed: got %q, %v", str, ok)
	}

	str, ok = typutil.AsString([]byte{65})
	if !ok || str != "A" {
		t.Errorf("AsString([]byte{65}) failed: got %q, %v", str, ok)
	}

	// AsInt
	num, ok := typutil.AsInt("42")
	if !ok || num != 42 {
		t.Errorf("AsInt('42') failed: got %d, %v", num, ok)
	}

	num, ok = typutil.AsInt(3.14)
	if !ok || num != 3 {
		t.Errorf("AsInt(3.14) failed: got %d, %v", num, ok)
	}

	// AsFloat
	f, ok := typutil.AsFloat("3.14")
	if !ok || f != 3.14 {
		t.Errorf("AsFloat('3.14') failed: got %f, %v", f, ok)
	}

	// AsBool
	b := typutil.AsBool("yes")
	if !b {
		t.Errorf("AsBool('yes') should be true")
	}

	b = typutil.AsBool(0)
	if b {
		t.Errorf("AsBool(0) should be false")
	}

	b = typutil.AsBool("non-empty")
	if !b {
		t.Errorf("AsBool('non-empty') should be true")
	}
}

func TestReadmeStructToMap(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	p := Person{Name: "Alice", Age: 30}

	var m map[string]any
	err := typutil.Assign(&m, p)
	if err != nil {
		t.Errorf("struct to map conversion failed: %s", err)
	}

	if m["Name"] != "Alice" {
		t.Errorf("expected Name='Alice', got %v", m["Name"])
	}
	if m["Age"] != 30 {
		t.Errorf("expected Age=30, got %v", m["Age"])
	}
}
