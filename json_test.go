package typutil_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/KarpelesLab/typutil"
)

func TestRawJsonMessageMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		message  typutil.RawJsonMessage
		expected []byte
	}{
		{"empty", typutil.RawJsonMessage{}, []byte{}},
		{"null", typutil.RawJsonMessage(`null`), []byte(`null`)},
		{"string", typutil.RawJsonMessage(`"test"`), []byte(`"test"`)},
		{"number", typutil.RawJsonMessage(`42`), []byte(`42`)},
		{"object", typutil.RawJsonMessage(`{"key":"value"}`), []byte(`{"key":"value"}`)},
		{"array", typutil.RawJsonMessage(`[1,2,3]`), []byte(`[1,2,3]`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.message.MarshalJSON()
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Expected %v, got %v", string(tt.expected), string(result))
			}
		})
	}
}

func TestRawJsonMessageUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected typutil.RawJsonMessage
	}{
		{"empty", []byte{}, typutil.RawJsonMessage{}},
		{"null", []byte(`null`), typutil.RawJsonMessage(`null`)},
		{"string", []byte(`"test"`), typutil.RawJsonMessage(`"test"`)},
		{"number", []byte(`42`), typutil.RawJsonMessage(`42`)},
		{"object", []byte(`{"key":"value"}`), typutil.RawJsonMessage(`{"key":"value"}`)},
		{"array", []byte(`[1,2,3]`), typutil.RawJsonMessage(`[1,2,3]`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result typutil.RawJsonMessage
			err := result.UnmarshalJSON(tt.data)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Expected %v, got %v", string(tt.expected), string(result))
			}
		})
	}
}

func TestRawJsonMessageAssignTo(t *testing.T) {
	tests := []struct {
		name     string
		message  typutil.RawJsonMessage
		target   interface{}
		expected interface{}
	}{
		{"string", typutil.RawJsonMessage(`"test"`), new(string), "test"},
		{"number", typutil.RawJsonMessage(`42`), new(int), 42},
		{"bool", typutil.RawJsonMessage(`true`), new(bool), true},
		{"array", typutil.RawJsonMessage(`[1,2,3]`), new([]int), []int{1, 2, 3}},
		{"object", typutil.RawJsonMessage(`{"name":"test","value":42}`), new(map[string]interface{}), map[string]interface{}{"name": "test", "value": float64(42)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.message.AssignTo(tt.target)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Need to get the value that the pointer points to
			targetValue := reflect.ValueOf(tt.target).Elem().Interface()
			if !reflect.DeepEqual(targetValue, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, targetValue)
			}
		})
	}
}

func TestRawJsonMessageWithStruct(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	jsonData := typutil.RawJsonMessage(`{"name":"test","value":42}`)
	expected := TestStruct{Name: "test", Value: 42}

	var actual TestStruct
	err := jsonData.AssignTo(&actual)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestRawJsonMessageInvalidJSON(t *testing.T) {
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	// Invalid JSON - missing closing brace
	jsonData := typutil.RawJsonMessage(`{"name":"test","value":42`)

	var target TestStruct
	err := jsonData.AssignTo(&target)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
