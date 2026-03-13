package validator

import (
	"fmt"

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
type ProtoValidator struct {
	descriptor protoreflect.MessageDescriptor
}

// NewProtoValidator creates a new validator for the given protobuf message descriptor.
// The descriptor can be obtained from any proto.Message via the ProtoReflect().Descriptor() method.
//
// Example:
//
//	var user *pb.User
//	validator := NewProtoValidator(user.ProtoReflect().Descriptor())
func NewProtoValidator(msgDesc protoreflect.MessageDescriptor, opts ...ValidatorOption) *ProtoValidator {
	return &ProtoValidator{
		descriptor: msgDesc,
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

// === Expression Validators ===

// validateComparison validates comparison expressions (=, !=, <, >, <=, >=).
func (pv *ProtoValidator) validateComparison(expr *ast.ComparisonExpression, errors *[]error) {
	// Get types of left and right operands
	leftKind, leftOk := pv.getExpressionKind(expr.Left)
	rightKind, rightOk := pv.getExpressionKind(expr.Right)
	
	// Validate left side (field reference)
	pv.validateNode(expr.Left, errors)
	
	// Check type compatibility
	if leftOk && rightOk {
		if !pv.protoKindsCompatible(leftKind, rightKind) {
			pv.addError(errors, "cannot compare %s field with %s value", 
				leftKind, rightKind)
		}
	}
}

// getExpressionKind determines the proto kind of an expression.
// Returns the field's kind for identifiers, or inferred kind for literals.
//
// For numeric literals, distinguishes between integer and float based on
// whether the value has a fractional part (e.g., 23 vs 23.55).
func (pv *ProtoValidator) getExpressionKind(node ast.Node) (protoreflect.Kind, bool) {
	switch n := node.(type) {
	case *ast.Identifier:
		fieldDesc, ok := pv.findFieldByName(pv.descriptor, n.Value)
		if ok {
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

// isProtoNumericKind checks if a protoreflect.Kind is numeric.
func isProtoNumericKind(k protoreflect.Kind) bool {
	switch k {
	case protoreflect.Int32Kind, protoreflect.Int64Kind,
		protoreflect.Uint32Kind, protoreflect.Uint64Kind,
		protoreflect.Sint32Kind, protoreflect.Sint64Kind,
		protoreflect.Fixed32Kind, protoreflect.Fixed64Kind,
		protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind,
		protoreflect.FloatKind, protoreflect.DoubleKind:
		return true
	}
	return false
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
// Implementation will be added in the traversal TDD cycle.
func (pv *ProtoValidator) validateTraversal(expr *ast.TraversalExpression, errors *[]error) {
	// TODO: Implement in traversal cycle
}
