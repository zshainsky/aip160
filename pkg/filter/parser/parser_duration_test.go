package parser

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// TestDurationLiteral tests parsing of duration literals per AIP-160
// AIP-160 specifies: "Durations expect a numeric representation followed by an 's' suffix"
// Examples: 20s (20 seconds), 1.2s (1.2 seconds)
//
// TDD Cycle 1A - RED Phase: These tests should FAIL initially
// The parser needs to recognize duration syntax and create appropriate AST nodes
func TestDurationLiteral(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedValue  float64
		expectedUnit   string
	}{
		{
			name:          "integer duration",
			input:         "20s",
			expectedValue: 20.0,
			expectedUnit:  "s",
		},
		{
			name:          "fractional duration",
			input:         "1.2s",
			expectedValue: 1.2,
			expectedUnit:  "s",
		},
		{
			name:          "zero duration",
			input:         "0s",
			expectedValue: 0.0,
			expectedUnit:  "s",
		},
		{
			name:          "sub-second duration",
			input:         "0.5s",
			expectedValue: 0.5,
			expectedUnit:  "s",
		},
		{
			name:          "large duration",
			input:         "3600s",
			expectedValue: 3600.0,
			expectedUnit:  "s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if program == nil {
				t.Fatal("ParseProgram() returned nil program")
			}

			if program.Expression == nil {
				t.Fatal("program.Expression is nil")
			}

			// Expect a DurationLiteral node
			dur, ok := program.Expression.(*ast.DurationLiteral)
			if !ok {
				t.Fatalf("expression not *ast.DurationLiteral. got=%T", program.Expression)
			}

			if dur.Value != tt.expectedValue {
				t.Errorf("duration.Value = %f, want %f", dur.Value, tt.expectedValue)
			}

			if dur.Unit != tt.expectedUnit {
				t.Errorf("duration.Unit = %q, want %q", dur.Unit, tt.expectedUnit)
			}
		})
	}
}

// TestDurationInComparison tests duration literals in comparison expressions
// AIP-160: timeout = 20s, timeout < 30s
//
// TDD Cycle 1A - RED Phase: These should FAIL initially
func TestDurationInComparison(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedField string
		expectedOp    string
		expectedValue float64
	}{
		{
			name:          "equality with duration",
			input:         "timeout = 20s",
			expectedField: "timeout",
			expectedOp:    "=",
			expectedValue: 20.0,
		},
		{
			name:          "less than with duration",
			input:         "timeout < 30s",
			expectedField: "timeout",
			expectedOp:    "<",
			expectedValue: 30.0,
		},
		{
			name:          "greater than with duration",
			input:         "delay > 5s",
			expectedField: "delay",
			expectedOp:    ">",
			expectedValue: 5.0,
		},
		{
			name:          "fractional duration in comparison",
			input:         "timeout = 1.5s",
			expectedField: "timeout",
			expectedOp:    "=",
			expectedValue: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			comp, ok := program.Expression.(*ast.ComparisonExpression)
			if !ok {
				t.Fatalf("expression not *ast.ComparisonExpression. got=%T", program.Expression)
			}

			// Check field name
			ident, ok := comp.Left.(*ast.Identifier)
			if !ok {
				t.Fatalf("comp.Left not *ast.Identifier. got=%T", comp.Left)
			}
			if ident.Value != tt.expectedField {
				t.Errorf("field = %q, want %q", ident.Value, tt.expectedField)
			}

			// Check operator
			if comp.Operator != tt.expectedOp {
				t.Errorf("operator = %q, want %q", comp.Operator, tt.expectedOp)
			}

			// Check duration literal on right side
			dur, ok := comp.Right.(*ast.DurationLiteral)
			if !ok {
				t.Fatalf("comp.Right not *ast.DurationLiteral. got=%T", comp.Right)
			}

			if dur.Value != tt.expectedValue {
				t.Errorf("duration.Value = %f, want %f", dur.Value, tt.expectedValue)
			}

			if dur.Unit != "s" {
				t.Errorf("duration.Unit = %q, want %q", dur.Unit, "s")
			}
		})
	}
}

// TestNegativeDuration tests negative duration literals
// Example: -5s (negative 5 seconds)
//
// TDD Cycle 1A - RED Phase: Should FAIL initially
func TestNegativeDuration(t *testing.T) {
	input := "timeout = -5s"

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	comp, ok := program.Expression.(*ast.ComparisonExpression)
	if !ok {
		t.Fatalf("expression not *ast.ComparisonExpression. got=%T", program.Expression)
	}

	// Negative duration should be parsed as UnaryExpression("-") with DurationLiteral
	unary, ok := comp.Right.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("comp.Right not *ast.UnaryExpression. got=%T", comp.Right)
	}

	if unary.Operator != "-" {
		t.Errorf("unary.Operator = %q, want %q", unary.Operator, "-")
	}

	dur, ok := unary.Right.(*ast.DurationLiteral)
	if !ok {
		t.Fatalf("unary.Right not *ast.DurationLiteral. got=%T", unary.Right)
	}

	if dur.Value != 5.0 {
		t.Errorf("duration.Value = %f, want %f", dur.Value, 5.0)
	}
}

// TestDurationVsNumber tests that we can distinguish durations from plain numbers
// Plain number: 20 (no suffix)
// Duration: 20s (with 's' suffix)
//
// This ensures the parser correctly identifies duration literals
func TestDurationVsNumber(t *testing.T) {
	t.Run("plain number", func(t *testing.T) {
		input := "age = 20"
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		comp := program.Expression.(*ast.ComparisonExpression)
		
		// Should be NumberLiteral, not DurationLiteral
		_, ok := comp.Right.(*ast.NumberLiteral)
		if !ok {
			t.Errorf("expected NumberLiteral for '20', got %T", comp.Right)
		}
	})

	t.Run("duration with suffix", func(t *testing.T) {
		input := "timeout = 20s"
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		comp := program.Expression.(*ast.ComparisonExpression)
		
		// Should be DurationLiteral, not NumberLiteral
		_, ok := comp.Right.(*ast.DurationLiteral)
		if !ok {
			t.Errorf("expected DurationLiteral for '20s', got %T", comp.Right)
		}
	})
}
