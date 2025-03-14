package typutil_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestBaseType(t *testing.T) {
	// Define some custom types
	type CustomString string
	type CustomInt int
	type CustomBool bool
	type CustomFloat float64
	type CustomSlice []int

	// Create test values
	customStr := CustomString("hello")
	customInt := CustomInt(42)
	customBool := CustomBool(true)
	customFloat := CustomFloat(3.14)
	customSlice := CustomSlice{1, 2, 3}

	// Create pointers
	strPtr := "pointer"
	strPtrPtr := &strPtr
	customStrPtr := &customStr

	tests := []struct {
		name string
		v    interface{}
		want interface{}
	}{
		{"string", "hello", "hello"},
		{"custom string", customStr, "hello"},
		{"int", 42, int64(42)},
		{"custom int", customInt, int64(42)},
		{"bool", true, true},
		{"custom bool", customBool, true},
		{"float", 3.14, 3.14},
		{"custom float", customFloat, 3.14},
		{"slice", []int{1, 2, 3}, []int{1, 2, 3}},
		{"custom slice", customSlice, customSlice}, // Slices stay as their original type
		{"string pointer", &strPtr, "pointer"},
		{"double string pointer", &strPtrPtr, "pointer"},
		{"custom string pointer", customStrPtr, "hello"},
		{"nil", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typutil.BaseType(tt.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseType(%v) = %v (%T), want %v (%T)", tt.v, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestBaseTypeWithReflectValue(t *testing.T) {
	// Test with reflect.Value
	str := "reflect-value"
	reflectVal := reflect.ValueOf(str)
	got := typutil.BaseType(reflectVal)

	if got != str {
		t.Errorf("BaseType(reflect.ValueOf(%v)) = %v, want %v", str, got, str)
	}
}

func TestBaseTypeWithJsonRawMessage(t *testing.T) {
	// Test with json.RawMessage
	jsonStr := json.RawMessage(`"json-string"`)
	got := typutil.BaseType(jsonStr)

	if got != "json-string" {
		t.Errorf("BaseType(json.RawMessage(%v)) = %v, want %v", string(jsonStr), got, "json-string")
	}

	jsonInt := json.RawMessage(`42`)
	got = typutil.BaseType(jsonInt)

	// The json parser will decode this as a float64 by default
	if got != float64(42) {
		t.Errorf("BaseType(json.RawMessage(%v)) = %v (%T), want %v (%T)", string(jsonInt), got, got, float64(42), float64(42))
	}

	jsonBool := json.RawMessage(`true`)
	got = typutil.BaseType(jsonBool)

	if got != true {
		t.Errorf("BaseType(json.RawMessage(%v)) = %v, want %v", string(jsonBool), got, true)
	}

	jsonArray := json.RawMessage(`[1,2,3]`)
	got = typutil.BaseType(jsonArray)

	expected := []interface{}{float64(1), float64(2), float64(3)}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("BaseType(json.RawMessage(%v)) = %v, want %v", string(jsonArray), got, expected)
	}
}

func TestBaseTypeWithNestedStructures(t *testing.T) {
	type Inner struct {
		Value string
	}

	type Outer struct {
		Inner *Inner
	}

	inner := Inner{Value: "test"}
	outer := Outer{Inner: &inner}

	// BaseType doesn't unwrap struct fields, so it should return the struct as is
	got := typutil.BaseType(outer)
	if !reflect.DeepEqual(got, outer) {
		t.Errorf("BaseType(%v) = %v, want %v", outer, got, outer)
	}
}
