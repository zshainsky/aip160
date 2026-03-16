package validator

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
)

// TestProtoValidator_Duration_Valid tests valid duration literal usage
// Per AIP-160: "Durations expect a numeric representation followed by an 's' suffix"
//
// TDD Cycle 1B - RED Phase: Should FAIL initially (duration type not handled in validator)
func TestProtoValidator_Duration_Valid(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name   string
		filter string
	}{
		{"integer duration", "timeout = 20s"},
		{"fractional duration", "timeout = 1.2s"},
		{"zero duration", "delay = 0s"},
		{"sub-second duration", "timeout = 0.5s"},
		{"large duration", "timeout = 3600s"},
		{"comparison less than", "timeout < 30s"},
		{"comparison greater than", "delay > 5s"},
		{"comparison less equal", "timeout <= 60s"},
		{"comparison greater equal", "delay >= 1s"},
		{"NOT equal", "timeout != 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errors := pv.Validate(program.Expression)

			if len(errors) > 0 {
				t.Errorf("Expected no validation errors for %q, got: %v", tt.filter, errors)
			}
		})
	}
}

// TestProtoValidator_Duration_TypeMismatch tests type mismatch detection
// Duration literals should only be valid on google.protobuf.Duration fields
//
// TDD Cycle 1B - RED Phase: Should FAIL initially
func TestProtoValidator_Duration_TypeMismatch(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name          string
		filter        string
		expectedError string
	}{
		{
			name:          "duration on string field",
			filter:        "name = 20s",
			expectedError: "type mismatch",
		},
		{
			name:          "duration on int32 field",
			filter:        "age = 20s",
			expectedError: "type mismatch",
		},
		{
			name:          "duration on bool field",
			filter:        "active = 20s",
			expectedError: "type mismatch",
		},
		{
			name:          "duration on float field",
			filter:        "score = 20s",
			expectedError: "type mismatch",
		},
		{
			name:          "duration on enum field",
			filter:        "task_status = 20s",
			expectedError: "type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errors := pv.Validate(program.Expression)

			if len(errors) == 0 {
				t.Errorf("Expected validation error for %q, got none", tt.filter)
			}

			// Check error message contains expected text
			found := false
			for _, err := range errors {
				errMsg := err.Error()
				if contains(errMsg, tt.expectedError) ||
					contains(errMsg, "duration literal cannot be used") ||
					contains(errMsg, "enum field") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing %q or duration/enum error, got: %v", tt.expectedError, errors)
			}
		})
	}
}

// TestProtoValidator_Duration_MissingSuffix tests that plain numbers are rejected on Duration fields
// This ensures we require the 's' suffix per AIP-160
//
// TDD Cycle 1B - RED Phase: Should FAIL initially
func TestProtoValidator_Duration_MissingSuffix(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name          string
		filter        string
		expectedError string
	}{
		{
			name:          "plain number on duration field",
			filter:        "timeout = 20",
			expectedError: "type mismatch",
		},
		{
			name:          "floating point on duration field",
			filter:        "delay = 1.5",
			expectedError: "type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errors := pv.Validate(program.Expression)

			if len(errors) == 0 {
				t.Errorf("Expected validation error for %q, got none", tt.filter)
			}

			// Check error message
			found := false
			for _, err := range errors {
				errMsg := err.Error()
				if contains(errMsg, tt.expectedError) ||
					contains(errMsg, "requires duration literal") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing %q or 'requires duration literal', got: %v", tt.expectedError, errors)
			}
		})
	}
}

// TestProtoValidator_Duration_Negative tests negative duration literals
func TestProtoValidator_Duration_Negative(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	filter := "timeout = -5s"

	l := lexer.New(filter)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	errors := pv.Validate(program.Expression)

	// Negative durations should be valid (parsed as unary - with duration literal)
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors for negative duration, got: %v", errors)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && hasSubstring(s, substr)))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
