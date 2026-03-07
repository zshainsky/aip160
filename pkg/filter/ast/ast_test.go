package ast

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func TestIdentifier(t *testing.T) {
	ident := &Identifier{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "name"},
		Value: "name",
	}

	if ident.TokenLiteral() != "name" {
		t.Errorf("TokenLiteral() = %q, want %q", ident.TokenLiteral(), "name")
	}

	if ident.String() != "name" {
		t.Errorf("String() = %q, want %q", ident.String(), "name")
	}
}

func TestStringLiteral(t *testing.T) {
	str := &StringLiteral{
		Token: lexer.Token{Type: lexer.STRING, Literal: "John"},
		Value: "John",
	}

	if str.TokenLiteral() != "John" {
		t.Errorf("TokenLiteral() = %q, want %q", str.TokenLiteral(), "John")
	}

	// String representation should include quotes
	if str.String() != `"John"` {
		t.Errorf("String() = %q, want %q", str.String(), `"John"`)
	}
}

func TestNumberLiteral(t *testing.T) {
	tests := []struct {
		name    string
		token   lexer.Token
		value   float64
		wantStr string
	}{
		{
			name:    "integer",
			token:   lexer.Token{Type: lexer.NUMBER, Literal: "42"},
			value:   42,
			wantStr: "42",
		},
		{
			name:    "float",
			token:   lexer.Token{Type: lexer.NUMBER, Literal: "3.14"},
			value:   3.14,
			wantStr: "3.14",
		},
		{
			name:    "scientific notation",
			token:   lexer.Token{Type: lexer.NUMBER, Literal: "2.5e10"},
			value:   2.5e10,
			wantStr: "2.5e+10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num := &NumberLiteral{
				Token: tt.token,
				Value: tt.value,
			}

			if num.TokenLiteral() != tt.token.Literal {
				t.Errorf("TokenLiteral() = %q, want %q", num.TokenLiteral(), tt.token.Literal)
			}

			if num.String() != tt.wantStr {
				t.Errorf("String() = %q, want %q", num.String(), tt.wantStr)
			}
		})
	}
}

func TestBooleanLiteral(t *testing.T) {
	tests := []struct {
		name  string
		token lexer.Token
		value bool
	}{
		{
			name:  "true",
			token: lexer.Token{Type: lexer.TRUE, Literal: "true"},
			value: true,
		},
		{
			name:  "false",
			token: lexer.Token{Type: lexer.FALSE, Literal: "false"},
			value: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BooleanLiteral{
				Token: tt.token,
				Value: tt.value,
			}

			if b.TokenLiteral() != tt.token.Literal {
				t.Errorf("TokenLiteral() = %q, want %q", b.TokenLiteral(), tt.token.Literal)
			}

			if tt.value {
				if b.String() != "true" {
					t.Errorf("String() = %q, want %q", b.String(), "true")
				}
			} else {
				if b.String() != "false" {
					t.Errorf("String() = %q, want %q", b.String(), "false")
				}
			}
		})
	}
}

func TestNullLiteral(t *testing.T) {
	null := &NullLiteral{
		Token: lexer.Token{Type: lexer.NULL, Literal: "null"},
	}

	if null.TokenLiteral() != "null" {
		t.Errorf("TokenLiteral() = %q, want %q", null.TokenLiteral(), "null")
	}

	if null.String() != "null" {
		t.Errorf("String() = %q, want %q", null.String(), "null")
	}
}

func TestComparisonExpression(t *testing.T) {
	// Test: age > 18
	expr := &ComparisonExpression{
		Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
		Left: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
			Value: "age",
		},
		Operator: ">",
		Right: &NumberLiteral{
			Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
			Value: 18,
		},
	}

	if expr.TokenLiteral() != ">" {
		t.Errorf("TokenLiteral() = %q, want %q", expr.TokenLiteral(), ">")
	}

	expected := "(age > 18)"
	if expr.String() != expected {
		t.Errorf("String() = %q, want %q", expr.String(), expected)
	}
}

func TestLogicalExpression(t *testing.T) {
	// Test: age > 18 AND status = "active"
	expr := &LogicalExpression{
		Token: lexer.Token{Type: lexer.AND, Literal: "AND"},
		Left: &ComparisonExpression{
			Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
			Left: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
				Value: "age",
			},
			Operator: ">",
			Right: &NumberLiteral{
				Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
				Value: 18,
			},
		},
		Operator: "AND",
		Right: &ComparisonExpression{
			Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
			Left: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "status"},
				Value: "status",
			},
			Operator: "=",
			Right: &StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "active"},
				Value: "active",
			},
		},
	}

	if expr.TokenLiteral() != "AND" {
		t.Errorf("TokenLiteral() = %q, want %q", expr.TokenLiteral(), "AND")
	}

	expected := `((age > 18) AND (status = "active"))`
	if expr.String() != expected {
		t.Errorf("String() = %q, want %q", expr.String(), expected)
	}
}

func TestUnaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		token    lexer.Token
		right    Expression
		expected string
	}{
		{
			name:     "NOT operator",
			operator: "NOT",
			token:    lexer.Token{Type: lexer.NOT, Literal: "NOT"},
			right: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "active"},
				Value: "active",
			},
			expected: "(NOT active)",
		},
		{
			name:     "negation",
			operator: "-",
			token:    lexer.Token{Type: lexer.MINUS, Literal: "-"},
			right: &NumberLiteral{
				Token: lexer.Token{Type: lexer.NUMBER, Literal: "5"},
				Value: 5,
			},
			expected: "(-5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &UnaryExpression{
				Token:    tt.token,
				Operator: tt.operator,
				Right:    tt.right,
			}

			if expr.TokenLiteral() != tt.token.Literal {
				t.Errorf("TokenLiteral() = %q, want %q", expr.TokenLiteral(), tt.token.Literal)
			}

			if expr.String() != tt.expected {
				t.Errorf("String() = %q, want %q", expr.String(), tt.expected)
			}
		})
	}
}

func TestTraversalExpression(t *testing.T) {
	// Test: user.email
	expr := &TraversalExpression{
		Token: lexer.Token{Type: lexer.DOT, Literal: "."},
		Left: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "user"},
			Value: "user",
		},
		Right: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "email"},
			Value: "email",
		},
	}

	if expr.TokenLiteral() != "." {
		t.Errorf("TokenLiteral() = %q, want %q", expr.TokenLiteral(), ".")
	}

	expected := "user.email"
	if expr.String() != expected {
		t.Errorf("String() = %q, want %q", expr.String(), expected)
	}
}

func TestNestedTraversalExpression(t *testing.T) {
	// Test: user.address.city (nested traversal)
	expr := &TraversalExpression{
		Token: lexer.Token{Type: lexer.DOT, Literal: "."},
		Left: &TraversalExpression{
			Token: lexer.Token{Type: lexer.DOT, Literal: "."},
			Left: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "user"},
				Value: "user",
			},
			Right: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "address"},
				Value: "address",
			},
		},
		Right: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "city"},
			Value: "city",
		},
	}

	expected := "user.address.city"
	if expr.String() != expected {
		t.Errorf("String() = %q, want %q", expr.String(), expected)
	}
}

func TestHasExpression(t *testing.T) {
	// Test: tags:urgent
	expr := &HasExpression{
		Token: lexer.Token{Type: lexer.HAS, Literal: ":"},
		Collection: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "tags"},
			Value: "tags",
		},
		Member: &Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "urgent"},
			Value: "urgent",
		},
	}

	if expr.TokenLiteral() != ":" {
		t.Errorf("TokenLiteral() = %q, want %q", expr.TokenLiteral(), ":")
	}

	expected := "tags:urgent"
	if expr.String() != expected {
		t.Errorf("String() = %q, want %q", expr.String(), expected)
	}
}

func TestFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		function string
		args     []Expression
		expected string
	}{
		{
			name:     "no arguments",
			function: "now",
			args:     []Expression{},
			expected: "now()",
		},
		{
			name:     "single argument",
			function: "timestamp",
			args: []Expression{
				&Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "created_at"},
					Value: "created_at",
				},
			},
			expected: "timestamp(created_at)",
		},
		{
			name:     "multiple arguments",
			function: "duration",
			args: []Expression{
				&Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "start"},
					Value: "start",
				},
				&Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "end"},
					Value: "end",
				},
			},
			expected: "duration(start, end)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := &FunctionCall{
				Token:     lexer.Token{Type: lexer.IDENTIFIER, Literal: tt.function},
				Function:  tt.function,
				Arguments: tt.args,
			}

			if fn.TokenLiteral() != tt.function {
				t.Errorf("TokenLiteral() = %q, want %q", fn.TokenLiteral(), tt.function)
			}

			if fn.String() != tt.expected {
				t.Errorf("String() = %q, want %q", fn.String(), tt.expected)
			}
		})
	}
}

func TestProgram(t *testing.T) {
	// Test: age > 18
	program := &Program{
		Expression: &ComparisonExpression{
			Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
			Left: &Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
				Value: "age",
			},
			Operator: ">",
			Right: &NumberLiteral{
				Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
				Value: 18,
			},
		},
	}

	expected := "(age > 18)"
	if program.String() != expected {
		t.Errorf("String() = %q, want %q", program.String(), expected)
	}

	if program.TokenLiteral() != ">" {
		t.Errorf("TokenLiteral() = %q, want %q", program.TokenLiteral(), ">")
	}
}

func TestComplexExpression(t *testing.T) {
	// Test: (age > 18 AND status = "active") OR role = "admin"
	program := &Program{
		Expression: &LogicalExpression{
			Token: lexer.Token{Type: lexer.OR, Literal: "OR"},
			Left: &LogicalExpression{
				Token: lexer.Token{Type: lexer.AND, Literal: "AND"},
				Left: &ComparisonExpression{
					Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
					Left: &Identifier{
						Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
						Value: "age",
					},
					Operator: ">",
					Right: &NumberLiteral{
						Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
						Value: 18,
					},
				},
				Operator: "AND",
				Right: &ComparisonExpression{
					Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
					Left: &Identifier{
						Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "status"},
						Value: "status",
					},
					Operator: "=",
					Right: &StringLiteral{
						Token: lexer.Token{Type: lexer.STRING, Literal: "active"},
						Value: "active",
					},
				},
			},
			Operator: "OR",
			Right: &ComparisonExpression{
				Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
				Left: &Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "role"},
					Value: "role",
				},
				Operator: "=",
				Right: &StringLiteral{
					Token: lexer.Token{Type: lexer.STRING, Literal: "admin"},
					Value: "admin",
				},
			},
		},
	}

	expected := `(((age > 18) AND (status = "active")) OR (role = "admin"))`
	if program.String() != expected {
		t.Errorf("String() = %q, want %q", program.String(), expected)
	}
}
