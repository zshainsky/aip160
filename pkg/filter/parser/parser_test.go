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

// ============================================================================
// Module 5 Tests: Advanced Features
// These tests cover field traversal, has operator, and function calls
// ============================================================================

// TestFieldTraversal tests parsing of dot notation for field access
func TestFieldTraversal(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user.name", "user.name"},
		{"address.city", "address.city"},
		{"a.b.c", "((a.b).c)"},
		{"user.profile.email", "((user.profile).email)"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program.Expression == nil {
			t.Fatalf("program.Expression is nil for input %q", tt.input)
		}

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("traversal string wrong for %q.\nexpected: %q\ngot:      %q",
				tt.input, tt.expected, actual)
		}
	}
}

// TestFieldTraversalInComparison tests traversal used in comparisons
func TestFieldTraversalInComparison(t *testing.T) {
	input := `user.name = "John"`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	comp, ok := program.Expression.(*ast.ComparisonExpression)
	if !ok {
		t.Fatalf("expression not *ast.ComparisonExpression. got=%T", program.Expression)
	}

	// Left side should be a traversal
	traversal, ok := comp.Left.(*ast.TraversalExpression)
	if !ok {
		t.Fatalf("left not *ast.TraversalExpression. got=%T", comp.Left)
	}

	// Check traversal structure
	leftIdent, ok := traversal.Left.(*ast.Identifier)
	if !ok || leftIdent.Value != "user" {
		t.Errorf("traversal left wrong. expected Identifier(user), got=%T", traversal.Left)
	}

	rightIdent, ok := traversal.Right.(*ast.Identifier)
	if !ok || rightIdent.Value != "name" {
		t.Errorf("traversal right wrong. expected Identifier(name), got=%T", traversal.Right)
	}

	// Check operator and right side
	if comp.Operator != "=" {
		t.Errorf("operator wrong. expected=, got=%q", comp.Operator)
	}

	str, ok := comp.Right.(*ast.StringLiteral)
	if !ok || str.Value != "John" {
		t.Errorf("right side wrong. expected StringLiteral(John), got=%T", comp.Right)
	}
}

// TestHasOperator tests parsing of the has operator
func TestHasOperator(t *testing.T) {
	tests := []struct {
		input      string
		collection string
		member     string
	}{
		{"tags:urgent", "tags", "urgent"},
		{"roles:admin", "roles", "admin"},
		{"labels:\"bug\"", "labels", "\"bug\""},
		{"permissions:write", "permissions", "write"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		hasExpr, ok := program.Expression.(*ast.HasExpression)
		if !ok {
			t.Fatalf("expression not *ast.HasExpression for %q. got=%T",
				tt.input, program.Expression)
		}

		if hasExpr.Collection.String() != tt.collection {
			t.Errorf("collection wrong for %q. expected=%q, got=%q",
				tt.input, tt.collection, hasExpr.Collection.String())
		}

		if hasExpr.Member.String() != tt.member {
			t.Errorf("member wrong for %q. expected=%q, got=%q",
				tt.input, tt.member, hasExpr.Member.String())
		}
	}
}

// TestHasOperatorWithTraversal tests combining has and traversal
func TestHasOperatorWithTraversal(t *testing.T) {
	input := `user.tags:urgent`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	hasExpr, ok := program.Expression.(*ast.HasExpression)
	if !ok {
		t.Fatalf("expression not *ast.HasExpression. got=%T", program.Expression)
	}

	// Collection should be a traversal
	traversal, ok := hasExpr.Collection.(*ast.TraversalExpression)
	if !ok {
		t.Fatalf("collection not *ast.TraversalExpression. got=%T", hasExpr.Collection)
	}

	if traversal.String() != "user.tags" {
		t.Errorf("traversal wrong. expected user.tags, got=%s", traversal.String())
	}

	// Member should be identifier
	member, ok := hasExpr.Member.(*ast.Identifier)
	if !ok || member.Value != "urgent" {
		t.Errorf("member wrong. expected Identifier(urgent), got=%T", hasExpr.Member)
	}
}

// TestHasOperatorPrecedence tests operator precedence with has
func TestHasOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Has binds tighter than comparison
		{"tags:urgent = true", "((tags:urgent) = true)"},
		// Has binds looser than traversal
		{"user.tags:admin", "(user.tags):admin"},
		// Multiple has operators
		{"tags:a OR roles:b", "((tags:a) OR (roles:b))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("precedence wrong for %q.\nexpected: %q\ngot:      %q",
				tt.input, tt.expected, actual)
		}
	}
}

// TestFunctionCall tests parsing of function calls
func TestFunctionCall(t *testing.T) {
	tests := []struct {
		input    string
		function string
		numArgs  int
	}{
		{"now()", "now", 0},
		{"timestamp(created_at)", "timestamp", 1},
		{"has(tags, \"urgent\")", "has", 2},
		{"duration(start, end, unit)", "duration", 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		funcCall, ok := program.Expression.(*ast.FunctionCall)
		if !ok {
			t.Fatalf("expression not *ast.FunctionCall for %q. got=%T",
				tt.input, program.Expression)
		}

		if funcCall.Function != tt.function {
			t.Errorf("function name wrong for %q. expected=%q, got=%q",
				tt.input, tt.function, funcCall.Function)
		}

		if len(funcCall.Arguments) != tt.numArgs {
			t.Errorf("argument count wrong for %q. expected=%d, got=%d",
				tt.input, tt.numArgs, len(funcCall.Arguments))
		}
	}
}

// TestFunctionCallWithComplexArguments tests functions with expression arguments
func TestFunctionCallWithComplexArguments(t *testing.T) {
	input := `validate(age > 18, status = "active")`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	funcCall, ok := program.Expression.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expression not *ast.FunctionCall. got=%T", program.Expression)
	}

	if funcCall.Function != "validate" {
		t.Errorf("function name wrong. expected=validate, got=%q", funcCall.Function)
	}

	if len(funcCall.Arguments) != 2 {
		t.Fatalf("expected 2 arguments, got=%d", len(funcCall.Arguments))
	}

	// First argument should be a comparison
	_, ok = funcCall.Arguments[0].(*ast.ComparisonExpression)
	if !ok {
		t.Errorf("first argument not comparison. got=%T", funcCall.Arguments[0])
	}

	// Second argument should be a comparison
	_, ok = funcCall.Arguments[1].(*ast.ComparisonExpression)
	if !ok {
		t.Errorf("second argument not comparison. got=%T", funcCall.Arguments[1])
	}
}

// TestFunctionCallInComparison tests functions used in larger expressions
func TestFunctionCallInComparison(t *testing.T) {
	input := `timestamp(created_at) > 1000`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	comp, ok := program.Expression.(*ast.ComparisonExpression)
	if !ok {
		t.Fatalf("expression not *ast.ComparisonExpression. got=%T", program.Expression)
	}

	// Left side should be function call
	funcCall, ok := comp.Left.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("left not *ast.FunctionCall. got=%T", comp.Left)
	}

	if funcCall.Function != "timestamp" {
		t.Errorf("function wrong. expected=timestamp, got=%q", funcCall.Function)
	}

	// Right side should be number
	num, ok := comp.Right.(*ast.NumberLiteral)
	if !ok || num.Value != 1000 {
		t.Errorf("right not NumberLiteral(1000). got=%T", comp.Right)
	}
}

// TestComplexNestedExpressions tests combinations of all Module 5 features
func TestComplexNestedExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "traversal with has in logical expression",
			input: `user.tags:urgent AND status = "active"`,
		},
		{
			name:  "function with traversal argument",
			input: `timestamp(user.created_at) > 1000`,
		},
		{
			name:  "has with function member",
			input: `tags:get_priority()`,
		},
		{
			name:  "deeply nested traversal",
			input: `company.department.manager.email = "admin@example.com"`,
		},
		{
			name:  "function in complex expression",
			input: `validate(user.age) AND user.status = "verified"`,
		},
		{
			name:  "multiple has operators",
			input: `roles:admin OR permissions:write`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if program.Expression == nil {
				t.Fatal("program.Expression is nil")
			}

			// Just verify it parses without error
			// Specific structure tests are in individual test functions
			_ = program.String()
		})
	}
}

// TestModule4StillWorks ensures that adding Module 5 features doesn't break Module 4
func TestModule4StillWorks(t *testing.T) {
	tests := []string{
		`age > 18`,
		`name = "John"`,
		`active AND premium`,
		`NOT disabled`,
		`age > 18 OR status = "vip"`,
		`(a OR b) AND c`,
	}

	for _, input := range tests {
		l := lexer.New(input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if program.Expression == nil {
			t.Fatalf("Module 4 test failed for: %q - expression is nil", input)
		}
	}
}

// TestTraversalDoesNotBreakSimpleIdentifiers ensures identifiers still work
func TestTraversalDoesNotBreakSimpleIdentifiers(t *testing.T) {
	input := `status`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	ident, ok := program.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression should still be Identifier. got=%T", program.Expression)
	}

	if ident.Value != "status" {
		t.Errorf("identifier value wrong. expected=status, got=%q", ident.Value)
	}
}

// TestFunctionDoesNotBreakSimpleIdentifiers ensures parsing functions doesn't break identifiers
func TestFunctionDoesNotBreakSimpleIdentifiers(t *testing.T) {
	// Identifier without parentheses should still be identifier
	input := `timestamp`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	ident, ok := program.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression should still be Identifier. got=%T", program.Expression)
	}

	if ident.Value != "timestamp" {
		t.Errorf("identifier value wrong. expected=timestamp, got=%q", ident.Value)
	}
}

// TestErrorHandlingModule5 tests that invalid syntax produces appropriate errors
func TestErrorHandlingModule5(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed function call",
			input: `func(arg`,
		},
		{
			name:  "missing argument after comma",
			input: `func(a,`,
		},
		{
			name:  "dot without following identifier",
			input: `user.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_ = p.ParseProgram()

			errors := p.Errors()
			if len(errors) == 0 {
				t.Errorf("expected errors for %q, got none", tt.input)
			}
		})
	}
}
