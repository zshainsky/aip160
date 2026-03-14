package validator

import (
	"strings"
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TODO: Parser Limitation - Negative Literals Not Supported
// The current filter parser does not support negative number literals (e.g., -10, -3.14).
// When the parser is updated to support negative literals, uncomment the TODO test cases
// throughout this file to test:
//   - Signed integer types (int32, int64, sint32, sint64, sfixed32, sfixed64) accept negatives
//   - Unsigned integer types (uint32, uint64, fixed32, fixed64) reject negatives
//   - Float types (float, double) accept negative decimals
//   - Negative scientific notation (e.g., -1.5e-3)
//
// Current workaround: Use comparison operators (e.g., temperature < 0 instead of temperature = -10)

// validateProtoFilter is a helper function that parses a filter string and validates it
// against a proto message descriptor. Returns validation errors.
func validateProtoFilter(t *testing.T, filterStr string, msgDesc protoreflect.MessageDescriptor) []error {
	t.Helper()

	l := lexer.New(filterStr)
	p := parser.New(l)
	ast := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse filter '%s': %v", filterStr, p.Errors())
	}

	validator := NewProtoValidator(msgDesc)
	return validator.Validate(ast)
}

// TestProtoValidator_FieldExists tests that validation passes for existing fields
// 🔴 RED: This test should FAIL initially because validator is just a stub
func TestProtoValidator_FieldExists(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	// Test single field that exists
	errs := validateProtoFilter(t, `name = "John"`, msgDesc)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for existing field 'name', got: %v", errs)
	}

	// Test multiple existing fields
	errs = validateProtoFilter(t, `name = "John" AND age = 25`, msgDesc)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for existing fields, got: %v", errs)
	}

	// Test all scalar field types
	errs = validateProtoFilter(t, `email = "test@example.com"`, msgDesc)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for existing field 'email', got: %v", errs)
	}
}

// TestProtoValidator_FieldDoesNotExist tests that validation fails for non-existent fields
// 🔴 RED: This test should FAIL initially because validator returns empty errors
func TestProtoValidator_FieldDoesNotExist(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	// Test non-existent field
	errs := validateProtoFilter(t, `invalid_field = "test"`, msgDesc)
	if len(errs) == 0 {
		t.Error("Expected error for non-existent field 'invalid_field', got no errors")
		return
	}

	// Verify error message mentions the field
	errStr := errs[0].Error()
	if !strings.Contains(errStr, "invalid_field") {
		t.Errorf("Error should mention 'invalid_field', got: %s", errStr)
	}
}

// TestProtoValidator_MultipleFields tests validation with multiple field references
// 🔴 RED: This test should FAIL initially
func TestProtoValidator_MultipleFields(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	// All valid fields should pass
	errs := validateProtoFilter(t, `name = "John" AND age = 25 AND email = "test@example.com"`, msgDesc)
	if len(errs) > 0 {
		t.Errorf("Expected no errors for all valid fields, got: %v", errs)
	}

	// Mix of valid and invalid should report only invalid
	errs = validateProtoFilter(t, `name = "John" AND invalid_field = "test"`, msgDesc)
	if len(errs) == 0 {
		t.Error("Expected error for 'invalid_field'")
		return
	}
	if len(errs) != 1 {
		t.Errorf("Expected 1 error, got %d: %v", len(errs), errs)
	}
}

// TestProtoValidator_MultipleInvalidFields tests that all invalid fields are reported
// 🔴 RED: This test should FAIL initially
func TestProtoValidator_MultipleInvalidFields(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	// Multiple invalid fields should all be reported
	errs := validateProtoFilter(t, `invalid1 = "test" AND invalid2 = "test"`, msgDesc)
	if len(errs) < 2 {
		t.Errorf("Expected at least 2 errors for invalid fields, got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// TDD Cycle 2: Type Compatibility Tests
// ============================================================================

// TestProtoValidator_TypeCompatibility_ValidTypes tests valid type combinations
// Tests ALL 15 proto3 scalar types for comprehensive coverage
func TestProtoValidator_ScientificNotation(t *testing.T) {
	tests := []struct {
		name   string
		filter string
		valid  bool
	}{
		// Float fields accept scientific notation with or without fractional parts
		{"float with scientific notation", `score = 2.997e9`, true},
		{"float with negative exponent", `score = 1.5E-3`, true},
		{"float with integer scientific", `score = 3e10`, true},

		// Integer fields accept scientific notation that resolves to integer
		{"int32 with integer scientific", `age = 3e10`, true},

		// Integer fields reject scientific notation with fractional parts
		{"int32 with fractional scientific", `age = 1.5E-3`, false},
		{"uint64 with fractional scientific", `balance = 1.5E-3`, false},

		// TODO: Add negative scientific notation tests when parser supports negative literals
		// {"float with negative scientific", `score = -2.997e9`, true},
		// {"int32 with negative scientific", `temperature = -1.5e2`, false}, // Has fractional part
		// {"sint32 with negative integer scientific", `temperature = -3e2`, true}, // -300 is valid
	}

	msg := &testdata.TestProtoData{}
	msgDesc := msg.ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			err := pv.Validate(program)

			if tt.valid {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			}
		})
	}
}

func TestProtoValidator_TypeCompatibility_ValidTypes(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// String types
		{"string field with string literal", `name = "John"`},
		{"bytes field with string literal", `data = "data"`},

		// Boolean
		{"bool field with boolean literal", `active = true`},
		{"bool field with false literal", `active = false`},

		// Standard signed integers
		{"int32 field with integer literal", `age = 25`},
		{"int64 field with integer literal", `user_id = 12345`},
		// TODO: Add negative literal tests when parser supports negative numbers (e.g., age = -10)
		// Currently parser doesn't support negative literals directly
		// {"int32 field with negative literal", `age = -10`},
		// {"int64 field with negative literal", `user_id = -999`},

		// Unsigned integers
		{"uint32 field with integer literal", `points = 100`},
		{"uint64 field with integer literal", `balance = 99999`},
		// TODO: Add test to reject negative literals on unsigned fields when parser supports them
		// {"uint32 field with negative literal", `points = -100`}, // Should fail
		// {"uint64 field with negative literal", `balance = -999`}, // Should fail

		// Signed integers (optimized for negatives)
		{"sint32 field with integer literal", `temperature = 72`},
		{"sint64 field with integer literal", `offset = 1000`},
		// TODO: Add negative literal tests when parser supports them
		// sint32/sint64 are specifically optimized for negative values via ZigZag encoding
		// {"sint32 field with negative literal", `temperature = -40`},
		// {"sint64 field with negative literal", `offset = -12345`},

		// Fixed-width unsigned integers
		{"fixed32 field with integer literal", `fixed_id = 12345`},
		{"fixed64 field with integer literal", `fixed_timestamp = 1234567890`},
		// TODO: Add test to reject negative literals when parser supports them
		// {"fixed32 field with negative literal", `fixed_id = -100`}, // Should fail

		// Fixed-width signed integers
		{"sfixed32 field with integer literal", `sfixed_coord_x = 100`},
		{"sfixed64 field with integer literal", `sfixed_coord_y = 200`},
		// TODO: Add negative literal tests when parser supports them
		// {"sfixed32 field with negative literal", `sfixed_coord_x = -100`},
		// {"sfixed64 field with negative literal", `sfixed_coord_y = -200`},

		// Floating point
		{"float field with float literal", `score = 3.14`},
		{"float field with integer literal", `score = 42`},
		{"double field with float literal", `rating = 4.5`},
		{"double field with integer literal", `rating = 5`},
		// TODO: Add negative literal tests when parser supports them
		// {"float field with negative literal", `score = -3.14`},
		// {"double field with negative literal", `rating = -4.5`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) > 0 {
				t.Errorf("Expected no errors for %s, got: %v", tt.name, errs)
			}
		})
	}
}

// TestProtoValidator_TypeCompatibility_InvalidTypes tests invalid type combinations
// Tests rejection of incompatible literals for all integer types
func TestProtoValidator_TypeCompatibility_InvalidTypes(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// String type errors
		{"string field with number literal", `name = 123`},
		{"bytes field with number literal", `data = 456`},

		// Boolean type errors
		{"bool field with string literal", `active = "true"`},
		{"bool field with number literal", `active = 1`},

		// Standard signed integer errors
		{"int32 field with string literal", `age = "twenty five"`},
		{"int32 field with boolean literal", `age = true`},
		{"int32 field with float literal", `age = 23.55`},
		{"int64 field with string literal", `user_id = "12345"`},
		{"int64 field with float literal", `user_id = 123.45`},

		// Unsigned integer errors
		{"uint32 field with string literal", `points = "100"`},
		{"uint32 field with float literal", `points = 10.5`},
		{"uint32 field with boolean literal", `points = false`},
		{"uint64 field with string literal", `balance = "99999"`},
		{"uint64 field with float literal", `balance = 999.99`},
		// TODO: Add tests for negative literals on unsigned fields when parser supports them
		// These should properly fail validation (can't assign negative to unsigned)
		// {"uint32 field with negative literal", `points = -100`},
		// {"uint64 field with negative literal", `balance = -999`},

		// Signed integer (optimized) errors
		{"sint32 field with string literal", `temperature = "cold"`},
		{"sint32 field with boolean literal", `temperature = true`},
		{"sint64 field with string literal", `offset = "large"`},
		{"sint64 field with float literal", `offset = 100.25`},

		// Fixed-width unsigned integer errors
		{"fixed32 field with string literal", `fixed_id = "abc"`},
		{"fixed32 field with float literal", `fixed_id = 123.45`},
		{"fixed64 field with string literal", `fixed_timestamp = "now"`},
		{"fixed64 field with float literal", `fixed_timestamp = 123.456`},

		// Fixed-width signed integer errors
		{"sfixed32 field with string literal", `sfixed_coord_x = "left"`},
		{"sfixed32 field with float literal", `sfixed_coord_x = 10.5`},
		{"sfixed64 field with string literal", `sfixed_coord_y = "top"`},
		{"sfixed64 field with float literal", `sfixed_coord_y = 20.5`},

		// Float type errors (floats accept integers but not strings/bools)
		{"float field with string literal", `score = "3.14"`},
		{"float field with boolean literal", `score = true`},
		{"double field with string literal", `rating = "4.5"`},
		{"double field with boolean literal", `rating = false`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) == 0 {
				t.Errorf("Expected type error for %s, got no errors", tt.name)
			}
		})
	}
}

// ============================================================================
// TDD Cycle 3: Enum Validation Tests (RED Phase)
// ============================================================================

// TestProtoValidator_EnumValidation_ValidValues tests that enum fields
// accept valid enum value names as strings
func TestProtoValidator_EnumValidation_ValidValues(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// Valid enum values (string representation)
		{"enum with ACTIVE value", `task_status = "TASK_STATUS_ACTIVE"`},
		{"enum with INACTIVE value", `task_status = "TASK_STATUS_INACTIVE"`},
		{"enum with PENDING value", `task_status = "TASK_STATUS_COMPLETED"`},
		{"enum with UNSPECIFIED value", `task_status = "TASK_STATUS_UNSPECIFIED"`},

		// Valid enum values (prefix-stripped - should match with prefix added)
		{"enum with ACTIVE value no prefix", `task_status = "ACTIVE"`},
		{"enum with INACTIVE value no prefix", `task_status = "INACTIVE"`},
		{"enum with COMPLETED value no prefix", `task_status = "COMPLETED"`},
		{"enum with UNSPECIFIED value no prefix", `task_status = "UNSPECIFIED"`},

		// Equality and inequality operators (only ones allowed for enums)
		{"enum with equality operator", `task_status = "TASK_STATUS_ACTIVE"`},
		{"enum with inequality operator", `task_status != "TASK_STATUS_INACTIVE"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) > 0 {
				t.Errorf("Expected no errors for %s, got: %v", tt.name, errs)
			}
		})
	}
}

// TestProtoValidator_EnumValidation_InvalidValues tests that enum fields
// reject invalid enum values
func TestProtoValidator_EnumValidation_InvalidValues(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// Invalid enum values (not in enum definition)
		{"enum with invalid value", `task_status = "INVALID_VALUE"`},
		{"enum with wrong case", `task_status = "status_active"`},
		{"enum with numeric value", `task_status = 1`},
		{"enum with boolean", `task_status = true`},

		// Invalid operators for enum (only = and != allowed)
		{"enum with less than", `status < "TASK_STATUS_ACTIVE"`},
		{"enum with greater than", `status > "TASK_STATUS_ACTIVE"`},
		{"enum with less or equal", `status <= "TASK_STATUS_COMPLETED"`},
		{"enum with greater or equal", `status >= "TASK_STATUS_COMPLETED"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) == 0 {
				t.Errorf("Expected validation error for %s, got none", tt.name)
			}
		})
	}
}

// TestProtoValidator_EnumValidation_NonPrefixedEnum_ValidValues tests valid enum values
// for TaskResult enum which has non-prefixed values (SUCCESS, FAILED, PENDING).
// Per proto3 spec, the prefix pattern is recommended but not required.
func TestProtoValidator_EnumValidation_NonPrefixedEnum_ValidValues(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// Valid non-prefixed enum values
		{"non-prefixed SUCCESS", `task_result = "SUCCESS"`},
		{"non-prefixed FAILED", `task_result = "FAILED"`},
		{"non-prefixed PENDING", `task_result = "PENDING"`},

		// TASK_RESULT_UNSPECIFIED is actually prefixed (zero value conventionally has prefix)
		{"prefixed TASK_RESULT_UNSPECIFIED exists", `task_result = "TASK_RESULT_UNSPECIFIED"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) > 0 {
				t.Errorf("Expected no errors, got: %v", errs)
			}
		})
	}
}

// TestProtoValidator_EnumValidation_NonPrefixedEnum_InvalidValues tests invalid enum values
// for TaskResult enum to ensure proper validation and that we don't incorrectly add prefixes.
func TestProtoValidator_EnumValidation_NonPrefixedEnum_InvalidValues(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// Case sensitivity still enforced (per AIP-160: case-sensitive)
		{"lowercase rejected", `task_result = "success"`},
		{"mixed case rejected", `task_result = "Success"`},
		{"wrong case failed", `task_result = "failed"`},
		
		// Prefixed versions don't exist for non-zero values (verify we don't add prefixes where they don't exist)
		{"prefixed TASK_RESULT_SUCCESS doesn't exist", `task_result = "TASK_RESULT_SUCCESS"`},
		{"prefixed TASK_RESULT_FAILED doesn't exist", `task_result = "TASK_RESULT_FAILED"`},
		{"prefixed TASK_RESULT_PENDING doesn't exist", `task_result = "TASK_RESULT_PENDING"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) == 0 {
				t.Errorf("Expected validation error, got none")
			}
		})
	}
}

// === Nested Traversal Tests (TDD Cycle 5) ===

// TestProtoValidator_NestedTraversal_Valid tests valid nested field access.
// Tests one-level and three-level deep nesting.
func TestProtoValidator_NestedTraversal_Valid(t *testing.T) {
	contact := &testdata.Contact{}
	msgDesc := contact.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		{"one level string", `email.address = "test@example.com"`},
		{"one level bool", `email.verified = true`},
		{"three level deep", `email.metadata.source = "web"`},
		{"three level deep int", `email.metadata.priority = 5`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) > 0 {
				t.Errorf("Expected no errors for valid nested field '%s', got: %v", tt.filter, errs)
			}
		})
	}
}

// TestProtoValidator_NestedTraversal_InvalidField tests traversal with non-existent nested fields.
func TestProtoValidator_NestedTraversal_InvalidField(t *testing.T) {
	contact := &testdata.Contact{}
	msgDesc := contact.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		{"nested field doesn't exist", `email.invalid = "test"`},
		{"top level doesn't exist", `nonexistent.field = "test"`},
		{"three level deep invalid", `email.metadata.invalid = "test"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) == 0 {
				t.Errorf("Expected validation error for '%s', got none", tt.filter)
			}
		})
	}
}

// TestProtoValidator_NestedTraversal_NonMessageType tests that traversal is rejected for scalar fields.
// In protobuf, only message-type fields have nested structure. Scalar types (string, bool, int, etc.)
// are primitive values and cannot be traversed with dot notation.
//
// Examples:
//   - VALID:   email.address         (email is a Message with an 'address' field)
//   - INVALID: name.something         (name is a string - strings have no sub-fields)
//   - INVALID: email.verified.nope    (verified is a bool - cannot traverse further)
//
// This test ensures the validator catches these semantic errors and provides clear error messages
// explaining that you cannot traverse into non-message fields.
func TestProtoValidator_NestedTraversal_NonMessageType(t *testing.T) {
	contact := &testdata.Contact{}
	msgDesc := contact.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		{"traverse into string", `name.invalid = "test"`},
		{"traverse into bool", `email.verified.nope = "test"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if len(errs) == 0 {
				t.Errorf("Expected validation error for non-message traversal '%s', got none", tt.filter)
			}
		})
	}
}

// TestProtoValidator_NestedTraversal_TypeValidation tests that type checking works on nested fields.
// The validator should enforce type compatibility for nested field comparisons, just like
// it does for top-level fields.
//
// Examples:
//   - email.address (string) can be compared with string literals
//   - email.verified (bool) can be compared with boolean literals
//   - email.metadata.priority (int32) can be compared with integer literals
//   - Type mismatches should be rejected (e.g., comparing string field with number)
func TestProtoValidator_NestedTraversal_TypeValidation(t *testing.T) {
	contact := &testdata.Contact{}
	msgDesc := contact.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
		valid  bool
	}{
		// Valid type matches
		{"string with string", `email.address = "test@example.com"`, true},
		{"bool with bool", `email.verified = true`, true},
		{"int with int", `email.metadata.priority = 5`, true},
		{"three level string", `email.metadata.source = "web"`, true},
		
		// Invalid type mismatches
		{"string with number", `email.address = 123`, false},
		{"bool with string", `email.verified = "true"`, false},
		{"int with string", `email.metadata.priority = "high"`, false},
		{"string with bool", `email.metadata.source = true`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.valid && len(errs) > 0 {
				t.Errorf("Expected no errors for valid type match '%s', got: %v", tt.filter, errs)
			}
			if !tt.valid && len(errs) == 0 {
				t.Errorf("Expected type mismatch error for '%s', got none", tt.filter)
			}
		})
	}
}
