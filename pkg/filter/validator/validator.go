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
	return func(v *Validator) {
		v.tagName = "json"
	}
}

// WithProtobufTags configures the validator to match field names against protobuf struct tags.
// It extracts the "name=" value from protobuf tags.
// For example, `Name string `protobuf:"bytes,2,opt,name=name,proto3"` would match "name".
func WithProtobufTags() ValidatorOption {
	return func(v *Validator) {
		v.tagName = "protobuf"
	}
}

// WithTag configures the validator to match field names against a custom struct tag.
// The tag value will be used as-is (comma-separated values will use the first part).
func WithTag(tagName string) ValidatorOption {
	return func(v *Validator) {
		v.tagName = tagName
	}
}

// NewValidator creates a new validator for the given struct type.
// By default, it matches against PascalCase field names.
// Optionally provide one tag option to configure tag-based matching:
//   - WithJSONTags(): match against json tag names
//   - WithProtobufTags(): match against protobuf name= values
//   - WithTag("custom"): match against any custom tag
//
// If multiple options are provided, only the first one is used.
//
// Example: NewValidator(reflect.TypeOf(User{}), WithJSONTags())
func NewValidator(structType reflect.Type, opts ...ValidatorOption) *Validator {
	v := &Validator{
		structType: structType,
	}
	// NOTE: Only apply the first option. Currently we only support one tagName option,
	// which is our only option type. This can be expanded in the future if more
	// option types (e.g., case sensitivity, custom validators) are needed.
	if len(opts) > 0 {
		opts[0](v)
	}
	return v
}

// Validate validates the entire AST and returns a list of all validation errors.
// Returns an empty slice if the AST is valid.
func (v *Validator) Validate(node ast.Node) []error {
	program, ok := node.(*ast.Program)
	if !ok {
		return []error{fmt.Errorf("expected Program node, got %T", node)}
	}

	return v.validateNode(program.Expression)
}

// validateNode validates a single AST node and its children recursively.
// Returns all validation errors found in this node and its subtree.
func (v *Validator) validateNode(node ast.Node) []error {
	var errors []error

	switch n := node.(type) {
	case *ast.LogicalExpression:
		errors = append(errors, v.validateLogical(n)...)

	case *ast.ComparisonExpression:
		errors = append(errors, v.validateComparison(n)...)

	case *ast.UnaryExpression:
		errors = append(errors, v.validateUnary(n)...)

	case *ast.Identifier:
		if err := v.validateIdentifier(n.Value); err != nil {
			errors = append(errors, err)
		}

	case *ast.TraversalExpression:
		_, err := v.validateTraversal(n)
		if err != nil {
			errors = append(errors, err)
		}

	case *ast.HasExpression:
		errors = append(errors, v.validateHasExpression(n)...)

	case *ast.FunctionCall:
		errors = append(errors, v.validateFunctionCall(n)...)

	case *ast.StringLiteral, *ast.NumberLiteral, *ast.BooleanLiteral, *ast.NullLiteral:
		// Literals don't need validation
		return nil

	default:
		// Unknown node type - shouldn't happen with current AST
		return []error{fmt.Errorf("unknown node type: %T", node)}
	}

	return errors
}

// Task 1: Validate simple identifier field existence
func (v *Validator) validateIdentifier(name string) error {
	_, ok := v.findFieldByName(v.structType, name)
	if !ok {
		return fmt.Errorf("field '%s' does not exist", name)
	}
	return nil
}

// Task 2: Validate nested field traversal (e.g., email.address)
func (v *Validator) validateTraversal(expr *ast.TraversalExpression) (reflect.Type, error) {
	// Get the type of the left side (could be identifier or another traversal)
	leftType := v.getExpressionType(expr.Left)
	if leftType == nil {
		return nil, fmt.Errorf("invalid left side of traversal")
	}

	// Check if left type is a struct
	if leftType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("cannot traverse into non-struct type %s", leftType.Kind())
	}

	// Get the right side field name
	rightIdent, ok := expr.Right.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier as traversal right side, got %T", expr.Right)
	}

	field, ok := v.findFieldByName(leftType, rightIdent.Value)
	if !ok {
		return nil, fmt.Errorf("field '%s' does not exist", rightIdent.Value)
	}

	return field.Type, nil
}

// Task 3: Validate comparison expression type compatibility
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) []error {
	var errors []error

	// First, recursively validate both sides to catch field existence errors
	errors = append(errors, v.validateNode(expr.Left)...)
	errors = append(errors, v.validateNode(expr.Right)...)

	// Now check type compatibility (if both sides are valid)
	leftType := v.getExpressionType(expr.Left)
	rightType := v.getExpressionType(expr.Right)

	// If we couldn't get types, the errors were already collected above
	if leftType == nil || rightType == nil {
		return errors
	}

	// Check if left side is array/slice - should use has operator instead
	if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
		errors = append(errors, fmt.Errorf("use has operator ':' for array/slice fields, not '%s'", expr.Operator))
		return errors
	}

	// AIP-160: Booleans only support = and != operators
	if leftType.Kind() == reflect.Bool {
		if expr.Operator != "=" && expr.Operator != "!=" {
			errors = append(errors, fmt.Errorf("boolean fields only support = and != operators, not '%s'", expr.Operator))
			return errors
		}
	}

	// Check type compatibility
	if !typesCompatible(leftType.Kind(), rightType.Kind()) {
		errors = append(errors, fmt.Errorf("type mismatch: cannot compare %s with %s", leftType.Kind(), rightType.Kind()))
	}

	return errors
}

// Task 4: Validate has expression (array:value)
func (v *Validator) validateHasExpression(expr *ast.HasExpression) []error {
	var errors []error

	// Validate the collection (left side) - this should be a field
	errors = append(errors, v.validateNode(expr.Collection)...)

	// For the member (right side), we DON'T validate it as a field reference
	// In AIP-160, identifiers in has expressions are treated as string literals
	// Example: Tags:urgent means check if "urgent" is in Tags array

	// Get the collection type
	collectionType := v.getExpressionType(expr.Collection)
	if collectionType == nil {
		// Error already collected above
		return errors
	}

	// Check that collection is actually an array/slice
	if collectionType.Kind() != reflect.Slice && collectionType.Kind() != reflect.Array {
		errors = append(errors, fmt.Errorf("has operator ':' requires array/slice field, got %s", collectionType.Kind()))
		return errors
	}

	// Get element type
	elemType := collectionType.Elem()

	// Determine member type
	var memberType reflect.Type
	switch expr.Member.(type) {
	case *ast.StringLiteral:
		memberType = reflect.TypeOf("")
	case *ast.NumberLiteral:
		memberType = reflect.TypeOf(float64(0))
	case *ast.BooleanLiteral:
		memberType = reflect.TypeOf(false)
	case *ast.Identifier:
		// In AIP-160, bare identifiers in has expressions are treated as strings
		// Example: Tags:urgent is equivalent to Tags:"urgent"
		memberType = reflect.TypeOf("")
	default:
		// For other expression types (functions, etc), try to get their type
		memberType = v.getExpressionType(expr.Member)
		if memberType == nil {
			errors = append(errors, fmt.Errorf("cannot determine type of has expression member"))
			return errors
		}
	}

	// Check type compatibility
	if !typesCompatible(elemType.Kind(), memberType.Kind()) {
		errors = append(errors, fmt.Errorf("has operator: element type %s incompatible with member type %s", elemType.Kind(), memberType.Kind()))
	}

	return errors
}

// Task 5: Validate function call
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) []error {
	var errors []error

	// Look up function in registry
	funcDef, ok := SupportedFunctions[expr.Function]
	if !ok {
		return []error{fmt.Errorf("function '%s' is not supported", expr.Function)}
	}

	// Check argument count
	if len(expr.Arguments) != funcDef.ArgCount {
		errors = append(errors, fmt.Errorf("function '%s' expects %d argument(s), got %d",
			expr.Function, funcDef.ArgCount, len(expr.Arguments)))
		return errors
	}

	// Validate each argument
	for i, arg := range expr.Arguments {
		// First validate that the argument node itself is valid
		argErrors := v.validateNode(arg)
		errors = append(errors, argErrors...)

		// Then check type compatibility
		argType := v.getExpressionType(arg)
		if argType != nil && !isValidKind(argType.Kind(), funcDef.ValidFieldKinds) {
			errors = append(errors, fmt.Errorf("function '%s' argument %d: expected field types %v, got %s",
				expr.Function, i+1, funcDef.ValidFieldKinds, argType.Kind()))
		}
	}

	return errors
}

// validateLogical validates logical AND/OR expressions
func (v *Validator) validateLogical(expr *ast.LogicalExpression) []error {
	var errors []error

	// Validate left side
	leftErrs := v.validateNode(expr.Left)
	errors = append(errors, leftErrs...)

	// Validate right side
	rightErrs := v.validateNode(expr.Right)
	errors = append(errors, rightErrs...)

	return errors
}

// validateUnary validates unary NOT expressions
func (v *Validator) validateUnary(expr *ast.UnaryExpression) []error {
	return v.validateNode(expr.Right)
}

// getExpressionType returns the Go type for an expression node
func (v *Validator) getExpressionType(node ast.Node) reflect.Type {
	switch n := node.(type) {
	case *ast.Identifier:
		field, ok := v.findFieldByName(v.structType, n.Value)
		if !ok {
			return nil
		}
		return field.Type

	case *ast.TraversalExpression:
		typ, err := v.validateTraversal(n)
		if err != nil {
			return nil
		}
		return typ

	case *ast.StringLiteral:
		var s string
		return reflect.TypeOf(s)

	case *ast.NumberLiteral:
		// Check if it's an integer or float based on the value
		if n.Value == float64(int64(n.Value)) {
			var i int
			return reflect.TypeOf(i)
		}
		var f float64
		return reflect.TypeOf(f)

	case *ast.BooleanLiteral:
		var b bool
		return reflect.TypeOf(b)

	case *ast.FunctionCall:
		// For now, assume functions return int64 (like timestamp)
		// In a more sophisticated system, you'd look up the return type
		var i int64
		return reflect.TypeOf(i)

	default:
		return nil
	}
}

// findFieldByName finds a struct field by name, respecting the configured tag name.
// If tagName is empty, it uses the field name directly (PascalCase).
// Otherwise, it searches through fields matching the tag value.
func (v *Validator) findFieldByName(structType reflect.Type, name string) (reflect.StructField, bool) {
	// Default behavior: use field name directly
	if v.tagName == "" {
		return structType.FieldByName(name)
	}

	// Search through all fields to find matching tag
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tagValue := field.Tag.Get(v.tagName)

		if tagValue == "" {
			continue
		}

		// Extract the field name from the tag
		var tagFieldName string
		if v.tagName == "protobuf" {
			// Protobuf format: "bytes,2,opt,name=field_name,proto3"
			tagFieldName = extractProtobufName(tagValue)
		} else {
			// JSON and most other tags: "field_name,omitempty"
			// Take the first part before comma
			tagFieldName = strings.Split(tagValue, ",")[0]
		}

		if tagFieldName == name {
			return field, true
		}
	}

	return reflect.StructField{}, false
}

// extractProtobufName extracts the "name=value" from a protobuf tag.
// Example: "bytes,2,opt,name=field_name,proto3" returns "field_name"
func extractProtobufName(tag string) string {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "name=") {
			return strings.TrimPrefix(part, "name=")
		}
	}
	return ""
}

// typesCompatible checks if two types are compatible for comparison
func typesCompatible(left, right reflect.Kind) bool {
	// Same type is always compatible
	if left == right {
		return true
	}

	// Strings only compatible with strings
	if left == reflect.String || right == reflect.String {
		return left == right
	}

	// Bools only compatible with bools
	if left == reflect.Bool || right == reflect.Bool {
		return left == right
	}

	// All numeric types are compatible with each other
	if isNumericKind(left) && isNumericKind(right) {
		return true
	}

	return false
}

// isNumericKind checks if a kind represents a numeric type
func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// isValidKind checks if a kind is in a list of valid kinds
func isValidKind(kind reflect.Kind, validKinds []reflect.Kind) bool {
	for _, validKind := range validKinds {
		if kind == validKind {
			return true
		}
	}
	return false
}
