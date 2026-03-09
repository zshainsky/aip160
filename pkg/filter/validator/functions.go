package validator

import "reflect"

// FunctionDef defines a supported function's validation rules.
type FunctionDef struct {
	Name            string         // Function name
	ArgCount        int            // Expected number of arguments
	ValidFieldKinds []reflect.Kind // Valid field types for arguments
}

// SupportedFunctions defines the registry of functions that can be used in filters.
// Add more functions here as needed for your use case.
var SupportedFunctions = map[string]FunctionDef{
	"timestamp": {
		Name:     "timestamp",
		ArgCount: 1,
		ValidFieldKinds: []reflect.Kind{
			reflect.Int64, // Unix timestamp (most common)
			reflect.Int32, // Also accept 32-bit timestamps
		},
	},
	"size": {
		Name:     "size",
		ArgCount: 1,
		ValidFieldKinds: []reflect.Kind{
			reflect.Slice,  // Size of arrays/slices
			reflect.Array,  // Size of fixed arrays
			reflect.String, // Length of strings
		},
	},
}
