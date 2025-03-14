package typutil

// This file defines generic type constraints without requiring external dependencies.
// Similar to golang.org/x/exp/constraints but defined locally for simplicity.

// Signed represents all signed integer types in Go.
// It includes both built-in types and user-defined types that have a signed integer as their underlying type.
// The tilde (~) symbol indicates that this interface matches both the specific type and any type derived from it.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned represents all unsigned integer types in Go.
// It includes both built-in types and user-defined types that have an unsigned integer as their underlying type.
// The tilde (~) symbol indicates that this interface matches both the specific type and any type derived from it.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}
