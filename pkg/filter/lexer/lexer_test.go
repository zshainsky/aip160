package lexer

import (
	"testing"
)

// TestNextToken_BasicOperators tests single-character operators and delimiters
func TestNextToken_BasicOperators(t *testing.T) {
	input := `= != < > <= >= : . ( ) * -`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{EQUALS, "="},
		{NOT_EQUALS, "!="},
		{LESS_THAN, "<"},
		{GREATER_THAN, ">"},
		{LESS_EQUAL, "<="},
		{GREATER_EQUAL, ">="},
		{HAS, ":"},
		{DOT, "."},
		{LPAREN, "("},
		{RPAREN, ")"},
		{ASTERISK, "*"},
		{MINUS, "-"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Identifiers tests identifier recognition
func TestNextToken_Identifiers(t *testing.T) {
	input := `name age user_id userName`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENTIFIER, "name"},
		{IDENTIFIER, "age"},
		{IDENTIFIER, "user_id"},
		{IDENTIFIER, "userName"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Keywords tests keyword recognition (case-insensitive)
func TestNextToken_Keywords(t *testing.T) {
	input := `AND OR NOT and or not True FALSE null`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{AND, "AND"},
		{OR, "OR"},
		{NOT, "NOT"},
		{AND, "and"},
		{OR, "or"},
		{NOT, "not"},
		{TRUE, "True"},
		{FALSE, "FALSE"},
		{NULL, "null"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Numbers tests number tokenization (integers and floats)
func TestNextToken_Numbers(t *testing.T) {
	input := `42 3.14 2.997e9 1e-10 0`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{NUMBER, "42"},
		{NUMBER, "3.14"},
		{NUMBER, "2.997e9"},
		{NUMBER, "1e-10"},
		{NUMBER, "0"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Strings tests string literal tokenization
func TestNextToken_Strings(t *testing.T) {
	input := `"hello" 'world' "with spaces" 'single quote'`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{STRING, "hello"},
		{STRING, "world"},
		{STRING, "with spaces"},
		{STRING, "single quote"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_SimpleFilter tests a complete simple filter
func TestNextToken_SimpleFilter(t *testing.T) {
	input := `name = "John"`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENTIFIER, "name"},
		{EQUALS, "="},
		{STRING, "John"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_ComplexFilter tests a complex filter with logical operators
func TestNextToken_ComplexFilter(t *testing.T) {
	input := `age > 18 AND status = "active" OR premium = true`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENTIFIER, "age"},
		{GREATER_THAN, ">"},
		{NUMBER, "18"},
		{AND, "AND"},
		{IDENTIFIER, "status"},
		{EQUALS, "="},
		{STRING, "active"},
		{OR, "OR"},
		{IDENTIFIER, "premium"},
		{EQUALS, "="},
		{TRUE, "true"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_TraversalAndHas tests the traversal and has operators
func TestNextToken_TraversalAndHas(t *testing.T) {
	input := `user.name = "John" AND tags:42`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENTIFIER, "user"},
		{DOT, "."},
		{IDENTIFIER, "name"},
		{EQUALS, "="},
		{STRING, "John"},
		{AND, "AND"},
		{IDENTIFIER, "tags"},
		{HAS, ":"},
		{NUMBER, "42"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Negation tests negation operators
func TestNextToken_Negation(t *testing.T) {
	input := `NOT active AND -deleted`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{NOT, "NOT"},
		{IDENTIFIER, "active"},
		{AND, "AND"},
		{MINUS, "-"},
		{IDENTIFIER, "deleted"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_Wildcards tests wildcard patterns
func TestNextToken_Wildcards(t *testing.T) {
	input := `name = "*.google.com" AND tags:*`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENTIFIER, "name"},
		{EQUALS, "="},
		{STRING, "*.google.com"},
		{AND, "AND"},
		{IDENTIFIER, "tags"},
		{HAS, ":"},
		{ASTERISK, "*"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_WithParentheses tests grouping with parentheses
func TestNextToken_WithParentheses(t *testing.T) {
	input := `(age > 18 AND age < 65) OR retired = true`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LPAREN, "("},
		{IDENTIFIER, "age"},
		{GREATER_THAN, ">"},
		{NUMBER, "18"},
		{AND, "AND"},
		{IDENTIFIER, "age"},
		{LESS_THAN, "<"},
		{NUMBER, "65"},
		{RPAREN, ")"},
		{OR, "OR"},
		{IDENTIFIER, "retired"},
		{EQUALS, "="},
		{TRUE, "true"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

// TestNextToken_EmptyString tests handling of empty input
func TestNextToken_EmptyString(t *testing.T) {
	input := ``

	l := New(input)
	tok := l.NextToken()

	if tok.Type != EOF {
		t.Fatalf("expected EOF token for empty input, got=%q", tok.Type)
	}
}

// TestNextToken_OnlyWhitespace tests handling of whitespace-only input
func TestNextToken_OnlyWhitespace(t *testing.T) {
	input := `   	
	`

	l := New(input)
	tok := l.NextToken()

	if tok.Type != EOF {
		t.Fatalf("expected EOF token for whitespace-only input, got=%q", tok.Type)
	}
}

// TestLookupIdentifier tests the keyword lookup function
func TestLookupIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected TokenType
	}{
		{"AND", AND},
		{"and", AND},
		{"Or", OR},
		{"NOT", NOT},
		{"true", TRUE},
		{"False", FALSE},
		{"NULL", NULL},
		{"name", IDENTIFIER},
		{"userName", IDENTIFIER},
		{"user_id", IDENTIFIER},
	}

	for _, tt := range tests {
		result := LookupIdentifier(tt.input)
		if result != tt.expected {
			t.Errorf("LookupIdentifier(%q) = %q, want %q",
				tt.input, result, tt.expected)
		}
	}
}
