package typutil

import (
	"errors"
	"fmt"
	"net/netip"
	"reflect"
)

func init() {
	SetValidator("not_empty", validateNotEmpty)
	SetValidatorArgs("minlength", validateMinLength)
	SetValidatorArgs("maxlength", validateMaxLength)
	SetValidator("ip_address", validateIpAddr)
	SetValidator("hex6color", validateHex6Color)
	SetValidator("hex64", validateHex64)
}

func validateNotEmpty(v any) error {
	switch t := v.(type) {
	case string:
		if len(t) == 0 {
			return ErrEmptyValue
		}
		return nil
	default:
		s := reflect.ValueOf(v)
		if s.Kind() == reflect.Pointer {
			return validateNotEmpty(s.Elem().Interface())
		}
		// AsBool will return true if value is non zero, non empty
		if AsBool(v) {
			return nil
		}
		return ErrEmptyValue
	}
}

func validateMinLength(v string, ln int) error {
	if len(v) < ln {
		return fmt.Errorf("string must be at least %d characters", ln)
	}
	return nil
}

func validateMaxLength(v string, ln int) error {
	if len(v) > ln {
		return fmt.Errorf("string must be at most %d characters", ln)
	}
	return nil
}

func validateIpAddr(ip string) error {
	if ip == "" {
		return nil
	}
	_, err := netip.ParseAddr(ip)
	return err
}

func validateHex6Color(color *string) error {
	// 6 digits hex color, for example 336699
	// if a # is found in first position it will be trimmed. If value is empty no error will be returned (chain with not_empty to check emptyness too)
	if *color == "" {
		return nil
	}
	if (*color)[0] == '#' {
		*color = (*color)[1:]
	}
	if len(*color) != 6 {
		return errors.New("expecting 6 digits hex color")
	}

	for _, n := range *color {
		if (n < '0' || n > '9') && (n < 'a' || n > 'f') && (n < 'A' || n > 'F') {
			return errors.New("invalid hex char in color")
		}
	}

	return nil
}

// validateHex64 ensures a given string is exactly 64 hexadecimal characters, for example a value of a 256bits hash such as sha256
func validateHex64(hex string) error {
	if hex == "" {
		return nil
	}
	if len(hex) != 64 {
		return errors.New("expected 64 hex chars")
	}
	for _, n := range hex {
		if (n < '0' || n > '9') && (n < 'a' || n > 'f') && (n < 'A' || n > 'F') {
			return errors.New("invalid hex char")
		}
	}

	return nil
}
