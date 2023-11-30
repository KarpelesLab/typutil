[![GoDoc](https://godoc.org/github.com/KarpelesLab/typutil?status.svg)](https://godoc.org/github.com/KarpelesLab/typutil)

# type conversion utils

This is useful when dealing with json parsed types for example.

# Assign

Assign is a tool that allows assigning any value to any other value and let the library handle the conversion in a somewhat intelligent way.

For example a `map[string]any` can be assigned to a struct (`json` tags will be taken into account for variable names) and values will be converted.
