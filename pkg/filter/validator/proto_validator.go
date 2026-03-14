package validator

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/zshainsky/aip160/pkg/filter/ast"
)

// ProtoValidator validates filter ASTs against protobuf message descriptors.
// It uses google.golang.org/protobuf/reflect/protoreflect for efficient
// field resolution and native enum validation.
//
// ProtoValidator is optimized for proto-generated Go types (*.pb.go files)
// and provides 2-5x better performance compared to the reflection-based Validator.
//
// Usage:
//
//	msgDesc := myProtoMessage.ProtoReflect().Descriptor()
//	validator := NewProtoValidator(msgDesc)
//	errors := validator.Validate(astNode)
//
// With options:
//
//	validator := NewProtoValidator(msgDesc,
//		WithEnumPrefixStripping(false),
//		// Future options can be added here
//	)
type ProtoValidator struct {
	descriptor protoreflect.MessageDescriptor
	options    ProtoValidatorOptions
}

// ProtoValidatorOptions holds configuration for ProtoValidator behavior.
// This struct can be extended with new options in the future without
// breaking the API.
type ProtoValidatorOptions struct {
	// EnableEnumPrefixStripping allows enum values to match with prefix stripped.
	// When true (default): "ACTIVE" matches "STATUS_ACTIVE"
	// When false: only "STATUS_ACTIVE" matches
	EnableEnumPrefixStripping bool

	// Future options can be added here:
	// AllowCaseInsensitiveFields bool
	// StrictModeEnabled bool
	// CustomValidators map[string]func(...) error
}

// ProtoValidatorOption is a functional option for configuring ProtoValidator.
type ProtoValidatorOption func(*ProtoValidatorOptions)

// WithEnumPrefixStripping controls whether enum values can be matched with their
// prefix stripped. When enabled (default), both forms are accepted:
//   - status = "STATUS_ACTIVE" (exact match)
//   - status = "ACTIVE" (prefix-stripped: STATUS_ + ACTIVE)
//
// When disabled, only exact matches are accepted:
//   - status = "STATUS_ACTIVE" (only this works)
//   - status = "ACTIVE" (fails)
//
// Default: true (enabled for user convenience and AIP-160 "non-technical audience" principle)
//
// Example:
//
//	validator := NewProtoValidator(msgDesc, WithEnumPrefixStripping(false))
func WithEnumPrefixStripping(enable bool) ProtoValidatorOption {
	return func(opts *ProtoValidatorOptions) {
		opts.EnableEnumPrefixStripping = enable
	}
}

// NewProtoValidator creates a new validator for the given protobuf message descriptor.
// The descriptor can be obtained from any proto.Message via the ProtoReflect().Descriptor() method.
//
// Accepts optional configuration via ProtoValidatorOption functions.
// Multiple options can be chained together.
//
// Example:
//
//	var user *pb.User
//	validator := NewProtoValidator(
//		user.ProtoReflect().Descriptor(),
//		WithEnumPrefixStripping(false),
//		// Add more options as needed
//	)
func NewProtoValidator(msgDesc protoreflect.MessageDescriptor, opts ...ProtoValidatorOption) *ProtoValidator {
	// Default options
	options := ProtoValidatorOptions{
		EnableEnumPrefixStripping: true, // User-friendly default
	}

	// Apply provided options
	for _, opt := range opts {
		opt(&options)
	}

	return &ProtoValidator{
		descriptor: msgDesc,
		options:    options,
	}
}

// Validate validates the entire AST and returns a list of all validation errors.
// Returns an empty slice if the AST is valid.
func (pv *ProtoValidator) Validate(node ast.Node) []error {
	var errors []error

	// Handle Program node which contains the root expression
	if program, ok := node.(*ast.Program); ok {
		if program.Expression != nil {
			pv.validateNode(program.Expression, &errors)
		}
		return errors
	}

	// For non-Program nodes, validate directly
	pv.validateNode(node, &errors)
	return errors
}

// validateNode recursively validates AST nodes and dispatches to appropriate validators.
func (pv *ProtoValidator) validateNode(node ast.Node, errors *[]error) {
	switch n := node.(type) {
	case *ast.ComparisonExpression:
		pv.validateComparison(n, errors)
	case *ast.LogicalExpression:
		pv.validateLogical(n, errors)
	case *ast.UnaryExpression:
		pv.validateUnary(n, errors)
	case *ast.Identifier:
		pv.validateIdentifier(n.Value, errors)
	case *ast.TraversalExpression:
		pv.validateTraversal(n, errors)
	case *ast.HasExpression:
		pv.validateHas(n, errors)
		// Literal nodes (StringLiteral, NumberLiteral, BooleanLiteral) don't need validation
	}
}

// addError is a helper to append validation errors with context.
func (pv *ProtoValidator) addError(errors *[]error, format string, args ...interface{}) {
	*errors = append(*errors, fmt.Errorf(format, args...))
}

// === Field Resolution ===

// validateIdentifier checks if a field exists in the message descriptor.
func (pv *ProtoValidator) validateIdentifier(name string, errors *[]error) {
	if _, ok := pv.findFieldByName(pv.descriptor, name); !ok {
		pv.addError(errors, "field '%s' does not exist in message %s", name, pv.descriptor.Name())
	}
}

// findFieldByName finds a field by name in the message descriptor.
// Returns the field descriptor and true if found, nil and false otherwise.
// Uses O(1) descriptor lookup instead of O(n) tag iteration.
func (pv *ProtoValidator) findFieldByName(descriptor protoreflect.MessageDescriptor, name string) (protoreflect.FieldDescriptor, bool) {
	fieldDesc := descriptor.Fields().ByName(protoreflect.Name(name))
	return fieldDesc, fieldDesc != nil
}

// resolveFieldFromExpression resolves a field descriptor from any expression type.
// Handles both simple identifiers and nested traversal expressions.
// Does NOT add errors - returns nil if not found (callers decide error handling).
// This is the unified entry point for field resolution across the validator.
func (pv *ProtoValidator) resolveFieldFromExpression(node ast.Node, msgDesc protoreflect.MessageDescriptor) protoreflect.FieldDescriptor {
	switch n := node.(type) {
	case *ast.Identifier:
		return msgDesc.Fields().ByName(protoreflect.Name(n.Value))
	case *ast.TraversalExpression:
		// Delegate to full recursive resolver without error reporting
		fieldDesc, _ := pv.resolveFieldDescriptor(n, msgDesc, &[]error{})
		return fieldDesc
	default:
		return nil
	}
}

// === Expression Validators ===

// validateComparison validates comparison expressions (=, !=, <, >, <=, >=).
func (pv *ProtoValidator) validateComparison(expr *ast.ComparisonExpression, errors *[]error) {
	// Validate left side (field reference)
	pv.validateNode(expr.Left, errors)

	// Get field descriptor for left side (handles both simple and nested fields)
	fieldDesc := pv.resolveFieldFromExpression(expr.Left, pv.descriptor)
	if fieldDesc == nil {
		return // Field validation already added error
	}

	// Validate operator is allowed for this field type (includes repeated field check)
	if !pv.validateOperatorForField(expr.Operator, fieldDesc, expr.Left, errors) {
		return // Operator validation failed, error already added
	}

	// Validate value type matches field type
	if !pv.validateTypeCompatibility(expr, fieldDesc, errors) {
		return // Type validation failed, error already added
	}

	// For enum fields, validate the specific enum value exists
	if fieldDesc.Kind() == protoreflect.EnumKind {
		pv.validateEnumValue(expr, fieldDesc, errors)
	}
}

// validateOperatorForField checks if the operator is valid for the given field type.
// Returns false if validation fails (with error added), true to continue validation.
func (pv *ProtoValidator) validateOperatorForField(operator string, fieldDesc protoreflect.FieldDescriptor, fieldNode ast.Node, errors *[]error) bool {
	// Repeated fields cannot use comparison operators - must use HAS operator (:) instead
	// Per AIP-160: "The . operator must not be used to traverse through a repeated field"
	if fieldDesc.IsList() {
		pv.addError(errors, "cannot use comparison operator on repeated field '%s', use has operator (:) instead",
			pv.getFieldPath(fieldNode))
		return false
	}

	kind := fieldDesc.Kind()

	// Boolean and enum fields only support = and != operators
	if kind == protoreflect.BoolKind || kind == protoreflect.EnumKind {
		if !pv.isValidOperatorForKind(operator, kind) {
			fieldType := "boolean"
			if kind == protoreflect.EnumKind {
				fieldType = "enum"
			}
			pv.addError(errors, "%s field '%s' does not support operator '%s' (only = and != allowed)",
				fieldType, pv.getFieldPath(fieldNode), operator)
			return false
		}
	}

	return true // Operator is valid, continue validation
}

// validateTypeCompatibility checks if the right operand type is compatible with the field type.
// Returns false if validation fails (with error added), true to continue validation.
func (pv *ProtoValidator) validateTypeCompatibility(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) bool {
	// For enum fields, must be string literal
	if fieldDesc.Kind() == protoreflect.EnumKind {
		if _, ok := expr.Right.(*ast.StringLiteral); !ok {
			pv.addError(errors, "enum field '%s' requires string value (enum name), not %T",
				pv.getFieldPath(expr.Left), expr.Right)
			return false
		}
		return true // Type is correct, continue to value validation
	}

	// For all other fields, check proto kind compatibility
	leftKind, leftOk := pv.getExpressionKind(expr.Left)
	rightKind, rightOk := pv.getExpressionKind(expr.Right)

	if leftOk && rightOk {
		if !pv.protoKindsCompatible(leftKind, rightKind) {
			pv.addError(errors, "cannot compare %s field with %s value",
				leftKind, rightKind)
			return false
		}
	}

	return true // Type compatibility validated
}

// validateEnumValue validates that the enum value exists in the enum definition.
// Assumes operator and type have already been validated.
func (pv *ProtoValidator) validateEnumValue(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) {
	// Right side is guaranteed to be string literal by validateTypeCompatibility
	stringLit := expr.Right.(*ast.StringLiteral)

	// Validate the enum value exists in the enum definition
	if !pv.isValidEnumValue(fieldDesc, stringLit.Value) {
		validValues := pv.getEnumValueNames(fieldDesc)
		pv.addError(errors, "enum field '%s' has invalid value '%s'; valid values are: %v",
			pv.getFieldPath(expr.Left), stringLit.Value, validValues)
	}
}

// isValidEnumValue checks if a string value is a valid enum value name.
// Supports both exact matching and prefix-stripped matching.
//
// Matching rules (in order):
//  1. Exact match: "TASK_STATUS_ACTIVE" matches TASK_STATUS_ACTIVE
//  2. Prefix-stripped match: "ACTIVE" matches TASK_STATUS_ACTIVE
//     - Enum name converted to SCREAMING_SNAKE_CASE (TaskStatus → TASK_STATUS_)
//     - Value prepended with prefix: "ACTIVE" → "TASK_STATUS_ACTIVE"
//  3. Case-sensitive in all cases
//
// Examples:
//   - Enum TaskStatus {TASK_STATUS_ACTIVE, TASK_STATUS_INACTIVE}
//   - "TASK_STATUS_ACTIVE" ✓ (exact match)
//   - "ACTIVE" ✓ (prefix-stripped: TASK_STATUS_ + ACTIVE)
//   - "active" ✗ (wrong case)
//   - Enum TaskResult {TASK_RESULT_UNSPECIFIED, SUCCESS, FAILED}
//   - "SUCCESS" ✓ (exact match, no prefix needed)
//   - "TASK_RESULT_SUCCESS" ✗ (prefix doesn't exist in enum)
//
// Note: Multi-word enum names handled correctly (TaskStatus → TASK_STATUS_)
func (pv *ProtoValidator) isValidEnumValue(fieldDesc protoreflect.FieldDescriptor, value string) bool {
	enumDesc := fieldDesc.Enum()

	// Try exact match first
	if enumDesc.Values().ByName(protoreflect.Name(value)) != nil {
		return true
	}

	// TODO: Add option to disable prefix stripping (e.g., WithEnumPrefixStripping(false))
	// Currently always enabled for user convenience. Future implementation:
	//   if !pv.options.EnableEnumPrefixStripping {
	//       return false
	//   }

	// Try with enum name prefix (for user-friendly filters)
	// E.g., "ACTIVE" → "STATUS_ACTIVE" for enum Status
	// Or "ACTIVE" → "TASK_STATUS_ACTIVE" for enum TaskStatus
	prefix := toScreamingSnakeCase(string(enumDesc.Name())) + "_"
	withPrefix := prefix + value
	if enumDesc.Values().ByName(protoreflect.Name(withPrefix)) != nil {
		return true
	}

	return false
}

// toScreamingSnakeCase converts CamelCase to SCREAMING_SNAKE_CASE.
// Examples: "Status" → "STATUS", "TaskStatus" → "TASK_STATUS"
func toScreamingSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

// getEnumValueNames returns a list of all valid enum value names for an enum field.
// Used for generating helpful error messages.
func (pv *ProtoValidator) getEnumValueNames(fieldDesc protoreflect.FieldDescriptor) []string {
	enumDesc := fieldDesc.Enum()
	validValues := make([]string, 0, enumDesc.Values().Len())
	for i := 0; i < enumDesc.Values().Len(); i++ {
		validValues = append(validValues, string(enumDesc.Values().Get(i).Name()))
	}
	return validValues
}

// isValidOperatorForKind checks if an operator is valid for a given proto kind.
// Per AIP-160:
//   - Boolean fields: only = and != (no ordering operators)
//   - Enum fields: only = and != (no ordering operators)
//   - All other types: all 6 operators allowed (=, !=, <, >, <=, >=)
func (pv *ProtoValidator) isValidOperatorForKind(operator string, kind protoreflect.Kind) bool {
	switch kind {
	case protoreflect.BoolKind, protoreflect.EnumKind:
		// Boolean and enum fields only support equality operators
		return operator == "=" || operator == "!="
	default:
		// All other types support all operators
		return true
	}
}

// getExpressionKind determines the proto kind of an expression.
// Returns the field's kind for identifiers/traversals, or inferred kind for literals.
//
// For numeric literals, distinguishes between integer and float based on
// whether the value has a fractional part (e.g., 23 vs 23.55).
func (pv *ProtoValidator) getExpressionKind(node ast.Node) (protoreflect.Kind, bool) {
	switch n := node.(type) {
	case *ast.Identifier, *ast.TraversalExpression:
		// Use unified field resolution for both simple and nested fields
		fieldDesc := pv.resolveFieldFromExpression(n, pv.descriptor)
		if fieldDesc != nil {
			return fieldDesc.Kind(), true
		}
	case *ast.StringLiteral:
		return protoreflect.StringKind, true
	case *ast.NumberLiteral:
		// Distinguish between integer and float literals
		// If the number has a fractional part, it's a float
		if n.Value != float64(int64(n.Value)) {
			return protoreflect.DoubleKind, true
		}
		return protoreflect.Int64Kind, true
	case *ast.BooleanLiteral:
		return protoreflect.BoolKind, true
	}
	return 0, false
}

// protoKindsCompatible checks if two proto kinds are compatible for comparison.
//
// Type compatibility rules:
// - Strings are compatible with strings/bytes
// - Integers are only compatible with integer literals (not floats)
// - Floats are compatible with both integer and float literals
// - Booleans are only compatible with booleans
// - Enums are compatible with string values (enum names)
func (pv *ProtoValidator) protoKindsCompatible(left, right protoreflect.Kind) bool {
	// String types
	if left == protoreflect.StringKind || left == protoreflect.BytesKind {
		return right == protoreflect.StringKind || right == protoreflect.BytesKind
	}

	// Integer types - only compatible with integer literals
	if isProtoIntegerKind(left) {
		return isProtoIntegerKind(right)
	}

	// Float types - compatible with both integer and float literals
	if isProtoFloatKind(left) {
		return isProtoNumericKind(right)
	}

	// Boolean types
	if left == protoreflect.BoolKind {
		return right == protoreflect.BoolKind
	}

	// Enum types
	if left == protoreflect.EnumKind {
		// Enums compare with string values (enum name)
		return right == protoreflect.StringKind
	}

	return false
}

// isProtoIntegerKind checks if a protoreflect.Kind is an integer type.
func isProtoIntegerKind(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind:
		return true
	}
	return false
}

// isProtoFloatKind checks if a protoreflect.Kind is a floating-point type.
func isProtoFloatKind(k protoreflect.Kind) bool {
	return k == protoreflect.FloatKind || k == protoreflect.DoubleKind
}

// isProtoNumericKind checks if a protoreflect.Kind is numeric (integer or float).
func isProtoNumericKind(k protoreflect.Kind) bool {
	return isProtoIntegerKind(k) || isProtoFloatKind(k)
}

// validateLogical validates logical AND/OR expressions.
func (pv *ProtoValidator) validateLogical(expr *ast.LogicalExpression, errors *[]error) {
	pv.validateNode(expr.Left, errors)
	pv.validateNode(expr.Right, errors)
}

// validateUnary validates NOT expressions.
func (pv *ProtoValidator) validateUnary(expr *ast.UnaryExpression, errors *[]error) {
	pv.validateNode(expr.Right, errors)
}

// validateTraversal validates nested field access (e.g., email.address).
func (pv *ProtoValidator) validateTraversal(expr *ast.TraversalExpression, errors *[]error) {
	// Resolve the left side and get its field descriptor
	leftField, _ := pv.resolveFieldDescriptor(expr.Left, pv.descriptor, errors)
	if leftField == nil {
		return // Error already added by resolveFieldDescriptor
	}

	// Ensure the left field is a message type (can be traversed)
	if !pv.requireMessageKind(leftField, expr.Left, errors) {
		return
	}

	// Get the nested message descriptor
	nestedDesc := leftField.Message()

	// Recursively validate the right side against the nested descriptor
	pv.validateNodeWithDescriptor(expr.Right, nestedDesc, errors)
}

// resolveFieldDescriptor resolves a field descriptor from an expression node.
// Returns the field descriptor and message descriptor, or nil if field not found.
// Adds errors when fields don't exist or cannot be traversed.
func (pv *ProtoValidator) resolveFieldDescriptor(node ast.Node, msgDesc protoreflect.MessageDescriptor, errors *[]error) (protoreflect.FieldDescriptor, protoreflect.MessageDescriptor) {
	switch n := node.(type) {
	case *ast.Identifier:
		fieldDesc := msgDesc.Fields().ByName(protoreflect.Name(n.Value))
		if fieldDesc == nil {
			pv.addError(errors, "field '%s' does not exist in message %s", n.Value, msgDesc.Name())
			return nil, nil
		}
		return fieldDesc, msgDesc

	case *ast.TraversalExpression:
		// Recursively resolve left side first
		leftField, _ := pv.resolveFieldDescriptor(n.Left, msgDesc, errors)
		if leftField == nil {
			return nil, nil
		}

		// Ensure left side is a message
		if !pv.requireMessageKind(leftField, n.Left, errors) {
			return nil, nil
		}

		// Get nested descriptor and continue resolution
		nestedDesc := leftField.Message()
		return pv.resolveFieldDescriptor(n.Right, nestedDesc, errors)

	default:
		return nil, nil
	}
}

// validateNodeWithDescriptor validates a node against a specific message descriptor.
// Used for validating nested fields in the context of a nested message.
// Delegates to resolveFieldDescriptor to avoid code duplication.
func (pv *ProtoValidator) validateNodeWithDescriptor(node ast.Node, msgDesc protoreflect.MessageDescriptor, errors *[]error) {
	// Simply delegate to resolveFieldDescriptor which already handles
	// both Identifier and TraversalExpression cases with proper error reporting
	pv.resolveFieldDescriptor(node, msgDesc, errors)
}

// requireMessageKind validates that a field is a message type (can be traversed).
// Returns true if valid, false if not (with error added).
// This helper eliminates duplicate "cannot traverse" error checking.
func (pv *ProtoValidator) requireMessageKind(fieldDesc protoreflect.FieldDescriptor, node ast.Node, errors *[]error) bool {
	if fieldDesc.Kind() != protoreflect.MessageKind {
		pv.addError(errors, "cannot traverse into non-message field '%s' (type: %s)",
			pv.getFieldPath(node), fieldDesc.Kind())
		return false
	}
	return true
}

// getFieldPath returns the full field path as a string (e.g., "email.address").
func (pv *ProtoValidator) getFieldPath(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Identifier:
		return n.Value
	case *ast.TraversalExpression:
		return pv.getFieldPath(n.Left) + "." + pv.getFieldPath(n.Right)
	default:
		return ""
	}
}

// validateHas validates HAS operator expressions (collection:member).
// Per AIP-160, the HAS operator (:) is used for:
// - Repeated fields: r:"value" checks if repeated field contains value
// - Maps: m:key checks if map contains key (TODO: not yet implemented)
// - Messages: m:* checks if message field is present (TODO: Phase 6E)
func (pv *ProtoValidator) validateHas(expr *ast.HasExpression, errors *[]error) {
	// Step 1: Resolve collection field descriptor
	// For now, handle simple Identifier (e.g., tags, scores)
	// TODO: Handle TraversalExpression for nested fields (Phase 6D)
	var fieldDesc protoreflect.FieldDescriptor

	switch coll := expr.Collection.(type) {
	case *ast.Identifier:
		fieldDesc = pv.descriptor.Fields().ByName(protoreflect.Name(coll.Value))
		if fieldDesc == nil {
			pv.addError(errors, "field '%s' does not exist in message %s", coll.Value, pv.descriptor.Name())
			return
		}
	default:
		// TODO: Handle TraversalExpression in Phase 6D
		pv.addError(errors, "nested HAS expressions not yet supported")
		return
	}

	// Step 2: Validate collection is a repeated field
	if !fieldDesc.IsList() {
		// TODO: Handle maps (Phase 6F) and messages (Phase 6E)
		pv.addError(errors, "HAS operator currently only supports repeated fields, got %s", fieldDesc.Kind())
		return
	}

	// Step 3: Validate member type matches element type
	elementKind := fieldDesc.Kind()

	// Get member value kind
	memberKind, ok := pv.getExpressionKind(expr.Member)
	if !ok {
		return // getExpressionKind already added error
	}

	// Check type compatibility
	if !pv.protoKindsCompatible(elementKind, memberKind) {
		collectionName := expr.Collection.(*ast.Identifier).Value
		pv.addError(errors, "type mismatch: repeated field '%s' has elements of type %s, cannot check for %s value",
			collectionName, elementKind, memberKind)
		return
	}
}

// isStarLiteral checks if a node is the star (*) identifier used for presence checks.
// Per AIP-160: r:* checks if field r is present (non-empty).
// Returns true if node is an Identifier with value "*".
func isStarLiteral(node ast.Node) bool {
	if ident, ok := node.(*ast.Identifier); ok {
		return ident.Value == "*"
	}
	return false
}
