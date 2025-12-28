[![GoDoc](https://godoc.org/github.com/KarpelesLab/typutil?status.svg)](https://godoc.org/github.com/KarpelesLab/typutil)
[![Coverage Status](https://coveralls.io/repos/github/KarpelesLab/typutil/badge.svg?branch=master)](https://coveralls.io/github/KarpelesLab/typutil?branch=master)

# typutil - Go Type Conversion & Validation Utilities

A powerful Go library for flexible type conversion, validation, and dynamic function calling. Particularly useful when working with JSON, APIs, or any scenario requiring robust type handling.

## Installation

```bash
go get github.com/KarpelesLab/typutil
```

## Features

- **Type Conversion**: Intelligent conversion between types with `Assign` and `As`
- **Validation**: Struct field validation with custom validators
- **Dynamic Functions**: Transform any function into a callable with automatic argument conversion
- **Map to Struct**: Convert `map[string]any` to structs with JSON tag support
- **Flexible Conversion**: Handle strings, numbers, booleans, slices, maps, and custom types

## Quick Start

### Type Conversion with `As`

Convert any value to a target type:

```go
// Convert string to int
result, err := typutil.As[int]("42")
// result = 42

// Convert map to struct
type User struct {
    Name string
    Age  int
}

m := map[string]any{"Name": "Alice", "Age": 30}
user, err := typutil.As[User](m)
// user = User{Name: "Alice", Age: 30}
```

### Type Assignment with `Assign`

Assign values with automatic conversion:

```go
var age int
err := typutil.Assign(&age, "42")
// age = 42

var user User
m := map[string]any{"Name": "Bob", "Age": "25"}
err := typutil.Assign(&user, m)
// user = User{Name: "Bob", Age: 25}
```

## Type Conversion

### `As[T](value)` - Generic Type Conversion

The `As` function provides type-safe conversion from any value to a target type `T`.

```go
// Basic types
i, _ := typutil.As[int]("123")           // 123
f, _ := typutil.As[float64]("3.14")      // 3.14
s, _ := typutil.As[string](42)           // "42"
b, _ := typutil.As[bool](1)              // true

// Struct conversion
type Person struct {
    Name string
    Age  int
}

type Employee struct {
    Name string
    Age  string
}

p := Person{Name: "Alice", Age: 30}
e, _ := typutil.As[Employee](p)
// e = Employee{Name: "Alice", Age: "30"}
```

### `Assign(dst, src)` - Pointer-Based Assignment

The `Assign` function assigns a value to a pointer with automatic conversion.

```go
var result int
err := typutil.Assign(&result, "42")

var user User
err := typutil.Assign(&user, map[string]any{
    "Name": "Charlie",
    "Age":  35,
})
```

### Map to Struct Conversion

Convert maps to structs with support for JSON tags:

```go
type Config struct {
    Host     string
    Port     int
    Timeout  float64
    Username string `json:"user"`
    Password string `json:"pass"`
}

m := map[string]any{
    "Host":    "localhost",
    "Port":    "8080",      // Converts string to int
    "Timeout": 30.5,
    "user":    "admin",     // Matches via json tag
    "pass":    "secret",
    "Extra":   "ignored",   // Unknown fields are ignored
}

config, err := typutil.As[Config](m)
```

### Supported Conversions

- **Primitives**: String, Int, Float, Bool, Byte slices
- **Pointers**: Automatic wrapping/unwrapping
- **Slices**: Element-wise conversion
- **Maps**: Key/value conversion
- **Structs**: Field-by-field conversion with JSON tag support
- **Custom Types**: Via `AssignableTo` and `valueScanner` interfaces

## Validators

Validators allow you to enforce constraints on struct fields during assignment.

### Using Built-in Validators

```go
type User struct {
    Email    string `validator:"not_empty"`
    Password string `validator:"minlength=8,maxlength=64"`
    Color    string `validator:"hex6color"`
    IP       string `validator:"ip_address"`
}

m := map[string]any{
    "Email":    "user@example.com",
    "Password": "secret123",
    "Color":    "#FF5733",
    "IP":       "192.168.1.1",
}

user, err := typutil.As[User](m)
// Validation runs automatically during conversion
```

### Built-in Validators

- `not_empty` - Ensures the value is not empty
- `minlength=N` - Minimum string length
- `maxlength=N` - Maximum string length
- `ip_address` - Valid IP address
- `hex6color` - Valid 6-character hex color (e.g., #FF5733)
- `hex64` - Valid 64-character hex string

### Creating Custom Validators

Register validators with `SetValidator`:

```go
func init() {
    // Simple validator
    typutil.SetValidator("required", func(s string) error {
        if s == "" {
            return errors.New("value is required")
        }
        return nil
    })

    // Validator with arguments
    typutil.SetValidatorArgs("min", func(i int, minVal int) error {
        if i < minVal {
            return fmt.Errorf("value must be at least %d", minVal)
        }
        return nil
    })
}

type Product struct {
    Name  string `validator:"required"`
    Price int    `validator:"min=0"`
}
```

### Multiple Validators

Apply multiple validators to a single field:

```go
type Account struct {
    Username string `validator:"not_empty,minlength=3,maxlength=20"`
    Email    string `validator:"not_empty,email"`
}
```

## Dynamic Function Calling

Transform any function into a generic callable with automatic argument conversion.

### Basic Usage

```go
func Add(a, b int) int {
    return a + b
}

f := typutil.Func(Add)
res, err := typutil.Call[int](f, ctx, 1, "2")
// res = 3 (string "2" converted to int)
```

### Default Arguments

```go
func Add(a, b int) int {
    return a + b
}

f := typutil.Func(Add).WithDefaults(typutil.Required, 42)
res, err := typutil.Call[int](f, ctx, 58)
// res = 100 (second argument defaults to 42)
```

### Context Support

Functions can accept `context.Context` as the first parameter:

```go
func ProcessData(ctx context.Context, data string) (string, error) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return "", ctx.Err()
    default:
        return strings.ToUpper(data), nil
    }
}

f := typutil.Func(ProcessData)
result, err := typutil.Call[string](f, ctx, "hello")
// result = "HELLO"
```

## Advanced Features

### Type Conversion Helpers

```go
// AsString - Convert any type to string
str, ok := typutil.AsString(42)          // "42", true
str, ok := typutil.AsString([]byte{65})  // "A", true

// AsInt - Convert to int64
num, ok := typutil.AsInt("42")           // 42, true
num, ok := typutil.AsInt(3.14)           // 3, true (rounds)

// AsFloat - Convert to float64
f, ok := typutil.AsFloat("3.14")         // 3.14, true

// AsBool - Convert to bool
b := typutil.AsBool("yes")               // true
b := typutil.AsBool(0)                   // false
b := typutil.AsBool("non-empty")         // true
```

### Struct to Map Conversion

```go
type Person struct {
    Name string
    Age  int
}

p := Person{Name: "Alice", Age: 30}

var m map[string]any
err := typutil.Assign(&m, p)
// m = map[string]any{"Name": "Alice", "Age": 30}
```

## Use Cases

### Working with JSON

```go
var rawData map[string]any
json.Unmarshal(jsonBytes, &rawData)

type Config struct {
    Host string
    Port int `json:"port"`
}

config, err := typutil.As[Config](rawData)
```

### API Response Handling

```go
type APIResponse struct {
    Status  string
    Code    int
    Message string
}

responseMap := map[string]any{
    "Status":  "success",
    "Code":    "200",  // String from JSON
    "Message": "OK",
}

response, err := typutil.As[APIResponse](responseMap)
```

### Form Data Processing

```go
type UserForm struct {
    Username string `validator:"not_empty,minlength=3"`
    Email    string `validator:"not_empty"`
    Age      int    `validator:"min=18"`
}

formData := map[string]any{
    "Username": r.FormValue("username"),
    "Email":    r.FormValue("email"),
    "Age":      r.FormValue("age"),  // String from form
}

user, err := typutil.As[UserForm](formData)
if err != nil {
    // Validation or conversion failed
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

## Error Handling

```go
user, err := typutil.As[User](data)
if err != nil {
    if errors.Is(err, typutil.ErrAssignImpossible) {
        // Type conversion not possible
    }
    if errors.Is(err, typutil.ErrInvalidSource) {
        // Source value is invalid (nil)
    }
    // Handle validation errors
}
```

## Performance Considerations

- Type conversion functions are cached for performance
- Validators are registered once at initialization
- Reflection is used internally but optimized with caching
- Zero allocations for simple type conversions

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

See LICENSE file for details.
