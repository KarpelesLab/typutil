package typutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/url"
	"strconv"
)

// some helper functions related to numbers
func AsBool(v any) bool {
	v = BaseType(v)
	switch r := v.(type) {
	case bool:
		return r
	case int:
		return r != 0
	case int64:
		return r != 0
	case uint64:
		return r != 0
	case float64:
		return r != 0
	case *bytes.Buffer:
		if r.Len() > 1 {
			return true
		}
		if r.Len() == 0 || r.String() == "0" {
			return false
		}
		return true
	case string:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r == "0" {
			return false
		}
		return true
	case []byte:
		if len(r) > 1 {
			return true
		}
		if len(r) == 0 || r[0] == '0' {
			return false
		}
		return true
	case map[string]any:
		if len(r) > 0 {
			return true
		}
		return false
	case []any:
		if len(r) > 0 {
			return true
		}
		return false
	case url.Values:
		return len(r) > 0
	default:
		return false
	}
}

func AsInt(v any) (int64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		if n&(1<<63) != 0 {
			return int64(n), false
		}
		return int64(n), true
	case uint:
		return int64(n), true
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case float32:
		x := math.Round(float64(n))
		y := int64(x)
		return y, float64(y) == x
	case float64:
		x := math.Round(n)
		y := int64(x)
		return y, float64(y) == x
	case string:
		res, err := strconv.ParseInt(n, 0, 64)
		return res, err == nil
	case []byte:
		res, err := strconv.ParseInt(string(n), 0, 64)
		return res, err == nil
	case *bytes.Buffer:
		return AsInt(string(n.Bytes()))
	case json.Number:
		return AsInt(string(n))
	case nil:
		return 0, true
	default:
		log.Printf("[number] failed to parse type %T", n)
	}

	return 0, false
}

func AsUint(v any) (uint64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return uint64(n), n > 0
	case int16:
		return uint64(n), n > 0
	case int32:
		return uint64(n), n > 0
	case int64:
		return uint64(n), n > 0
	case int:
		return uint64(n), n > 0
	case uint8:
		return uint64(n), true
	case uint16:
		return uint64(n), true
	case uint32:
		return uint64(n), true
	case uint64:
		return n, true
	case uint:
		return uint64(n), true
	case float32:
		if n < 0 {
			return 0, false
		}
		x := math.Round(float64(n))
		y := uint64(x)
		return y, float64(y) == x
	case float64:
		if n < 0 {
			return 0, false
		}
		x := math.Round(n)
		y := uint64(x)
		return y, float64(y) == x
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		res, err := strconv.ParseUint(n, 0, 64)
		return res, err == nil
	case json.Number:
		return AsUint(string(n))
	case nil:
		return 0, true
	}

	return 0, false
}

func AsFloat(v any) (float64, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case int:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uintptr:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		res, err := strconv.ParseFloat(n, 64)
		return res, err == nil
	case nil:
		return 0, true
	}

	res, ok := AsInt(v)
	return float64(res), ok
}

func AsNumber(v any) (any, bool) {
	v = BaseType(v)
	switch n := v.(type) {
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return uint64(n), true
	case uintptr:
		return uint64(n), true
	case uint:
		return int64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case nil:
		return 0, true
	case bool:
		return n, true
	case string:
		if res, err := strconv.ParseInt(n, 0, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseUint(n, 0, 64); err == nil {
			return res, true
		}
		if res, err := strconv.ParseFloat(n, 64); err == nil {
			return res, true
		}
		return AsBool(n), false
	case *bytes.Buffer:
		if n.Len() > 100 {
			return nil, false
		}
		return AsNumber(n.String())
	}

	return nil, false
}

func AsString(v any) (string, bool) {
	v = BaseType(v)
	switch s := v.(type) {
	case string:
		return s, true
	case []byte:
		return string(s), true
	case *bytes.Buffer:
		return s.String(), true
	case int64:
		return strconv.FormatInt(s, 10), true
	case int:
		return strconv.FormatInt(int64(s), 10), true
	case int32:
		return strconv.FormatInt(int64(s), 10), true
	case int16:
		return strconv.FormatInt(int64(s), 10), true
	case int8:
		return strconv.FormatInt(int64(s), 10), true
	case uint64:
		return strconv.FormatUint(s, 10), true
	case uint:
		return strconv.FormatUint(uint64(s), 10), true
	case uint32:
		return strconv.FormatUint(uint64(s), 10), true
	case uint16:
		return strconv.FormatUint(uint64(s), 10), true
	case uint8:
		return strconv.FormatUint(uint64(s), 10), true
	case bool:
		if s {
			return "1", true
		} else {
			return "0", true
		}
	default:
		return fmt.Sprintf("%v", v), false
	}
}
