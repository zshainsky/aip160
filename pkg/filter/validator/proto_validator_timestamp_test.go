package validator

import (
	"strings"
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
)

// =============================================================================
// Phase 2: Timestamp Support (AIP-160 RFC-3339)
// TDD Cycle: RED → GREEN → REFACTOR
// =============================================================================

// TestProtoValidator_Timestamp_Valid tests valid RFC-3339 timestamp literals.
// Per AIP-160: "Timestamps expect RFC-3339 format"
// Format: YYYY-MM-DDTHH:MM:SS[.fraction](Z|±HH:MM)
// Per RFC-3339: time-secfrac = "." 1*DIGIT (variable precision, 1+ digits)
func TestProtoValidator_Timestamp_Valid(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name   string
		filter string
	}{
		// Basic UTC timestamp
		{"UTC timestamp", `created_at = "2024-03-16T05:00:00Z"`},
		{"UTC with fractional", `created_at = "2024-03-16T05:00:00.123456Z"`},

		// Timezone offsets
		{"positive offset", `created_at = "2024-03-16T14:30:00+09:00"`},
		{"negative offset", `created_at = "2024-03-16T05:00:00-04:00"`},
		{"AIP-160 example", `created_at = "2012-04-21T11:30:00-04:00"`},

		// Comparison operators
		{"less than", `created_at < "2024-03-16T05:00:00Z"`},
		{"greater than", `created_at > "2024-03-16T05:00:00Z"`},
		{"less equal", `created_at <= "2024-03-16T05:00:00Z"`},
		{"greater equal", `created_at >= "2024-03-16T05:00:00Z"`},
		{"not equal", `created_at != "2024-03-16T05:00:00Z"`},

		// Multiple fields
		{"two timestamp fields", `created_at >= "2024-01-01T00:00:00Z" AND updated_at < "2024-12-31T23:59:59Z"`},

		// Logical operators
		{"OR operator", `created_at = "2024-03-16T05:00:00Z" OR updated_at = "2024-03-16T05:00:00Z"`},
		{"NOT operator", `NOT (created_at = "2024-03-16T05:00:00Z")`},
		{"NOT with minus", `-created_at = "2024-03-16T05:00:00Z"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errs := pv.Validate(program.Expression)
			if len(errs) > 0 {
				t.Errorf("Expected valid filter, got errors: %v", errs)
			}
		})
	}
}

// TestProtoValidator_Timestamp_InvalidFormat tests invalid RFC-3339 formats.
func TestProtoValidator_Timestamp_InvalidFormat(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name        string
		filter      string
		wantErrText string
	}{
		{"missing timezone", `created_at = "2024-03-16T05:00:00"`, "RFC-3339"},
		{"missing seconds", `created_at = "2024-03-16T05:00Z"`, "RFC-3339"},
		{"space instead of T", `created_at = "2024-03-16 05:00:00Z"`, "RFC-3339"},
		{"date only", `created_at = "2024-03-16"`, "RFC-3339"},
		{"time only", `created_at = "05:00:00Z"`, "RFC-3339"},
		{"malformed timezone", `created_at = "2024-03-16T05:00:00+9"`, "RFC-3339"},
		{"no T separator", `created_at = "20240316050000Z"`, "RFC-3339"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errs := pv.Validate(program.Expression)
			if len(errs) == 0 {
				t.Errorf("Expected validation error, got none")
				return
			}

			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), tt.wantErrText) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing '%s', got: %v", tt.wantErrText, errs)
			}
		})
	}
}

// TestProtoValidator_Timestamp_TypeMismatch tests type mismatches.
func TestProtoValidator_Timestamp_TypeMismatch(t *testing.T) {
	msgDesc := (&testdata.TestProtoData{}).ProtoReflect().Descriptor()
	pv := NewProtoValidator(msgDesc)

	tests := []struct {
		name        string
		filter      string
		wantErrText string
	}{
		{"number on timestamp", `created_at = 123`, "RFC-3339 string"},
		{"boolean on timestamp", `created_at = true`, "RFC-3339 string"},
		{"duration on timestamp", `created_at = 20s`, "RFC-3339 string"},
		{"non-RFC3339 string", `created_at = "hello world"`, "RFC-3339"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.filter)
			p := parser.New(l)
			program := p.ParseProgram()

			if len(p.Errors()) > 0 {
				t.Fatalf("Parser errors: %v", p.Errors())
			}

			errs := pv.Validate(program.Expression)
			if len(errs) == 0 {
				t.Errorf("Expected validation error, got none")
				return
			}

			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), tt.wantErrText) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error containing '%s', got: %v", tt.wantErrText, errs)
			}
		})
	}
}
