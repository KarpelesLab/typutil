package typutil

import (
	"github.com/KarpelesLab/pjson"
)

// RawJsonMessage is similar to json.RawMessage, but also implements additional functionality.
//
// This type represents a raw JSON message as a byte slice. It provides:
// 1. Standard JSON marshaling/unmarshaling (like json.RawMessage)
// 2. The ability to assign its value to another variable via pjson.Unmarshal
//
// RawJsonMessage is particularly useful when working with the Callable.Call method,
// which can extract JSON data from context and use it as function arguments.
type RawJsonMessage []byte

// MarshalJSON implements the json.Marshaler interface.
//
// This method simply returns the raw JSON bytes, allowing the JSON data
// to be embedded directly within another JSON structure.
func (m RawJsonMessage) MarshalJSON() ([]byte, error) {
	return []byte(m), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// This method stores the raw JSON bytes without parsing them, which allows
// deferring the actual JSON parsing until it's needed.
func (m *RawJsonMessage) UnmarshalJSON(data []byte) error {
	*m = data
	return nil
}

// AssignTo unmarshals the raw JSON message into the provided value.
//
// This method uses pjson.Unmarshal (an enhanced JSON unmarshaler) to parse
// the raw JSON data and assign it to the target value. It's useful for
// converting JSON data to Go types, particularly when working with function
// arguments in the Callable.Call method.
//
// Parameters:
//   - v: The target value to unmarshal the JSON into (passed by reference)
//
// Returns:
//   - An error if the JSON parsing fails
func (m RawJsonMessage) AssignTo(v any) error {
	return pjson.Unmarshal([]byte(m), v)
}
