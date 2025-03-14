// Package typutil provides utilities for type conversion and function argument handling in Go.
package typutil

import "errors"

// Error constants for various operations in the typutil package.
// These errors are returned from functions like Assign, Func, and validation methods.
var (
	// Assignment-related errors
	ErrAssignDestNotPointer      = errors.New("assign destination must be a pointer")
	ErrAssignImpossible          = errors.New("the requested assign is not possible")
	ErrNilPointerRead            = errors.New("attempt to read from a nil pointer")
	ErrDestinationNotAddressable = errors.New("assign: destination cannot be addressed")
	ErrInvalidSource             = errors.New("assign source is not valid")

	// Validation-related errors
	ErrEmptyValue        = errors.New("validator: value must not be empty")
	ErrStructPtrRequired = errors.New("parameter must be a pointer to a struct")

	// Function calling errors
	ErrMissingArgs   = errors.New("missing arguments")
	ErrTooManyArgs   = errors.New("too many arguments")
	ErrDifferentType = errors.New("wrong type in function call")

	// Offset-related errors
	ErrBadOffset = errors.New("bad offset type")
)
