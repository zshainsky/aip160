// Package ast defines the Abstract Syntax Tree nodes for AIP-160 filter expressions.
//
// The AST represents the parsed structure of a filter expression as a tree of nodes.
// Each node implements the Node interface and represents a specific element of the filter.
package ast

import (
	"fmt"
	"strings"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// Node is the base interface that all AST nodes must implement.
// It provides common functionality for all nodes in the tree.
type Node interface {
	// TokenLiteral returns the literal value of the token this node is associated with.
	// Useful for debugging and error messages.
	TokenLiteral() string

	// String returns a string representation of the node (for debugging/testing).
	String() string
}

// Expression represents any node that can be evaluated to produce a value.
// All filter components (comparisons, literals, identifiers, etc.) are expressions.
type Expression interface {
	Node
	expressionNode() // marker method to distinguish expressions from other node types
}

// Program is the root node of the AST.
// For AIP-160 filters, a program contains a single filter expression.
type Program struct {
	Expression Expression
}

func (p *Program) TokenLiteral() string {
	if p.Expression != nil {
		return p.Expression.TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	if p.Expression != nil {
		return p.Expression.String()
	}
	return ""
}

// Identifier represents a field name in a filter expression.
// Examples: name, age, user, email
type Identifier struct {
	Token lexer.Token // The IDENTIFIER token
	Value string      // The actual identifier name
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// StringLiteral represents a string value in quotes.
// Examples: "John", 'active'
type StringLiteral struct {
	Token lexer.Token // The STRING token
	Value string      // The string value (without quotes)
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return "\"" + s.Value + "\"" }

// NumberLiteral represents a numeric value (integer or float).
// Examples: 42, 3.14, 2.5e10
type NumberLiteral struct {
	Token lexer.Token // The NUMBER token
	Value float64     // The parsed numeric value
}

func (n *NumberLiteral) expressionNode()      {}
func (n *NumberLiteral) TokenLiteral() string { return n.Token.Literal }

// TODO: Implement String() method
// Hint: Use fmt.Sprintf() with %g format specifier
func (n *NumberLiteral) String() string {
	// TODO: Return the number as a formatted string
	return fmt.Sprintf("%g", n.Value)
}

// BooleanLiteral represents a boolean value.
// Examples: true, false
type BooleanLiteral struct {
	Token lexer.Token // The TRUE or FALSE token
	Value bool        // The boolean value
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }

// TODO: Implement String() method
// Hint: Return "true" or "false" based on the Value field
func (b *BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NullLiteral represents a null value.
type NullLiteral struct {
	Token lexer.Token // The NULL token
}

func (n *NullLiteral) expressionNode()      {}
func (n *NullLiteral) TokenLiteral() string { return n.Token.Literal }

// TODO: Implement String() method
// Hint: Just return "null"
func (n *NullLiteral) String() string {
	return "null"
}

// ComparisonExpression represents a comparison between two expressions.
// Examples: age > 18, name = "John", status != "inactive"
type ComparisonExpression struct {
	Token    lexer.Token // The operator token (=, !=, <, >, <=, >=)
	Left     Expression  // Left side of the comparison
	Operator string      // The comparison operator as a string
	Right    Expression  // Right side of the comparison
}

func (c *ComparisonExpression) expressionNode()      {}
func (c *ComparisonExpression) TokenLiteral() string { return c.Token.Literal }

// TODO: Implement String() method
// Hint: Format should be "(<left> <operator> <right>)"
// Remember to call .String() on Left and Right!
func (c *ComparisonExpression) String() string {
	// TODO: Return formatted string with parentheses
	return fmt.Sprintf("(%s %s %s)", c.Left.String(), c.Operator, c.Right.String())
}

// LogicalExpression represents a logical operation between two expressions.
// Examples: age > 18 AND status = "active", role = "admin" OR role = "owner"
type LogicalExpression struct {
	Token    lexer.Token // The operator token (AND, OR)
	Left     Expression  // Left side of the logical operation
	Operator string      // The logical operator ("AND" or "OR")
	Right    Expression  // Right side of the logical operation
}

func (l *LogicalExpression) expressionNode()      {}
func (l *LogicalExpression) TokenLiteral() string { return l.Token.Literal }

// TODO: Implement String() method
// Hint: Same format as ComparisonExpression: "(<left> <operator> <right>)"
func (l *LogicalExpression) String() string {
	// TODO: Return formatted string with parentheses
	return fmt.Sprintf("(%s %s %s)", l.Left.String(), l.Operator, l.Right.String())
}

// UnaryExpression represents a unary operation (currently just NOT or negation).
// Examples: NOT status = "inactive", -age
type UnaryExpression struct {
	Token    lexer.Token // The operator token (NOT, MINUS)
	Operator string      // The unary operator
	Right    Expression  // The expression being operated on
}

func (u *UnaryExpression) expressionNode()      {}
func (u *UnaryExpression) TokenLiteral() string { return u.Token.Literal }

// TODO: Implement String() method
// Hint: Format should be "(<operator><right>)" - no space between operator and right
func (u *UnaryExpression) String() string {
	// TODO: Return formatted string with parentheses
	if u.Operator == "NOT" {
		return fmt.Sprintf("(%s %s)", u.Operator, u.Right.String())
	}
	return fmt.Sprintf("(%s%s)", u.Operator, u.Right.String())
}

// TraversalExpression represents field traversal with the dot operator.
// Examples: user.name, address.city, author.email
type TraversalExpression struct {
	Token lexer.Token // The DOT token
	Left  Expression  // Left side (an identifier or another traversal)
	Right Expression  // Right side (an identifier)
}

func (t *TraversalExpression) expressionNode()      {}
func (t *TraversalExpression) TokenLiteral() string { return t.Token.Literal }

// TODO: Implement String() method
// Hint: Format should be "<left>.<right>" - NO parentheses for this one!
func (t *TraversalExpression) String() string {
	// TODO: Return formatted string with dot separator
	return fmt.Sprintf("%s.%s", t.Left.String(), t.Right.String())
}

// HasExpression represents the has operator for checking collection membership.
// Examples: tags:urgent, labels:"bug", roles:admin
type HasExpression struct {
	Token      lexer.Token // The HAS (:) token
	Collection Expression  // The collection field
	Member     Expression  // The member to check for
}

func (h *HasExpression) expressionNode()      {}
func (h *HasExpression) TokenLiteral() string { return h.Token.Literal }

// TODO: Implement String() method
// Hint: Format should be "<collection>:<member>" - NO parentheses
func (h *HasExpression) String() string {
	// TODO: Return formatted string with colon separator
	return fmt.Sprintf("%s:%s", h.Collection.String(), h.Member.String())
}

// FunctionCall represents a function call expression.
// Examples: timestamp(created_at), duration(start, end)
type FunctionCall struct {
	Token     lexer.Token  // The IDENTIFIER token (function name)
	Function  string       // Function name
	Arguments []Expression // List of argument expressions
}

func (f *FunctionCall) expressionNode()      {}
func (f *FunctionCall) TokenLiteral() string { return f.Token.Literal }

// TODO: Implement String() method
// Hint: Format should be "<function>(<arg1>, <arg2>, ...)"
// You'll need to loop through Arguments and call String() on each
// Then join them with ", " using strings.Join()
func (f *FunctionCall) String() string {
	// TODO: Build argument list and return formatted string
	var args []string
	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Function, strings.Join(args, ", "))
}
