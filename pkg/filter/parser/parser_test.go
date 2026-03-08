package parser

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// Helper function to check parser errors
func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

// TestParserCreation tests that the parser is properly initialized
func TestParserCreation(t *testing.T) {
	input := `age > 18`

	l := lexer.New(input)
	p := New(l)

	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.currentToken.Type != lexer.IDENTIFIER {
		t.Errorf("currentToken wrong. expected IDENTIFIER, got=%q", p.currentToken.Type)
	}

	if p.peekToken.Type != lexer.GREATER_THAN {
		t.Errorf("peekToken wrong. expected GREATER_THAN, got=%q", p.peekToken.Type)
	}
}

// TestStringLiteral tests parsing of string literals
func TestStringLiteral(t *testing.T) {
	input := `"hello world"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program.Expression == nil {
		t.Fatal("program.Expression is nil")
	}

	str, ok := program.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expression not *ast.StringLiteral. got=%T", program.Expression)
	}

	if str.Value != "hello world" {
		t.Errorf("str.Value wrong. expected=%q, got=%q", "hello world", str.Value)
	}
}

// TestNumberLiteral tests parsing of numeric literals
func TestNumberLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"0", 0},
		{"2.998e8", 2.998e8},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		num, ok := program.Expression.(*ast.NumberLiteral)
		if !ok {
			t.Fatalf("expression not *ast.NumberLiteral. got=%T", program.Expression)
		}

		if num.Value != tt.expected {
			t.Errorf("num.Value wrong. expected=%f, got=%f", tt.expected, num.Value)
		}
	}
}

// TestBooleanLiteral tests parsing of boolean literals
func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"TRUE", true},
		{"FALSE", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		b, ok := program.Expression.(*ast.BooleanLiteral)
		if !ok {
			t.Fatalf("expression not *ast.BooleanLiteral. got=%T", program.Expression)
		}

		if b.Value != tt.expected {
			t.Errorf("b.Value wrong for input %q. expected=%v, got=%v", tt.input, tt.expected, b.Value)
		}
	}
}

// TestNullLiteral tests parsing of null literals
func TestNullLiteral(t *testing.T) {
	input := `null`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	_, ok := program.Expression.(*ast.NullLiteral)
	if !ok {
		t.Fatalf("expression not *ast.NullLiteral. got=%T", program.Expression)
	}
}

// TestIdentifier tests parsing of simple identifiers
func TestIdentifier(t *testing.T) {
	input := `age`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	ident, ok := program.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression not *ast.Identifier. got=%T", program.Expression)
	}

	if ident.Value != "age" {
		t.Errorf("ident.Value wrong. expected=%q, got=%q", "age", ident.Value)
	}
}

// TestSimpleComparison tests parsing of comparison expressions
func TestSimpleComparison(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		left     string
		right    string
	}{
		{"age > 18", ">", "age", "18"},
		{"name = \"John\"", "=", "name", "\"John\""},
		{"active != false", "!=", "active", "false"},
		{"score < 100", "<", "score", "100"},
		{"rating >= 4.5", ">=", "rating", "4.5"},
		{"count <= 10", "<=", "count", "10"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		comp, ok := program.Expression.(*ast.ComparisonExpression)
		if !ok {
			t.Fatalf("expression not *ast.ComparisonExpression for %q. got=%T", tt.input, program.Expression)
		}

		if comp.Operator != tt.operator {
			t.Errorf("comp.Operator wrong for %q. expected=%q, got=%q", tt.input, tt.operator, comp.Operator)
		}

		if comp.Left.String() != tt.left {
			t.Errorf("comp.Left wrong for %q. expected=%q, got=%q", tt.input, tt.left, comp.Left.String())
		}

		if comp.Right.String() != tt.right {
			t.Errorf("comp.Right wrong for %q. expected=%q, got=%q", tt.input, tt.right, comp.Right.String())
		}
	}
}

// TestLogicalAnd tests parsing of AND expressions
func TestLogicalAnd(t *testing.T) {
	input := `age > 18 AND active = true`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	logical, ok := program.Expression.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expression not *ast.LogicalExpression. got=%T", program.Expression)
	}

	if logical.Operator != "AND" {
		t.Errorf("logical.Operator wrong. expected=%q, got=%q", "AND", logical.Operator)
	}

	// Check left side is a comparison
	_, ok = logical.Left.(*ast.ComparisonExpression)
	if !ok {
		t.Errorf("logical.Left not ComparisonExpression. got=%T", logical.Left)
	}

	// Check right side is a comparison
	_, ok = logical.Right.(*ast.ComparisonExpression)
	if !ok {
		t.Errorf("logical.Right not ComparisonExpression. got=%T", logical.Right)
	}
}

// TestLogicalOr tests parsing of OR expressions
func TestLogicalOr(t *testing.T) {
	input := `premium = true OR trial = true`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	logical, ok := program.Expression.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expression not *ast.LogicalExpression. got=%T", program.Expression)
	}

	if logical.Operator != "OR" {
		t.Errorf("logical.Operator wrong. expected=%q, got=%q", "OR", logical.Operator)
	}
}

// TestNotExpression tests parsing of NOT expressions
func TestNotExpression(t *testing.T) {
	input := `NOT active`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	unary, ok := program.Expression.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("expression not *ast.UnaryExpression. got=%T", program.Expression)
	}

	if unary.Operator != "NOT" {
		t.Errorf("unary.Operator wrong. expected=%q, got=%q", "NOT", unary.Operator)
	}

	_, ok = unary.Right.(*ast.Identifier)
	if !ok {
		t.Errorf("unary.Right not Identifier. got=%T", unary.Right)
	}
}

// TestOperatorPrecedence tests that operator precedence is correct
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// OR has lowest precedence
		{"a OR b AND c", "(a OR (b AND c))"},

		// AND binds tighter than OR
		{"a AND b OR c", "((a AND b) OR c)"},

		// NOT has highest precedence
		{"NOT a AND b", "((NOT a) AND b)"},

		// Multiple ORs
		{"a OR b OR c", "((a OR b) OR c)"},

		// Multiple ANDs
		{"a AND b AND c", "((a AND b) AND c)"},

		// Parentheses override precedence
		{"(a OR b) AND c", "((a OR b) AND c)"},

		// Complex precedence
		{"a OR b AND c OR d", "((a OR (b AND c)) OR d)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("precedence wrong for %q.\nexpected: %q\ngot:      %q", tt.input, tt.expected, actual)
		}
	}
}

// TestGroupedExpression tests parenthesized expressions
func TestGroupedExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(age > 18)", "(age > 18)"},
		{"((age > 18))", "(age > 18)"},
		{"(a OR b) AND c", "((a OR b) AND c)"},
		{"a AND (b OR c)", "(a AND (b OR c))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("grouped expression wrong for %q.\nexpected: %q\ngot:      %q", tt.input, tt.expected, actual)
		}
	}
}

// TestComplexExpression tests a complex real-world filter
func TestComplexExpression(t *testing.T) {
	input := `age >= 21 AND (status = "premium" OR trial = true)`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should be a LogicalExpression with AND
	logical, ok := program.Expression.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expression not *ast.LogicalExpression. got=%T", program.Expression)
	}

	if logical.Operator != "AND" {
		t.Errorf("top-level operator wrong. expected AND, got %q", logical.Operator)
	}

	// Left should be a comparison
	_, ok = logical.Left.(*ast.ComparisonExpression)
	if !ok {
		t.Errorf("left not ComparisonExpression. got=%T", logical.Left)
	}

	// Right should be an OR expression
	rightLogical, ok := logical.Right.(*ast.LogicalExpression)
	if !ok {
		t.Errorf("right not LogicalExpression. got=%T", logical.Right)
	}

	if rightLogical.Operator != "OR" {
		t.Errorf("right operator wrong. expected OR, got %q", rightLogical.Operator)
	}
}

// TestEmptyInput tests parsing empty input
func TestEmptyInput(t *testing.T) {
	input := ``

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if program == nil {
		t.Fatal("ParseProgram() returned nil")
	}

	if program.Expression != nil {
		t.Errorf("expected nil expression for empty input, got %T", program.Expression)
	}
}

// TestBareIdentifierInExpression tests identifiers without comparisons
func TestBareIdentifierInExpression(t *testing.T) {
	// This tests that we can have identifiers without comparison operators
	// Useful for truthy checks like "active AND premium"
	input := `active AND premium`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	logical, ok := program.Expression.(*ast.LogicalExpression)
	if !ok {
		t.Fatalf("expression not *ast.LogicalExpression. got=%T", program.Expression)
	}

	// Both sides should be Identifiers
	_, ok = logical.Left.(*ast.Identifier)
	if !ok {
		t.Errorf("left not Identifier. got=%T", logical.Left)
	}

	_, ok = logical.Right.(*ast.Identifier)
	if !ok {
		t.Errorf("right not Identifier. got=%T", logical.Right)
	}
}

// TestMultipleNOTs tests chained NOT operators
func TestMultipleNOTs(t *testing.T) {
	input := `NOT NOT active`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Outer NOT
	outer, ok := program.Expression.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("expression not *ast.UnaryExpression. got=%T", program.Expression)
	}

	// Inner NOT
	inner, ok := outer.Right.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("inner not *ast.UnaryExpression. got=%T", outer.Right)
	}

	// Innermost identifier
	_, ok = inner.Right.(*ast.Identifier)
	if !ok {
		t.Errorf("innermost not Identifier. got=%T", inner.Right)
	}
}

// TestErrorRecovery tests that parser collects errors instead of panicking
func TestErrorRecovery(t *testing.T) {
	tests := []struct {
		input        string
		expectedErrs int
	}{
		{"(age > 18", 1}, // Missing closing paren
		{"age > ", 1},    // Missing right operand (will try to parse EOF)
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		_ = p.ParseProgram()

		errors := p.Errors()
		if len(errors) < tt.expectedErrs {
			t.Errorf("expected at least %d errors for %q, got %d", tt.expectedErrs, tt.input, len(errors))
		}
	}
}
