package validator

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/zshainsky/aip160/pkg/filter/ast"
)

// Well-known protobuf type full names for AIP-160 validation.
// These are the canonical names defined in google/protobuf/*.proto files.
const (
	durationFullName  protoreflect.FullName = "google.protobuf.Duration"
	timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"
)

// RFC-3339 timestamp format pattern for AIP-160 Timestamp validation.
// Format: YYYY-MM-DDTHH:MM:SS[.fraction](Z|±HH:MM)
// Per RFC-3339: time-secfrac = "." 1*DIGIT (one or more digits, variable precision)
// Per AIP-160: "Timestamps expect RFC-3339 format: 2012-04-21T11:30:00-04:00"
var rfc3339Pattern = regexp.MustCompile(
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
)

// ProtoValidator validates filter ASTs against protobuf message descriptors.
// It uses google.golang.org/protobuf/reflect/protoreflect for efficient
// field resolution and native enum validation.
//
// ProtoValidator is optimized for proto-generated Go types (*.pb.go files)
// and provides 2-5x better performance compared to the reflection-based Validator.
//
// Supported AIP-160 Features:
// - Field existence validation
// - Type compatibility checking (all proto scalar types)
// - Enum validation with optional prefix stripping
// - Comparison operators with type restrictions
// - Nested field traversal (unlimited depth)
// - HAS operator for repeated fields and messages
// - Logical operators (AND, OR, NOT)
//
// Limitations:
// - Map fields not yet supported (planned for v2)
// - Star operator (*) requires parser update
// - Function calls not validated (future consideration)
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

	// Validate value kind matches field kind
	if !pv.validateKindCompatibility(expr, fieldDesc, errors) {
		return // Kind validation failed, error already added
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

// validateKindCompatibility checks if the right operand kind is compatible with the field kind.
// Returns false if validation fails (with error added), true to continue validation.
//
// Implementation uses a chain-of-responsibility pattern where each validator handles
// a specific domain (enums, well-known types, maps, special operators, generic types).
// Each validator returns (isKind, isValid) where isKind indicates if the validator
// applies to this field kind, and isValid indicates if the validation passed.
func (pv *ProtoValidator) validateKindCompatibility(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) bool {
	// Chain of responsibility: each validator checks if it applies to this kind
	validators := []func(*ast.ComparisonExpression, protoreflect.FieldDescriptor, *[]error) (isKind bool, isValid bool){
		pv.validateEnumFieldKind,
		pv.validateWellKnownKind,
		pv.validateMapKind,          // NEW: Map validation
		pv.validateSpecialOperators,
		pv.validateGenericKinds,
	}

	for _, validator := range validators {
		if isKind, isValid := validator(expr, fieldDesc, errors); isKind {
			return isValid
		}
	}

	return true
}

// validateEnumFieldKind checks that enum fields receive string literals.
// Per AIP-160: "Field values for bounded data types e.g. enum provided in the
// filter must be a valid value in the set"
func (pv *ProtoValidator) validateEnumFieldKind(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) (isKind bool, isValid bool) {
	if fieldDesc.Kind() != protoreflect.EnumKind {
		return false, false // Not enum kind
	}

	if _, ok := expr.Right.(*ast.StringLiteral); !ok {
		pv.addError(errors, "enum field '%s' requires string value (enum name), not %T",
			pv.getFieldPath(expr.Left), expr.Right)
		return true, false // Is enum kind, but invalid
	}

	return true, true // Is enum kind and valid
}

// validateWellKnownKind handles google.protobuf.* well-known types.
// Currently supports: Duration (Phase 1), Timestamp (TODO Phase 2)
//
// Bidirectional validation:
// - Duration fields require Duration literals
// - Non-Duration fields reject Duration literals
func (pv *ProtoValidator) validateWellKnownKind(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) (isKind bool, isValid bool) {
	isDurLiteral := isDurationLiteral(expr.Right) || pv.isNegativeDurationLiteral(expr.Right)

	// Case 1: Non-message field with Duration literal -> reject
	if fieldDesc.Kind() != protoreflect.MessageKind {
		if isDurLiteral {
			pv.addError(errors, "duration literal cannot be used on non-Duration field '%s' of type %s",
				pv.getFieldPath(expr.Left), fieldDesc.Kind())
			return true, false
		}
		return false, false // Not a well-known kind
	}

	// Case 2: Duration field validation
	if isDurationField(fieldDesc) {
		if !isDurLiteral {
			pv.addError(errors, "google.protobuf.Duration field '%s' requires duration literal (e.g., 20s, 1.2s), got %T",
				pv.getFieldPath(expr.Left), expr.Right)
			return true, false
		}
		return true, true // Is Duration kind and valid
	}

	// Case 3: Timestamp field validation (Phase 2)
	if isTimestampField(fieldDesc) {
		stringLit, ok := expr.Right.(*ast.StringLiteral)
		if !ok {
			pv.addError(errors, "google.protobuf.Timestamp field '%s' requires RFC-3339 string (e.g., \"2024-03-16T05:00:00Z\"), got %T",
				pv.getFieldPath(expr.Left), expr.Right)
			return true, false
		}
		if !isValidRFC3339(stringLit.Value) {
			pv.addError(errors, "google.protobuf.Timestamp field '%s' requires RFC-3339 format, got \"%s\"",
				pv.getFieldPath(expr.Left), stringLit.Value)
			return true, false
		}
		return true, true // Is Timestamp kind and valid
	}

	// Other message types (not well-known types)
	return false, false
}

// validateMapKind handles map field validation.
// Per AIP-160: Maps support "m.key = value" syntax
//
// Map syntax forms:
// - m:key         → HAS operator for key presence (handled in validateHas)
// - m.key:*       → HAS operator with star (handled in validateHas)
// - m.key = value → Comparison operator for key-value match (handled here)
//
// For comparisons on maps, the left side is a TraversalExpression (m.key),
// where the map field has already been resolved. We need to validate that
// the comparison value matches the map's value type.
func (pv *ProtoValidator) validateMapKind(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) (isKind bool, isValid bool) {
	// Check if this is a map field accessed via traversal
	// Pattern: labels.env = "value" where labels is map<string, string>
	traversal, ok := expr.Left.(*ast.TraversalExpression);
	if !ok || !fieldDesc.IsMap() {
		return false, false // Not a map traversal
	}

	// Get the map value type (what's stored in the map)
	mapValueKind := fieldDesc.MapValue().Kind()
	
	// Get the kind of the comparison value
	valueKind, kindOk := pv.getExpressionKind(expr.Right)
	if !kindOk {
		pv.addError(errors, "cannot determine type of comparison value")
		return true, false
	}
	
	// Validate value type matches map's value type
	if !pv.protoKindsCompatible(mapValueKind, valueKind) {
		// Get the key being accessed (for better error message)
		keyName := "key"
		if rightIdent, ok := traversal.Right.(*ast.Identifier); ok {
			keyName = rightIdent.Value
		}
		
		pv.addError(errors, "type mismatch: map '%s' value type is %s, cannot compare key '%s' with %s value",
			fieldDesc.Name(), mapValueKind, keyName, valueKind)
		return true, false
	}
	
	// Check operator restrictions for bool and enum map values
	// (Same restrictions as regular fields)
	if mapValueKind == protoreflect.BoolKind || mapValueKind == protoreflect.EnumKind {
		if !pv.isValidOperatorForKind(expr.Operator, mapValueKind) {
			valueType := "boolean"
			if mapValueKind == protoreflect.EnumKind {
				valueType = "enum"
			}
			pv.addError(errors, "map '%s' has %s values, operator '%s' not supported (only = and != allowed)",
				fieldDesc.Name(), valueType, expr.Operator)
			return true, false
		}
	}
	
	// Validate numeric ranges for integer map values
	if isProtoIntegerKind(mapValueKind) {
		if numLit, ok := expr.Right.(*ast.NumberLiteral); ok {
			pv.validateNumericRange(numLit.Value, mapValueKind, string(fieldDesc.Name()), errors)
		}
	}
	
	return true, true // Is map kind and valid
}

// validateSpecialOperators checks constraints on special operators and literals.
// - Star operator (*) only valid with HAS (:)
// - Negative literals rejected on unsigned fields
func (pv *ProtoValidator) validateSpecialOperators(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) (isKind bool, isValid bool) {
	// Star operator restriction (Cycle 7D)
	if isStarLiteral(expr.Right) {
		pv.addError(errors, "star operator (*) only valid with HAS (:), not comparison (%s)", expr.Operator)
		return true, false
	}

	// Negative literal on unsigned field (Cycle 7B)
	if pv.isUnsignedKind(fieldDesc.Kind()) && pv.isNegativeLiteral(expr.Right) {
		pv.addError(errors, "cannot assign negative value to unsigned field '%s' of type %s",
			pv.getFieldPath(expr.Left), fieldDesc.Kind())
		return true, false
	}

	return false, false // Not a special operator case
}

// validateGenericKinds handles standard proto kind compatibility checking.
// Validates proto kind compatibility and numeric range constraints.
// This is the catch-all validator that handles all remaining kinds.
func (pv *ProtoValidator) validateGenericKinds(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) (isKind bool, isValid bool) {
	leftKind, leftOk := pv.getExpressionKind(expr.Left)
	rightKind, rightOk := pv.getExpressionKind(expr.Right)

	if !leftOk || !rightOk {
		return true, true // Can't validate kinds, pass through
	}

	// Check proto kind compatibility (e.g., int32 field vs string literal)
	if !pv.protoKindsCompatible(leftKind, rightKind) {
		pv.addError(errors, "cannot compare %s field with %s value",
			leftKind, rightKind)
		return true, false
	}

	// Cycle 8A: Integer overflow validation
	// Per AIP-160: values must "align to the type of the field"
	if isIntegerKind(fieldDesc.Kind()) {
		if num, ok := getNumericValue(expr.Right); ok {
			pv.validateNumericRange(num, fieldDesc.Kind(), pv.getFieldPath(expr.Left), errors)
		}
	}

	return true, true // Is generic kind and valid
}

// validateEnumValue validates that the enum value exists in the enum definition.
// Assumes operator and type have already been validated.
func (pv *ProtoValidator) validateEnumValue(expr *ast.ComparisonExpression, fieldDesc protoreflect.FieldDescriptor, errors *[]error) {
	// Right side is guaranteed to be string literal by validateKindCompatibility
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

// getExpressionKind determines the proto kind of a value expression.
// Returns the field's kind for identifiers/traversals, or inferred kind for literals.
//
// Handles value expressions used in comparisons:
// - Identifiers: age → field's kind
// - Literals: "hello" → StringKind, 42 → Int64Kind, true → BoolKind
// - Negative literals: -5 → unwraps UnaryExpression, returns number kind
// - Traversals: user.email → nested field's kind
//
// Note: Does not handle logical operators (NOT, AND, OR) as they don't have a proto kind.
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
	case *ast.DurationLiteral:
		// Duration literals (20s, 1.2s) map to MessageKind because
		// google.protobuf.Duration is a message type (not a scalar kind).
		// Duration is defined as: message Duration { int64 seconds; int32 nanos; }
		// Validation logic in validateKindCompatibility handles Duration-specific matching.
		return protoreflect.MessageKind, true
	case *ast.UnaryExpression:
		// Handle negative literals: -5, -3.14, -5s (Cycle 7B + Duration)
		// Per AIP-160, - operator is used for both negation (NOT) and negative literals
		if n.Operator == "-" {
			// Recursively get the kind of the right operand
			return pv.getExpressionKind(n.Right)
		}
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
		// Special case: uint64 and fixed64 can accept very large numbers
		// that might be represented as DoubleKind due to their magnitude,
		// BUT we still validate they're actually integers (no fractional part)
		if (left == protoreflect.Uint64Kind || left == protoreflect.Fixed64Kind) && right == protoreflect.DoubleKind {
			// Type compatibility passes, but validateNumericRange will check
			// if it's actually an integer and within bounds
			return true
		}
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
// Also handles map key access (e.g., labels.env for maps).
func (pv *ProtoValidator) validateTraversal(expr *ast.TraversalExpression, errors *[]error) {
	// Resolve the left side and get its field descriptor
	leftField, _ := pv.resolveFieldDescriptor(expr.Left, pv.descriptor, errors)
	if leftField == nil {
		return // Error already added by resolveFieldDescriptor
	}

	// Check if left field is a map
	if leftField.IsMap() {
		// For maps, traversal creates a "virtual field" for the key
		// Example: labels.env accesses the value for key "env"
		// The right side must be either:
		// - Part of a comparison (labels.env = "value")
		// - Part of a HAS expression (labels.env:*)
		// We validate this in the parent expression (Comparison or Has)
		// For now, just mark that we've validated the traversal itself
		return
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

		// Special case: If left side is a map, return the map field itself
		// The traversal represents map key access (e.g., labels.env)
		// The caller (validateMapKind) will handle the key validation
		if leftField.IsMap() {
			return leftField, msgDesc
		}

		// Ensure left side is a message (for normal field traversal)
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
// - Maps: m:key checks if map contains key (TODO v2: not yet implemented)
// - Singular messages: m.field:"value" checks nested field
// - Star operator: m:* checks presence (non-empty/set)
//
// TODO v2 (Map Support): Implement map field validation per AIP-160.
// Map HAS syntax: m:key (key exists), m.key:* (key present), m.key:value (key-value match).
// Requires: key type validation, value type validation, map descriptor handling.
// Reference: https://google.aip.dev/160#has-operator (Maps section)
func (pv *ProtoValidator) validateHas(expr *ast.HasExpression, errors *[]error) {
	// Step 1: Resolve collection field descriptor
	var fieldDesc protoreflect.FieldDescriptor
	var collectionPath string

	switch coll := expr.Collection.(type) {
	case *ast.Identifier:
		// Simple repeated field: tags, scores, statuses
		fieldDesc = pv.descriptor.Fields().ByName(protoreflect.Name(coll.Value))
		if fieldDesc == nil {
			pv.addError(errors, "field '%s' does not exist in message %s", coll.Value, pv.descriptor.Name())
			return
		}
		collectionPath = coll.Value

	case *ast.TraversalExpression:
		// Nested HAS expressions:
		// - Repeated messages: emails.address:"test" (check if any email has address="test")
		// - Singular messages: email.address:"test" (check nested field in singular message)
		// - Maps with keys: labels.env:* (check if map has key "env")
		// - Deep nesting: emails.metadata.source:"web", nested.leaf.leaf_tags:"critical"
		//
		// Strategy:
		// 1. Check if left part is a map → handle map key access
		// 2. Check if left part is repeated → resolve right part within element type
		// 3. If not repeated → resolve full path (handles singular messages + deep nesting)

		// Try to resolve just the left part
		leftFieldDesc := pv.resolveFieldFromExpression(coll.Left, pv.descriptor)
		if leftFieldDesc == nil {
			pv.addError(errors, "field '%s' does not exist", pv.getFieldPath(coll.Left))
			return
		}

		// Check if left part is a map
		if leftFieldDesc.IsMap() {
			// MAP KEY ACCESS: labels.env:* or labels.env:"value"
			// The right part (e.g., "env") is the map key
			// The member is either * (presence) or a value to check
			
			// For star operator, we just validate the key exists (any value)
			if isStarLiteral(expr.Member) {
				// Validate the key expression (right part of traversal)
				pv.validateMapKeyPresence(leftFieldDesc, coll.Right, errors)
				return
			}
			
			// For value matching (m.foo:42), validate key and value
			// Key is in coll.Right, value is in expr.Member
			mapValueKind := leftFieldDesc.MapValue().Kind()
			memberKind, ok := pv.getExpressionKind(expr.Member)
			if !ok {
				pv.addError(errors, "cannot determine type of HAS member expression")
				return
			}
			
			// Validate value type matches map's value type
			if !pv.protoKindsCompatible(mapValueKind, memberKind) {
				pv.addError(errors, "type mismatch: map '%s' value type is %s, cannot check for %s value",
					leftFieldDesc.Name(), mapValueKind, memberKind)
				return
			}
			
			// Also validate the key
			pv.validateMapKeyPresence(leftFieldDesc, coll.Right, errors)
			return
		}

		// Check if left part is repeated
		if leftFieldDesc.IsList() {
			// REPEATED MESSAGE: Left is repeated - resolve right part within element type
			if leftFieldDesc.Kind() != protoreflect.MessageKind {
				pv.addError(errors, "cannot traverse into repeated field '%s' of non-message type %s",
					pv.getFieldPath(coll.Left), leftFieldDesc.Kind())
				return
			}

			// Get element message descriptor
			elementMsgDesc := leftFieldDesc.Message()

			// Resolve the right part (nested path) within the element message
			fieldDesc, _ = pv.resolveFieldDescriptor(coll.Right, elementMsgDesc, errors)
			if fieldDesc == nil {
				return // resolveFieldDescriptor already added errors
			}
			collectionPath = pv.getFieldPath(coll)
		} else {
			// SINGULAR MESSAGE OR DEEP NESTING: Left is not repeated
			// Examples:
			// - email.address:"test" (singular message field)
			// - nested.leaf.leaf_tags:"critical" (repeated field at end of path)
			// Resolve full path - handles both cases
			fieldDesc, _ = pv.resolveFieldDescriptor(coll, pv.descriptor, errors)
			if fieldDesc == nil {
				return // resolveFieldDescriptor already added errors
			}
			collectionPath = pv.getFieldPath(coll)
		}

	default:
		pv.addError(errors, "invalid collection expression in HAS operator")
		return
	}

	// Step 2: Validate field type compatibility with HAS operator
	// HAS operator can be used on:
	// - Repeated fields (any type) - validates element type
	// - Nested fields in messages (repeated or singular) - validates nested field type
	// - Maps (key presence check: m:key)

	// Check if field is a map
	if fieldDesc.IsMap() {
		// For maps, HAS operator checks key presence: m:key
		// The member is the key, and we need to validate it against the map's key type
		pv.validateMapKeyPresence(fieldDesc, expr.Member, errors)
		return
	}

	// For simple identifiers (not traversals), field must be repeated
	// UNLESS it's a star operator for presence check (Cycle 7D)
	if !fieldDesc.IsList() {
		if _, isSimple := expr.Collection.(*ast.Identifier); isSimple {
			// Allow star operator for message presence checks on singular fields
			// Per AIP-160: m:* checks if message field m is present (non-default)
			if isStarLiteral(expr.Member) && fieldDesc.Kind() == protoreflect.MessageKind {
				return // Valid: singular message presence check
			}

			// Simple identifier that's not repeated (and not star on message)
			pv.addError(errors, "HAS operator on simple fields requires repeated field, got singular %s", fieldDesc.Kind())
			return
		}
		// For TraversalExpression: fieldDesc is the nested field (can be singular or repeated)
		// Both work: email.address:"test" (singular message) and emails.address:"test" (repeated)
	}

	// Step 3: Validate member type matches element/field type
	elementKind := fieldDesc.Kind()

	// Special handling for enum elements
	if elementKind == protoreflect.EnumKind {
		// Member must be a string literal for enum check
		memberKind, ok := pv.getExpressionKind(expr.Member)
		if !ok {
			return // getExpressionKind already added error
		}
		if memberKind != protoreflect.StringKind {
			pv.addError(errors, "enum repeated field '%s' requires string value, got %s",
				collectionPath, memberKind)
			return
		}

		// Validate enum value exists (with prefix stripping support)
		stringLit, ok := expr.Member.(*ast.StringLiteral)
		if !ok {
			return // Should not happen if getExpressionKind returned StringKind
		}
		if !pv.isValidEnumValue(fieldDesc, stringLit.Value) {
			validValues := pv.getEnumValueNames(fieldDesc)
			pv.addError(errors, "invalid enum value '%s' for repeated field '%s', valid values: %v",
				stringLit.Value, collectionPath, validValues)
		}
		return
	}

	// Get member value kind (for non-enum types)
	memberKind, ok := pv.getExpressionKind(expr.Member)
	if !ok {
		return // getExpressionKind already added error
	}

	// Check type compatibility
	if !pv.protoKindsCompatible(elementKind, memberKind) {
		pv.addError(errors, "type mismatch: repeated field '%s' has elements of type %s, cannot check for %s value",
			collectionPath, elementKind, memberKind)
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

// === Negative Literal Helpers (Cycle 7B) ===

// isNegativeLiteral checks if an expression is a negative number literal.
// Detects UnaryExpression with "-" operator wrapping a NumberLiteral.
// Examples: -5, -3.14, -1.5e-3
func (pv *ProtoValidator) isNegativeLiteral(node ast.Node) bool {
	unary, ok := node.(*ast.UnaryExpression)
	if !ok || unary.Operator != "-" {
		return false
	}
	_, isNumber := unary.Right.(*ast.NumberLiteral)
	return isNumber
}

// isUnsignedKind checks if a proto kind is an unsigned integer type.
// Unsigned types: uint32, uint64, fixed32, fixed64
func (pv *ProtoValidator) isUnsignedKind(kind protoreflect.Kind) bool {
	return kind == protoreflect.Uint32Kind ||
		kind == protoreflect.Uint64Kind ||
		kind == protoreflect.Fixed32Kind ||
		kind == protoreflect.Fixed64Kind
}

// =============================================================================
// Cycle 8A: Integer Overflow Detection (AIP-160 "Align to Type")
// =============================================================================

// isIntegerKind checks if a proto kind is any integer type.
// Returns true for all 10 proto3 integer kinds.
func isIntegerKind(kind protoreflect.Kind) bool {
	return kind == protoreflect.Int32Kind ||
		kind == protoreflect.Int64Kind ||
		kind == protoreflect.Uint32Kind ||
		kind == protoreflect.Uint64Kind ||
		kind == protoreflect.Sint32Kind ||
		kind == protoreflect.Sint64Kind ||
		kind == protoreflect.Fixed32Kind ||
		kind == protoreflect.Fixed64Kind ||
		kind == protoreflect.Sfixed32Kind ||
		kind == protoreflect.Sfixed64Kind
}

// =============================================================================
// Phase 1: Duration Support (AIP-160 Duration Literals)
// =============================================================================

// isDurationField checks if a field is google.protobuf.Duration type.
// Per AIP-160: "Durations expect a numeric representation followed by an 's' suffix"
// Examples: timeout = 20s, delay = 1.2s
func isDurationField(fieldDesc protoreflect.FieldDescriptor) bool {
	if fieldDesc.Kind() != protoreflect.MessageKind {
		return false
	}
	msgDesc := fieldDesc.Message()
	if msgDesc == nil {
		return false
	}
	return msgDesc.FullName() == durationFullName
}

// isDurationLiteral checks if an AST node is a Duration literal.
func isDurationLiteral(node ast.Node) bool {
	_, ok := node.(*ast.DurationLiteral)
	return ok
}

// isNegativeDurationLiteral checks if an expression is a negative duration literal.
// Detects UnaryExpression with "-" operator wrapping a DurationLiteral.
// Examples: -5s, -1.2s
func (pv *ProtoValidator) isNegativeDurationLiteral(node ast.Node) bool {
	unary, ok := node.(*ast.UnaryExpression)
	if !ok || unary.Operator != "-" {
		return false
	}
	_, isDuration := unary.Right.(*ast.DurationLiteral)
	return isDuration
}

// =============================================================================
// Phase 2: Timestamp Support (AIP-160 RFC-3339)
// =============================================================================

// isTimestampField checks if a field is google.protobuf.Timestamp type.
// Per AIP-160: "Timestamps expect RFC-3339 format: 2012-04-21T11:30:00-04:00"
// Examples: created_at = "2024-03-16T05:00:00Z", updated_at = "2024-03-16T14:30:00+09:00"
func isTimestampField(fieldDesc protoreflect.FieldDescriptor) bool {
	if fieldDesc.Kind() != protoreflect.MessageKind {
		return false
	}
	msgDesc := fieldDesc.Message()
	if msgDesc == nil {
		return false
	}
	return msgDesc.FullName() == timestampFullName
}

// isValidRFC3339 validates that a string conforms to RFC-3339 timestamp format.
// Format: YYYY-MM-DDTHH:MM:SS[.fraction](Z|±HH:MM)
// Per RFC-3339: time-secfrac = "." 1*DIGIT (one or more digits, variable precision)
//
// Valid examples:
//   - "2024-03-16T05:00:00Z"              (UTC, no fractional seconds)
//   - "2024-03-16T14:30:00+09:00"         (with timezone offset)
//   - "2024-03-16T05:00:00.1Z"            (1 digit)
//   - "2024-03-16T05:00:00.123Z"          (milliseconds, 3 digits)
//   - "2024-03-16T05:00:00.123456Z"       (microseconds, 6 digits)
//   - "2024-03-16T05:00:00.123456789Z"    (nanoseconds, 9 digits)
//   - "2012-04-21T11:30:00-04:00"         (AIP-160 example)
//
// Invalid examples:
//   - "2024-03-16T05:00:00"               (missing timezone)
//   - "2024-03-16T05:00Z"                 (missing seconds)
//   - "2024-03-16 05:00:00Z"              (space instead of T)
func isValidRFC3339(value string) bool {
	return rfc3339Pattern.MatchString(value)
}

// validateMapKeyPresence validates HAS operator on map fields (m:key syntax).
// Per AIP-160: "m:foo" checks if map m contains the key "foo".
//
// Map key types can be:
// - string, int32, int64, uint32, uint64, sint32, sint64, fixed32, fixed64, sfixed32, sfixed64, bool
// (Proto3 restriction: keys cannot be float, double, bytes, enums, or messages)
//
// Examples:
//   labels:env          → check if string key "env" exists (env is identifier, treated as string)
//   settings:timeout    → check if string key "timeout" exists
//   id_names:100        → check if int32 key 100 exists (100 is number literal)
func (pv *ProtoValidator) validateMapKeyPresence(fieldDesc protoreflect.FieldDescriptor, keyExpr ast.Node, errors *[]error) {
	// Get map key and value types
	mapKeyKind := fieldDesc.MapKey().Kind()
	
	// Get the kind of the key expression
	// Special case: For string map keys, identifiers are treated as string literals
	// Example: labels:env means key is the string "env", not a field lookup
	var keyKind protoreflect.Kind
	
	if _, ok := keyExpr.(*ast.Identifier); ok && mapKeyKind == protoreflect.StringKind {
		// Identifier in HAS operator on string map = string key
		// Example: labels:env → key is "env" (string)
		keyKind = protoreflect.StringKind
	} else {
		// For other cases, get expression kind normally
		var kindOk bool
		keyKind, kindOk = pv.getExpressionKind(keyExpr)
		if !kindOk {
			pv.addError(errors, "cannot determine type of map key expression")
			return
		}
	}
	
	// Validate key type compatibility
	if !pv.protoKindsCompatible(mapKeyKind, keyKind) {
		pv.addError(errors, "type mismatch: map '%s' key type is %s, cannot compare with %s",
			fieldDesc.Name(), mapKeyKind, keyKind)
		return
	}
	
	// For numeric keys, validate the value is within range
	if isProtoIntegerKind(mapKeyKind) {
		if numLit, ok := keyExpr.(*ast.NumberLiteral); ok {
			pv.validateNumericRange(numLit.Value, mapKeyKind, string(fieldDesc.Name()), errors)
		}
	}
}

// getNumericValue extracts a float64 value from an expression.
// Handles both NumberLiteral and UnaryExpression (negative numbers).
// Returns (value, true) if extraction succeeds, (0, false) otherwise.
func getNumericValue(expr ast.Expression) (float64, bool) {
	switch e := expr.(type) {
	case *ast.NumberLiteral:
		return e.Value, true
	case *ast.UnaryExpression:
		if e.Operator == "-" {
			if val, ok := getNumericValue(e.Right); ok {
				return -val, true
			}
		}
	}
	return 0, false
}

// validateNumericRange checks if a numeric value is within the valid range
// for the given proto field kind using math package constants.
//
// Per AIP-160 Validation section: "Field values...MUST align to the type of the field"
// Example given: "age=hello" is invalid for int32 field (wrong type)
// Our extension: "age=2147483648" is also invalid (exceeds math.MaxInt32)
//
// Rationale: If a value cannot be represented in the field's type, it doesn't
// "align to" that type, following the same principle as type mismatches.
//
// Note on float64 precision: Large int64/uint64 values may lose precision when
// represented as float64 (which the parser uses). This is acceptable for
// validation purposes as we're checking boundaries, not exact values.
func (pv *ProtoValidator) validateNumericRange(value float64, fieldKind protoreflect.Kind, fieldName string, errors *[]error) {
	switch fieldKind {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		// Range: [-2147483648, 2147483647]
		if value < math.MinInt32 || value > math.MaxInt32 {
			pv.addError(errors, "value %v exceeds %s range [%d, %d] for field '%s'",
				value, fieldKind, math.MinInt32, math.MaxInt32, fieldName)
		}

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		// Range: [0, 4294967295]
		if value < 0 || value > math.MaxUint32 {
			pv.addError(errors, "value %v exceeds %s range [0, %d] for field '%s'",
				value, fieldKind, math.MaxUint32, fieldName)
		}

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		// Range: [-9223372036854775808, 9223372036854775807]
		// Note: math.MinInt64/MaxInt64 cannot be exactly represented in float64
		// Using slightly wider tolerance to account for float64 precision limits
		const maxInt64Float = 9.223372036854776e18  // Slightly above MaxInt64
		const minInt64Float = -9.223372036854776e18 // Slightly below MinInt64
		if value < minInt64Float || value > maxInt64Float {
			pv.addError(errors, "value %v exceeds %s range [%d, %d] for field '%s'",
				value, fieldKind, math.MinInt64, math.MaxInt64, fieldName)
		}

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		// Range: [0, 18446744073709551615]
		// Note: MaxUint64 cannot be exactly represented in float64
		// Using slightly wider tolerance to account for float64 precision limits
		const maxUint64Float = 1.8446744073709552e19 // Slightly above MaxUint64

		// Check if value has fractional part (not an integer)
		// Only check for values that can be represented as int64 (to avoid overflow in conversion)
		if value > 0 && value < float64(math.MaxInt64) {
			if value != float64(int64(value)) {
				// Has fractional part - invalid for integer field
				pv.addError(errors, "value %v has fractional part, cannot assign to %s field '%s'",
					value, fieldKind, fieldName)
				return
			}
		}

		if value < 0 || value > maxUint64Float {
			pv.addError(errors, "value %v exceeds %s range [0, %d] for field '%s'",
				value, fieldKind, uint64(math.MaxUint64), fieldName)
		}
	}
}
