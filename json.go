package typutil

import (
	"github.com/KarpelesLab/pjson"
)

// RawJsonMessage is similar to json.RawMessage, but also implements
type RawJsonMessage []byte

func (m RawJsonMessage) MarshalJSON() ([]byte, error) {
	return []byte(m), nil
}

func (m *RawJsonMessage) UnmarshalJSON(data []byte) error {
	*m = data
	return nil
}

// AssignTo will unmarshal the raw json message into v
func (m RawJsonMessage) AssignTo(v any) error {
	return pjson.Unmarshal([]byte(m), v)
}
