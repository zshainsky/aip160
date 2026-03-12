package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zshainsky/aip160/pkg/filter/ast"
)

// Validator validates filter ASTs against a Go struct schema using reflection.
// It checks field existence, type compatibility, array operators, and function calls.
type Validator struct {
	structType reflect.Type
	tagName    string // Tag name to use for field matching (e.g., "json", "protobuf"), empty = use field name
}

// ValidatorOption is a function that configures a Validator.
type ValidatorOption func(*Validator)

// WithJSONTags configures the validator to match field names against json struct tags.
// For example, a field `Name string `json:"name,omitempty"` would match the filter field "name".
func WithJSONTags() ValidatorOption {
	// TODO: Return a function that sets v.tagName = "json"
	return nil
}

// WithProtobufTags configures the validator to match field names against protobuf struct tags.
// It extracts the "name=" value from protobuf tags.
// For example, `Name string `protobuf:"bytes,2,opt,name=name,proto3"` would match "name".
func WithProtobufTags() ValidatorOption {
	// TODO: Return a function that sets v.tagName = "protobuf"
	return nil
}

// WithTag configures the validator to match field names against a custom struct tag.
// The tag value will be used as-is (comma-separated values will use the first part).
func WithTag(tagName string) ValidatorOption {
	// TODO: Return a function that sets v.tagName to the provided tagName
	return nil
}

// NewValidator creates a new validator for the given struct type.
// By default, it matches against PascalCase field names.
// Optionally provide one tag option to configure tag-based matching.
//
// Example: NewValidator(reflect.TypeOf(User{}), WithJSONTags())
func NewValidator(structType reflect.Type, opts ...ValidatorOption) *Validator {
	// TODO: Create a new Validator with the given structType
	// TODO: If opts are provided, apply the first option
	// Hint: Only apply the first option if len(opts) > 0
	return nil
}

// Validate validates the entire AST and returns a list of all validation errors.
// Returns an empty slice if the AST is valid.
func (v *Validator) Validate(node ast.Node) []error {
	// TODO: Check that node is a *ast.Program
	// TODO: If not, return an error
	// TODO: Call validateNode on the program's Expression
	// TODO: Return the collected errors
	return nil
}

// validateNode validates a single AST node and its children recursively.
// Returns all validation errors found in this node and its subtree.
func (v *Validator) validateNode(node ast.Node) []error {
	var errors []error

	// TODO: Use a type switch on node to handle different AST node types
	// TODO: For each node type, call the appropriate validation method
	// Hint: Handle these cases:
	//   - *ast.LogicalExpression
	//   - *ast.ComparisonExpression
	//   - *ast.UnaryExpression
	//   - *ast.Identifier
	//   - *ast.TraversalExpression
	//   - *ast.HasExpression
	//   - *ast.FunctionCall
	//   - Literal types (StringLiteral, NumberLiteral, BooleanLiteral, NullLiteral) don't need validation
	//   - default case for unknown types

	return errors
}

// Task 1: Validate simple identifier field existence
// This method should check if a field with the given name exists in the struct
func (v *Validator) validateIdentifier(name string) error {
	// TODO: Use v.findFieldByName to check if the field exists
	// TODO: If not found, return an error: "field 'X' does not exist"
	// TODO: If found, return nil
	return nil
}

// Task 2: Validate nested field traversal (e.g., user.email)
// This method should validate that we can traverse through nested structs
func (v *Validator) validateTraversal(expr *ast.TraversalExpression) (reflect.Type, error) {
	// TODO: Get the type of the left side using v.getExpressionType
	// TODO: If leftType is nil, return error: "invalid left side of traversal"
	// TODO: Check if leftType is a struct (use leftType.Kind() != reflect.Struct)
	// TODO: If not a struct, return error: "cannot traverse into non-struct type X"
	// TODO: Get the right side as an identifier
	// TODO: Use v.findFieldByName to find the field in leftType
	// TODO: If not found, return error: "field 'X' does not exist"
	// TODO: If found, return the field type
	return nil, nil
}

// Task 3: Validate comparison expression type compatibility
// This method should check that both sides of a comparison are compatible types
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) []error {
	var errors []error

	// TODO: Recursively validate both expr.Left and expr.Right
	// TODO: Get the types of both sides using v.getExpressionType
	// TODO: If either type is nil, return the errors already collected
	// TODO: Check if left side is array/slice - should use has operator instead
	// TODO: Check if left side is boolean - only = and != are allowed
	// TODO: Check type compatibility using typesCompatible function
	// TODO: If types are not compatible, add error: "type mismatch: cannot compare X with Y"

	return errors
}

// Task 4: Validate has expression (collection:member)
// This method validates the has operator (e.g., tags:"urgent")
func (v *Validator) validateHasExpression(expr *ast.HasExpression) []error {
	var errors []error

	// TODO: Validate the collection (left side)
	// TODO: Get the collection type using v.getExpressionType
	// TODO: Check that collection is an array or slice
	// TODO: If not, return error: "has operator ':' requires array/slice field, got X"
	// TODO: Get the element type using collectionType.Elem()
	// TODO: Determine the member type (right side)
	// TODO: For identifiers in has expressions, treat as strings (AIP-160 spec)
	// TODO: Check type compatibility between element type and member type
	// TODO: If not compatible, add error: "has operator: element type X incompatible with member type Y"

	return errors
}

// Task 5: Validate function call
// This method validates that a function exists and its arguments are valid
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) []error {
	var errors []error

	// TODO: Look up the function in SupportedFunctions registry
	// TODO: If not found, return error: "function 'X' is not supported"
	// TODO: Check that argument count matches funcDef.ArgCount
	// TODO: If mismatch, return error: "function 'X' expects Y argument(s), got Z"
	// TODO: For each argument:
	//   - Validate the argument node recursively
	//   - Get the argument type
	//   - Check if the type is valid for this function using isValidKind
	//   - If not, add error with details

	return errors
}

// validateLogical validates logical AND/OR expressions
func (v *Validator) validateLogical(expr *ast.LogicalExpression) []error {
	var errors []error

	// TODO: Validate left side by calling v.validateNode(expr.Left)
	// TODO: Validate right side by calling v.validateNode(expr.Right)
	// TODO: Append all errors and return

	return errors
}

// validateUnary validates unary NOT/- expressions
func (v *Validator) validateUnary(expr *ast.UnaryExpression) []error {
	// TODO: Validate the right side by calling v.validateNode(expr.Right)
	// TODO: Return any errors found
	return nil
}

// Helper: getExpressionType returns the Go type of an expression
// This is used to check type compatibility in comparisons
func (v *Validator) getExpressionType(expr ast.Expression) reflect.Type {
	// TODO: Use a type switch to determine the type based on the expression
	// Hint: Handle these cases:
	//   - *ast.Identifier: look up field type
	//   - *ast.TraversalExpression: get the type of the traversal result
	//   - *ast.StringLiteral: return reflect.TypeOf("")
	//   - *ast.NumberLiteral: return reflect.TypeOf(float64(0))
	//   - *ast.BooleanLiteral: return reflect.TypeOf(false)
	//   - *ast.FunctionCall: look up return type from SupportedFunctions
	//   - default: return nil for unknown types
	return nil
}

// Helper: findFieldByName searches for a field in a struct by name
// It handles both direct field names and tag-based matching
func (v *Validator) findFieldByName(structType reflect.Type, name string) (reflect.StructField, bool) {
	// TODO: Iterate through all fields in the struct
	// TODO: For each field, check if it matches:
	//   - If v.tagName is empty, match against field.Name
	//   - If v.tagName is set, match against the tag value
	// TODO: Handle nested anonymous fields (embedded structs) recursively
	// Hint: Use structType.NumField() and structType.Field(i)
	// Hint: For tags, use field.Tag.Get(v.tagName)
	// Hint: Tag values may be comma-separated, use the first part
	return reflect.StructField{}, false
}

// Helper: extractTagValue extracts the field name from a struct tag
func (v *Validator) extractTagValue(tag string) string {
	// TODO: Handle different tag formats:
	//   - For "json": extract from "name,omitempty" -> "name"
	//   - For "protobuf": extract from "bytes,1,opt,name=field_name,proto3" -> "field_name"
	//   - For other tags: use comma-separated first part
	// Hint: Check v.tagName to determine format
	// Hint: Use strings.Split for comma-separated values
	// Hint: For protobuf, look for "name=" prefix
	return ""
}

// Helper: typesCompatible checks if two kinds are compatible for comparison
func typesCompatible(left, right reflect.Kind) bool {
	// TODO: Implement type compatibility rules:
	//   - Numeric types are compatible with each other
	//   - String is compatible with String
	//   - Bool is compatible with Bool
	//   - Interface{} is compatible with anything
	// Hint: Group numeric types: Int, Int8, Int16, Int32, Int64, Uint, Uint8, etc., Float32, Float64
	return false
}

// Helper: isValidKind checks if a kind is in a list of valid kinds
func isValidKind(kind reflect.Kind, validKinds []reflect.Kind) bool {
	// TODO: Check if kind is in the validKinds slice
	// TODO: Return true if found, false otherwise
	return false
}
