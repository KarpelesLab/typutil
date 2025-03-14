package typutil

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// requiredArg is a special type used to denote required arguments when specifying defaults.
//
// This type is used with the Required constant to mark parameters that must be provided
// by the caller when using WithDefaults. It's a sentinel value that prevents default values
// from being used for critical parameters.
//
// See WithDefaults method and Required constant for more information on how to use this.
type requiredArg int

// funcOption is a function that configures a Callable instance.
//
// This follows the functional options pattern for configuring structs, which allows
// for a clean and flexible API for specifying optional configuration parameters.
// Each option is a function that takes a *Callable and modifies it.
//
// Currently implemented options include:
// - StrictArgs: Enforces strict type checking for function arguments
//
// Custom options can be implemented by creating functions that match this signature
// and modify the Callable as needed.
type funcOption func(*Callable)

// Required is a sentinel value that marks a parameter as required when using WithDefaults.
// When specified as a default value, it indicates that the parameter must be provided
// by the caller and cannot be defaulted.
const Required requiredArg = 1

// StrictArgs is a functional option for Func that enforces strict type checking of arguments.
// When enabled, arguments passed to the function must match the expected types exactly,
// rather than using the more flexible conversion via the Assign function.
//
// Example:
//
//	// Without StrictArgs, "42" would be converted to int(42)
//	f := Func(myIntFunc)                 // Allows "42" to be passed to an int parameter
//	f2 := Func(myIntFunc, StrictArgs)    // Requires an actual int, rejects "42"
func StrictArgs(c *Callable) {
	c.strict = true
}

// Callable represents a wrapped function that can be called with flexible argument handling.
// It provides several enhancements over standard Go function calls:
//
// 1. Context parameter detection - automatically detects and handles context.Context parameters
// 2. Type conversion - automatically converts arguments to the required parameter types
// 3. Default values - allows specifying default values for parameters
// 4. Variadic support - handles variadic functions (functions with ...T parameters)
// 5. JSON deserialization - can handle input from JSON
//
// This type is not typically constructed directly but through the Func() function.
type Callable struct {
	fn       reflect.Value   // The wrapped function
	cnt      int             // Number of actual args (excluding context)
	ctxPos   int             // Position of context.Context argument, or -1 if none
	arg      []reflect.Type  // Types of the arguments to the method
	def      []reflect.Value // Default values for arguments
	variadic bool            // Whether the function's last argument is variadic (...)
	strict   bool            // Whether to enforce strict type checking
	vartyp   reflect.Type    // Type of the variadic argument (element type of the slice)
}

var (
	// ctxTyp is the reflect.Type representing context.Context interface
	ctxTyp = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// Func wraps a Go function as a Callable object, enabling flexible argument handling.
//
// The wrapped function can have several forms:
// - Standard function: func(a, b, c) result
// - Context-aware function: func(ctx context.Context, a, b, c) result
// - Variadic function: func(a, b, ...c) result
// - Error-returning function: func(a, b) (result, error)
// - Any combination of the above
//
// The returned Callable provides several enhancements:
// - Automatic conversion of argument types using Assign
// - Support for default parameter values (via WithDefaults)
// - Automatic handling of context parameters
// - JSON input parsing from context
// - Variadic function support
//
// Parameters:
//   - method: The function to wrap (can also be a *Callable to apply additional options)
//   - options: Optional configuration options (e.g., StrictArgs)
//
// Returns:
//   - A Callable object that wraps the function
//
// Example:
//
//	func Add(a, b int) int { return a + b }
//
//	// Create a Callable that allows flexible type conversion
//	callable := Func(Add)
//
//	// Call with arguments that will be automatically converted
//	result, err := callable.CallArg(ctx, 5, "10")  // result = 15
//
//	// Or with generic type inference
//	result, err := Call[int](callable, ctx, 5, "10")  // result = 15, strongly typed
func Func(method any, options ...funcOption) *Callable {
	// If a Callable is passed, create a new Callable with the same values and apply the provided options
	if v, ok := method.(*Callable); ok {
		// This allows applying additional options to an existing Callable
		// Examples:
		//   f := Func(add)              // Create a Callable
		//   f2 := Func(f, StrictArgs)   // Apply StrictArgs to f
		//
		// It's also useful for passing a *Callable to a function expecting a func()
		// and using Func() to extract it back into a Callable
		nv := &Callable{}
		*nv = *v // Copy all values from the original Callable

		// Apply all provided options
		for _, opt := range options {
			opt(nv)
		}
		return nv
	}

	// Verify the method is actually a function
	v := reflect.ValueOf(method)
	if v.Kind() != reflect.Func {
		panic("static method not a method")
	}

	// Get the type information for the function
	typ := v.Type()

	// Create a new Callable with default values
	// ctxPos is -1 initially, meaning no context parameter detected yet
	// cnt is set to the total number of input parameters
	res := &Callable{fn: v, ctxPos: -1, cnt: typ.NumIn()}

	ni := res.cnt

	// Iterate through all the input parameters
	for i := 0; i < ni; i += 1 {
		in := typ.In(i)

		// Check if this parameter implements context.Context
		if in.Implements(ctxTyp) {
			// If we already found a context parameter, that's an error
			if res.ctxPos != -1 {
				panic("method taking multiple ctx arguments")
			}

			// Store the position of the context parameter
			res.ctxPos = i

			// Decrement the count of actual (non-context) arguments
			res.cnt -= 1
			continue
		}

		// For non-context parameters, add the type to our arguments list
		res.arg = append(res.arg, in)
	}

	// Handle variadic functions (those with ...T parameters)
	if typ.IsVariadic() {
		res.variadic = true
		ln := len(res.arg)

		// For a variadic function, the last argument is a slice []T
		// We extract the element type (T) and store it
		res.vartyp = res.arg[ln-1].Elem()

		// Remove the variadic parameter from our regular args list
		// It will be handled specially during calls
		res.arg = res.arg[:ln-1]
	}

	// Apply any provided options
	for _, opt := range options {
		opt(res)
	}

	return res
}

// String returns a string representation of the function's signature.
//
// This provides a human-readable view of the function's parameter types,
// including the variadic parameters if present. The context parameter
// is not included in this representation.
//
// The format is similar to a Go function signature:
//
//	func(type1, type2, ...type3)
//
// This is useful for debugging and logging purposes.
func (s *Callable) String() string {
	var args []string
	for _, arg := range s.arg {
		args = append(args, arg.String())
	}
	if s.variadic {
		args = append(args, "..."+s.vartyp.String())
	}
	return "func(" + strings.Join(args, ", ") + ")"
}

// WithDefaults creates a new Callable with default argument values.
//
// Default values are used when a call to CallArg doesn't provide enough arguments.
// This enables creating functions where some parameters are optional.
//
// Parameters:
//   - args: Default values for each parameter. Use typutil.Required for parameters
//     that must be provided by the caller.
//
// Returns:
//   - A new Callable with the specified default values
//
// Panics if:
//   - Not enough default arguments are provided
//   - Too many default arguments are provided for a non-variadic function
//   - A default value cannot be converted to the expected parameter type
//
// Example:
//
//	// Original function
//	func myFunc(a, b, c int) int { return a + b + c }
//
//	// Create a Callable with defaults for parameters
//	f := Func(myFunc).WithDefaults(typutil.Required, typutil.Required, 42)
//
//	// Call with only the required parameters (c uses default value)
//	result, _ := f.CallArg(context.Background(), 10, 20) // equivalent to myFunc(10, 20, 42)
//
//	// You can also provide all parameters, overriding the defaults
//	result, _ := f.CallArg(context.Background(), 10, 20, 30) // equivalent to myFunc(10, 20, 30)
//
// This pattern is especially useful for creating API handlers where some parameters are optional.
func (s *Callable) WithDefaults(args ...any) *Callable {
	// Ensure enough default arguments are provided (one for each parameter)
	if len(args) < s.cnt {
		panic("WithDefaults requires at least the same number of arguments as the function")
	}

	// For non-variadic functions, ensure we don't provide too many defaults
	if len(args) > s.cnt && !s.variadic {
		panic(ErrTooManyArgs)
	}

	// Create an array to hold the default values
	def := make([]reflect.Value, len(args))

	// Process each default argument
	for argN, arg := range args {
		// If this argument is marked as Required, skip converting it
		// It will remain as an invalid value in the defaults array
		if arg == Required {
			// keep this as an invalid value that will be checked later
			continue
		}

		// Create a new reflect.Value to hold the default value
		var argV reflect.Value
		if argN >= len(s.arg) {
			// This is a variadic argument, create a new instance of the variadic type
			argV = reflect.New(s.vartyp)
		} else {
			// This is a regular argument, create a new instance of its type
			argV = reflect.New(s.arg[argN])
		}

		// Either directly set the value (if strict mode) or use Assign for type conversion
		if s.strict {
			// In strict mode, directly set the value without type conversion
			argV.Set(reflect.ValueOf(arg))
		} else {
			// In non-strict mode, use Assign to convert the type if needed
			err := AssignReflect(argV, reflect.ValueOf(arg))
			if err != nil {
				// If the conversion fails, panic with the error
				panic(err)
			}
		}

		// Store the converted value (not the pointer) in the defaults array
		def[argN] = argV.Elem()
	}

	// Create a new Callable with the same attributes as the original
	res := &Callable{}
	*res = *s

	// Update it with the new default values
	res.def = def

	return res
}

// Call invokes the function without explicit arguments, looking for input from context if needed.
//
// This method is particularly useful when working with API handlers or middleware where
// the arguments need to be parsed from a request body or context.
//
// If the function requires arguments (has parameters), Call will:
// 1. Look for a value stored under "input_json" in the context
// 2. If found, parse it as JSON and pass it as arguments to the function
// 3. For multiple parameters, the JSON should be an array with one element per parameter
// 4. For a single parameter, the JSON can be any value that's convertible to that parameter type
//
// Parameters:
//   - ctx: The context to pass to the function (and to extract arguments from if needed)
//
// Returns:
//   - The value returned by the function (or nil if it returns nothing)
//   - An error if the function call fails or returns an error
//
// Example:
//
//	// Function that adds two numbers
//	func add(a, b int) int { return a + b }
//	callable := Func(add)
//
//	// Store JSON input in context
//	ctx := context.WithValue(context.Background(), "input_json", json.RawMessage(`[5, 10]`))
//
//	// Call the function using arguments from context
//	result, _ := callable.Call(ctx) // result = 15
func (s *Callable) Call(ctx context.Context) (any, error) {
	// If the function expects arguments (non-context parameters)
	if s.cnt > 0 {
		// Try to get JSON data from the context under the "input_json" key
		input, ok := ctx.Value("input_json").(json.RawMessage)
		if ok {
			// Found JSON input in the context, use it for the function arguments

			if s.cnt > 1 {
				// For functions with multiple parameters, the JSON should be an array
				// We parse it into a []RawJsonMessage to handle each element separately later
				var args []RawJsonMessage
				err := json.Unmarshal(input, &args)
				if err != nil {
					// Return an error if the JSON is invalid or not an array
					return nil, err
				}

				// Convert each JSON element to an interface{} value
				anyArgs := make([]any, len(args))
				for n, v := range args {
					anyArgs[n] = v
				}

				// Call the function with the extracted arguments
				return s.CallArg(ctx, anyArgs...)
			}

			// For a single-parameter function, pass the entire JSON as that parameter
			return s.CallArg(ctx, RawJsonMessage(input))
		}
	}

	// If there are no arguments or no JSON input was found,
	// call the function without arguments (will use defaults if defined)
	return s.CallArg(ctx)
}

// CallArg calls the function with explicitly provided arguments.
//
// This is the core method for invoking a wrapped function. It handles:
// - Converting arguments to the correct parameter types
// - Adding the context parameter if needed
// - Supplying default values for missing arguments (if WithDefaults was used)
// - Handling variadic arguments
// - Processing return values including errors
//
// Parameters:
//   - ctx: The context to pass to the function (if it accepts one)
//   - arg: The arguments to pass to the function
//
// Returns:
//   - The value returned by the function (or nil if it returns nothing)
//   - An error if:
//   - Not enough arguments are provided (and no defaults are available)
//   - Too many arguments are provided to a non-variadic function
//   - An argument can't be converted to the expected parameter type
//   - The function call itself returns an error
//
// Example:
//
//	// Function that adds two numbers
//	func add(a, b int) int { return a + b }
//	callable := Func(add)
//
//	// Call with properly typed arguments
//	result, _ := callable.CallArg(ctx, 5, 10) // result = 15
//
//	// Call with arguments that need conversion
//	result, _ := callable.CallArg(ctx, "5", 10.0) // result = 15
//
//	// When using defaults, you can omit some arguments
//	callable = callable.WithDefaults(typutil.Required, 10)
//	result, _ := callable.CallArg(ctx, 5) // result = 15
func (s *Callable) CallArg(ctx context.Context, arg ...any) (any, error) {
	// Special case: function takes no arguments (other than possibly context)
	if s.cnt == 0 {
		// Create slice to hold only the context argument (if needed)
		var args []reflect.Value

		// If the function takes a context parameter, add it
		if s.ctxPos == 0 {
			args = append(args, reflect.ValueOf(ctx))
		}

		// Call the function and parse the result
		return s.parseResult(s.fn.Call(args))
	}

	// Check if we have enough arguments
	if len(arg) < s.cnt && s.def == nil {
		// Not enough arguments and no defaults available
		return nil, ErrMissingArgs
	}

	// Check if we have too many arguments for a non-variadic function
	if len(arg) > s.cnt && !s.variadic {
		return nil, ErrTooManyArgs
	}

	// Prepare the arguments slice and tracking variables
	var args []reflect.Value // Will hold all arguments to pass to the function
	var ctxPos int           // Position of context parameter in the function
	var ctxCnt int           // Number of context parameters (0 or 1)

	// Handle functions that take a context parameter
	if s.ctxPos != -1 {
		// The function takes a context, allocate space for it plus all arguments
		args = make([]reflect.Value, len(arg)+1)

		// Add the context parameter at the correct position
		args[s.ctxPos] = reflect.ValueOf(ctx)

		// Set tracking variables
		ctxPos = s.ctxPos
		ctxCnt = 1
	} else {
		// No context parameter, just allocate space for the regular arguments
		args = make([]reflect.Value, len(arg))

		// Set ctxPos to a value that ensures correct index calculations below
		ctxPos = len(arg) + 1
	}

	// Process each provided argument
	for argN, v := range arg {
		// Create a new value to hold the converted argument
		var argV reflect.Value

		if argN >= len(s.arg) {
			// This is a variadic argument, create a value of the variadic element type
			argV = reflect.New(s.vartyp)
		} else {
			// This is a regular argument, create a value of its type
			argV = reflect.New(s.arg[argN])
		}

		// Convert the argument to the target type
		if s.strict {
			// In strict mode, only allow directly assignable types
			subv := reflect.ValueOf(v)
			if !subv.Type().AssignableTo(argV.Type()) {
				return nil, ErrAssignImpossible
			}
			argV.Set(reflect.ValueOf(v))
		} else {
			// In normal mode, use Assign to convert between types
			err := AssignReflect(argV, reflect.ValueOf(v))
			if err != nil {
				return nil, err
			}
		}

		// Store the argument in the args slice, accounting for the context position
		if argN >= ctxPos {
			// If this argument comes after the context parameter,
			// shift its position by 1 to account for the context
			args[argN+1] = argV.Elem()
		} else {
			// If this argument comes before the context parameter,
			// its position in args is the same as in the original arguments
			args[argN] = argV.Elem()
		}
	}

	// If we have defaults defined and didn't provide enough arguments,
	// add the default values to complete the call
	if len(args)-ctxCnt < len(s.def) {
		// Get the default values we need to add
		add := s.def[len(args)-ctxCnt:]

		// Check that all required defaults are valid (not marked as Required)
		for _, v := range add {
			if !v.IsValid() {
				return nil, ErrMissingArgs
			}
		}

		// Append the default values to complete the arguments
		args = append(args, add...)
	}

	// Call the function with all arguments and parse the result
	return s.parseResult(s.fn.Call(args))
}

// IsStringArg returns true if the nth argument of the callable is a string, or a type related to string.
//
// This is a utility method for quickly checking if an argument is string-based,
// which is useful when implementing custom argument handling logic.
//
// Parameters:
//   - n: The zero-based index of the argument to check
//
// Returns:
//   - true if the argument is a string type
//   - false otherwise (including when the index is out of bounds)
//
// Example:
//
//	callable := Func(func(name string, age int) {})
//	isString := callable.IsStringArg(0) // true
//	isString = callable.IsStringArg(1)  // false
func (s *Callable) IsStringArg(n int) bool {
	return s.ArgKind(n) == reflect.String
}

// ArgKind returns the reflect.Kind for the nth argument of the callable.
//
// This method provides type information about function parameters without
// needing to use reflection directly. It's useful for implementing custom
// argument handling or validation logic.
//
// Parameters:
//   - n: The zero-based index of the argument to check
//
// Returns:
//   - The reflect.Kind of the argument's type
//   - reflect.Invalid if n is out of bounds
//
// Example:
//
//	callable := Func(func(name string, age int, valid bool) {})
//	kind := callable.ArgKind(0) // reflect.String
//	kind = callable.ArgKind(1)  // reflect.Int
//	kind = callable.ArgKind(2)  // reflect.Bool
//	kind = callable.ArgKind(3)  // reflect.Invalid (out of bounds)
func (s *Callable) ArgKind(n int) reflect.Kind {
	if n >= len(s.arg) {
		return reflect.Invalid
	}
	return s.arg[n].Kind()
}

// errTyp is the reflect.Type representing the error interface
var errTyp = reflect.TypeOf((*error)(nil)).Elem()

// parseResult processes the return values from a function call.
//
// This internal method interprets the reflect.Value slice returned by calling a function
// through reflection. It handles two common return patterns in Go:
// 1. A single return value (just a result)
// 2. A result and an error (value, error)
//
// For functions that return multiple values, it identifies which one is an error
// (by checking if it implements the error interface) and which is the result value.
//
// Parameters:
//   - res: A slice of reflect.Value objects representing the function's return values
//
// Returns:
//   - output: The non-error return value (or nil if none)
//   - err: The error return value (or nil if no error)
func (s *Callable) parseResult(res []reflect.Value) (output any, err error) {
	// For each value in res, try to find which one is an error and which one is a result
	for _, v := range res {
		if v.Type().Implements(errTyp) {
			err, _ = v.Interface().(error)
			continue
		}
		output = v.Interface()
	}
	return
}

// Call invokes a Callable and returns a strongly typed result.
//
// This generic function provides type safety for calling wrapped functions.
// It automatically converts the return value to the requested type T,
// which eliminates the need for type assertions in the calling code.
//
// Parameters:
//   - s: The Callable to invoke
//   - ctx: The context to pass to the function
//   - arg: The arguments to pass to the function
//
// Returns:
//   - A value of type T (the function's return value)
//   - An error if:
//   - The function call failed (see CallArg for possible errors)
//   - The function's return value could not be converted to type T
//
// Example:
//
//	func add(a, b int) int { return a + b }
//	callable := Func(add)
//
//	// Get a strongly typed result
//	result, err := Call[int](callable, ctx, 5, 10) // result is an int, not an interface{}
//
//	// This will fail with ErrDifferentType
//	result, err := Call[string](callable, ctx, 5, 10) // err indicates type mismatch
//
// Generic functions like this are particularly useful in code that needs to maintain
// type safety while still supporting flexible function calling patterns.
func Call[T any](s *Callable, ctx context.Context, arg ...any) (T, error) {
	res, err := s.CallArg(ctx, arg...)
	if v, ok := res.(T); ok {
		return v, err
	} else if err == nil {
		err = fmt.Errorf("%w: %T", ErrDifferentType, res)
	}
	return reflect.New(reflect.TypeFor[T]()).Elem().Interface().(T), err
}
