package typutil

import "errors"

var (
	ErrNilPointerRead = errors.New("attempt to read from a nil pointer")
	ErrEmptyValue     = errors.New("validator: value must not be empty")
)
