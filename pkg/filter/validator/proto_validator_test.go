package validator

import (
	"strings"
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
	"google.golang.org/protobuf/reflect/protoreflect"
)

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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

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
	}

	msg := &testdata.User{}
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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// String types
		{"string field with string literal", `name = "John"`},
		{"bytes field with string literal", `profile_image = "data"`},
		
		// Boolean
		{"bool field with boolean literal", `active = true`},
		{"bool field with false literal", `active = false`},
		
		// Standard signed integers
		{"int32 field with integer literal", `age = 25`},
		{"int64 field with integer literal", `user_id = 12345`},
		
		// Unsigned integers
		{"uint32 field with integer literal", `points = 100`},
		{"uint64 field with integer literal", `balance = 99999`},
		
		// Signed integers (optimized for negatives)
		{"sint32 field with integer literal", `temperature = 72`},
		{"sint64 field with integer literal", `offset = 1000`},
		
		// Fixed-width unsigned integers
		{"fixed32 field with integer literal", `fixed_id = 12345`},
		{"fixed64 field with integer literal", `fixed_timestamp = 1234567890`},
		
		// Fixed-width signed integers
		{"sfixed32 field with integer literal", `sfixed_coord_x = 100`},
		{"sfixed64 field with integer literal", `sfixed_coord_y = 200`},
		
		// Floating point
		{"float field with float literal", `score = 3.14`},
		{"float field with integer literal", `score = 42`},
		{"double field with float literal", `rating = 4.5`},
		{"double field with integer literal", `rating = 5`},
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
	user := &testdata.User{}
	msgDesc := user.ProtoReflect().Descriptor()

	tests := []struct {
		name   string
		filter string
	}{
		// String type errors
		{"string field with number literal", `name = 123`},
		{"bytes field with number literal", `profile_image = 456`},
		
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
