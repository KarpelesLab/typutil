package typutil

import "errors"

var (
	ErrAssignDestNotPointer      = errors.New("assign destination must be a pointer")
	ErrAssignImpossible          = errors.New("the requested assign is not possible")
	ErrNilPointerRead            = errors.New("attempt to read from a nil pointer")
	ErrEmptyValue                = errors.New("validator: value must not be empty")
	ErrDestinationNotAddressable = errors.New("assign: destination cannot be addressed")
	ErrStructPtrRequired         = errors.New("parameter must be a pointer to a struct")
	ErrBadOffset                 = errors.New("bad offset type")
	ErrMissingArgs               = errors.New("missing arguments")
	errTooManyArgs               = errors.New("too many arguments")
)
