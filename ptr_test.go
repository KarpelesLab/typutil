package typutil_test

import (
	"reflect"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestIsNil(t *testing.T) {
	// Create variables for testing
	var nilPtr *string
	var nilSlice []string
	var nilMap map[string]string
	var nilChan chan int
	var nilFunc func()
	var nilInterface interface{}

	s := "hello"
	notNilPtr := &s
	notNilSlice := []string{}
	notNilMap := map[string]string{}
	notNilChan := make(chan int)
	notNilFunc := func() {}
	notNilInterface := interface{}("hello")

	// Multiple levels of nesting
	var nestedNilPtr ***string
	var temp **string
	nestedNotNilPtr := &temp
	temp2 := &s
	nestedNotFullyNilPtr := &temp2

	tests := []struct {
		name string
		v    interface{}
		want bool
	}{
		{"nil value", nil, true},
		{"nil pointer", nilPtr, true},
		{"nil slice", nilSlice, true},
		{"nil map", nilMap, true},
		{"nil channel", nilChan, true},
		{"nil function", nilFunc, true},
		{"nil interface", nilInterface, true},
		{"string value", "hello", false},
		{"int value", 42, false},
		{"bool value", true, false},
		{"non-nil pointer", notNilPtr, false},
		{"non-nil slice", notNilSlice, false},
		{"non-nil map", notNilMap, false},
		{"non-nil channel", notNilChan, false},
		{"non-nil function", notNilFunc, false},
		{"non-nil interface", notNilInterface, false},
		{"deeply nested nil pointer", nestedNilPtr, true},
		{"partially nested nil pointer", nestedNotNilPtr, true}, // The outer pointer is not nil, but points to a nil pointer
		{"nested non-nil pointer", nestedNotFullyNilPtr, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := typutil.IsNil(tt.v); got != tt.want {
				t.Errorf("IsNil(%v) = %v, want %v", tt.v, got, tt.want)
			}
		})
	}
}

func TestFlatten(t *testing.T) {
	// Create test values
	s := "hello"
	sPtr := &s
	sPtrPtr := &sPtr

	i := 42
	iPtr := &i

	var nilPtr *string

	tests := []struct {
		name string
		a    interface{}
		want interface{}
	}{
		{"nil", nil, nil},
		{"string", "hello", "hello"},
		{"int", 42, 42},
		{"bool", true, true},
		{"string pointer", sPtr, "hello"},
		{"double string pointer", sPtrPtr, "hello"},
		{"int pointer", iPtr, 42},
		{"nil pointer", nilPtr, nil},
		{"slice", []int{1, 2, 3}, []int{1, 2, 3}},
		{"map", map[string]int{"a": 1}, map[string]int{"a": 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typutil.Flatten(tt.a)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten(%v) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}

// This test checks that Flatten works correctly with interfaces.
func TestFlattenWithInterfaces(t *testing.T) {
	s := "hello"
	sPtr := &s
	var i interface{} = sPtr       // interface containing *string
	var iPtr interface{} = &i      // interface containing interface containing *string
	var nilIface interface{} = nil // nil interface

	tests := []struct {
		name string
		a    interface{}
		want interface{}
	}{
		{"interface with string pointer", i, "hello"},
		{"pointer to interface with string pointer", iPtr, "hello"},
		{"nil interface", nilIface, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typutil.Flatten(tt.a)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten(%v) = %v, want %v", tt.a, got, tt.want)
			}
		})
	}
}
