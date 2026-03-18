package validator

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
)

// parseFilter is a helper to parse filter strings for tests.
func parseFilter(t *testing.T, filterStr string) *ast.Program {
	t.Helper()
	l := lexer.New(filterStr)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse filter '%s': %v", filterStr, p.Errors())
	}
	return program
}

// TestEnumPrefixStripping_DefaultEnabled tests enum validation with prefix stripping enabled (default behavior).
// Tests both exact matches and prefix-stripped values with the TaskStatus enum.
func TestEnumPrefixStripping_DefaultEnabled(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	validator := NewProtoValidator(msgDesc) // Default: EnableEnumPrefixStripping = true

	tests := []struct {
		name      string
		filter    string
		wantValid bool
		desc      string
	}{
		{
			name:      "exact match with full prefix",
			filter:    "task_status = \"TASK_STATUS_ACTIVE\"",
			wantValid: true,
			desc:      "Exact match should work with prefix stripping enabled",
		},
		{
			name:      "prefix stripped ACTIVE",
			filter:    "task_status = \"ACTIVE\"",
			wantValid: true,
			desc:      "Prefix-stripped should work when enabled",
		},
		{
			name:      "invalid value without prefix",
			filter:    "task_status = \"INVALID\"",
			wantValid: false,
			desc:      "Invalid prefix-stripped value should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseFilter(t, tt.filter)
			errs := validator.Validate(expr)
			isValid := len(errs) == 0

			if isValid != tt.wantValid {
				if tt.wantValid {
					t.Errorf("Expected valid filter, got errors: %v\n%s", errs, tt.desc)
				} else {
					t.Errorf("Expected validation errors, but filter was valid\n%s", tt.desc)
				}
			}
		})
	}
}

// TestEnumPrefixStripping_Disabled tests enum validation with prefix stripping disabled.
func TestEnumPrefixStripping_Disabled(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	validator := NewProtoValidator(msgDesc, WithEnumPrefixStripping(false))

	tests := []struct {
		name      string
		filter    string
		wantValid bool
		desc      string
	}{
		{
			name:      "exact match ACTIVE",
			filter:    "task_status = \"TASK_STATUS_ACTIVE\"",
			wantValid: true,
			desc:      "Exact match should work even with prefix stripping disabled",
		},
		{
			name:      "prefix stripped ACTIVE should fail",
			filter:    "task_status = \"ACTIVE\"",
			wantValid: false,
			desc:      "Prefix-stripped should fail when disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := parseFilter(t, tt.filter)
			errs := validator.Validate(expr)
			isValid := len(errs) == 0

			if isValid != tt.wantValid {
				if tt.wantValid {
					t.Errorf("Expected valid filter, got errors: %v\n%s", errs, tt.desc)
				} else {
					t.Errorf("Expected validation errors, but filter was valid\n%s", tt.desc)
				}
			}
		})
	}
}
