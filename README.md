[![GoDoc](https://godoc.org/github.com/KarpelesLab/typutil?status.svg)](https://godoc.org/github.com/KarpelesLab/typutil)

# type conversion utils

This is useful when dealing with json parsed types for example.

# Assign

Assign is a tool that allows assigning any value to any other value and let the library handle the conversion in a somewhat intelligent way.

For example a `map[string]any` can be assigned to a struct (`json` tags will be taken into account for variable names) and values will be converted.

# Func

It is possible to transform any func into a generic callable that can be used in various ways, with its arguments automatically converted to match the required values.

For example:

```go
func Add(a, b int) int {
    return a + b
}

f := typutil.Func(Add)
res, err := typutil.Call[int](f, ctx, 1, "2") // res=3
```
