package typutil_test

import (
	"errors"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestAssignSimple(t *testing.T) {
	var a string
	err := typutil.Assign(&a, "hello world")

	if err != nil {
		t.Errorf("simple assign failed: %s", err)
	} else if a != "hello world" {
		t.Errorf("unexpected value %v", a)
	}

	var b *string
	err = typutil.Assign(&b, "hello too")

	if err != nil {
		t.Errorf("simple assign failed: %s", err)
	} else if *b != "hello too" {
		t.Errorf("unexpected value %v", a)
	}
}

type objA struct {
	A string
	B int `json:"X"`
	C float64
	D float64
}

type objB struct {
	A string
	B float64 `json:"X"`
	C int32
	D uint16
}

func TestAssignAs(t *testing.T) {
	a, err := typutil.As[int]("42")

	if err != nil {
		t.Errorf("got error = %s", err)
	}

	if a != 42 {
		t.Errorf("expected 42, got %v", a)
	}

	b, err := typutil.As[int](nil)
	if !errors.Is(err, typutil.ErrInvalidSource) {
		t.Errorf("got unexpected error = %s", err)
	}

	if b != 0 {
		t.Errorf("expected 0, got %v", b)
	}

	c, err := typutil.As[string](42)
	if err != nil {
		t.Errorf("got error = %s", err)
	}

	if c != "42" {
		t.Errorf("expected 42, got %v", c)
	}
}

func TestAssignAssignTo(t *testing.T) {
	a, err := typutil.As[objA](typutil.RawJsonMessage(`{"A":"value A","X":42,"C":123.456,"D":1e-19}`))
	if err != nil {
		t.Errorf("got error = %s", err)
		return
	}

	if a.A != "value A" {
		t.Errorf("unexpected value for a.A: %s", a.A)
	}
	if a.B != 42 {
		t.Errorf("unexpected value for a.B: %v", a.B)
	}
}

func TestAssignStruct(t *testing.T) {
	a := &objA{A: "hello world", B: 42, C: 123456.789321, D: 555.22}
	var b *objB

	err := typutil.Assign(&b, a)
	if err != nil {
		t.Errorf("struct to struct assign failed: %s", err)
		return
	}

	if b.A != "hello world" {
		t.Errorf("struct to struct string missing")
	}
	if b.B != 42 {
		t.Errorf("struct to struct B missing")
	}
	if b.C != 123457 {
		t.Errorf("struct to struct C invalid value: %v", b.C)
	}
	if b.D != 555 {
		t.Errorf("struct to struct D invalid value: %v", b.D)
	}

	// test assign from map
	b = nil
	var arg any
	arg = map[string]any{"A": "from a map", "X": 99, "C": 123.456}
	err = typutil.Assign(&b, &arg)
	if err != nil {
		t.Errorf("map to struct assign failed: %s", err)
		return
	}

	if b.A != "from a map" {
		t.Errorf("map to struct string invalid")
	}
	if b.B != 99 {
		t.Errorf("map to struct B invalid")
	}
	if b.C != 123 {
		t.Errorf("map to struct C invalid")
	}

	var m map[string]any

	err = typutil.Assign(&m, b)
	if err != nil {
		t.Errorf("struct to map assign failed: %s", err)
		return
	}

	if m["A"].(string) != "from a map" {
		t.Errorf("struct to map A failed")
	}
}

func TestAsMapToStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Age     int
		Active  bool
		Score   float64
		Tag     string `json:"customTag"`
		Missing string // not in map, should be zero value
	}

	m := map[string]any{
		"Name":      "Alice",
		"Age":       "30",      // string that should convert to int
		"Active":    true,
		"Score":     95.5,
		"customTag": "tagged",  // should match via json tag
		"Extra":     "ignored", // should be ignored
	}

	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("As[TestStruct](map[string]any) failed: %s", err)
		return
	}

	if result.Name != "Alice" {
		t.Errorf("expected Name='Alice', got %q", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("expected Age=30, got %d", result.Age)
	}
	if !result.Active {
		t.Errorf("expected Active=true, got %v", result.Active)
	}
	if result.Score != 95.5 {
		t.Errorf("expected Score=95.5, got %f", result.Score)
	}
	if result.Tag != "tagged" {
		t.Errorf("expected Tag='tagged' (via json tag), got %q", result.Tag)
	}
	if result.Missing != "" {
		t.Errorf("expected Missing='' (zero value), got %q", result.Missing)
	}
}

func TestAsMapToStructPointer(t *testing.T) {
	type TestStruct struct {
		Value string
	}

	m := map[string]any{"Value": "test"}

	// Test that As returns the struct value, not a pointer
	result, err := typutil.As[TestStruct](m)
	if err != nil {
		t.Errorf("As[TestStruct](map) failed: %s", err)
		return
	}

	if result.Value != "test" {
		t.Errorf("expected Value='test', got %q", result.Value)
	}

	// Test with pointer type
	ptrResult, err := typutil.As[*TestStruct](m)
	if err != nil {
		t.Errorf("As[*TestStruct](map) failed: %s", err)
		return
	}

	if ptrResult == nil {
		t.Errorf("expected non-nil pointer")
		return
	}

	if ptrResult.Value != "test" {
		t.Errorf("expected Value='test', got %q", ptrResult.Value)
	}
}
