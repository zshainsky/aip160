package parser

import (
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// Parser performs parsing of filter expressions
type Parser struct {
	lexer        *lexer.Lexer
	errors       []string
	currentToken lexer.Token
	peekToken    lexer.Token
}

// New creates a new Parser instance
func New(l *lexer.Lexer) *Parser {
	// TODO: Create Parser instance
	// TODO: Initialize errors slice
	// TODO: Call nextToken() twice to prime the pump
	// TODO: Return the parser
	return nil
}

// nextToken advances both the current and peek tokens
func (p *Parser) nextToken() {
	// TODO: Move peekToken to currentToken
	// TODO: Get next token from lexer and set as peekToken
}

// currentTokenIs checks if the current token matches the given type
func (p *Parser) currentTokenIs(t lexer.TokenType) bool {
	// TODO: Compare p.currentToken.Type with t
	return false
}

// peekTokenIs checks if the peek token matches the given type
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	// TODO: Compare p.peekToken.Type with t
	return false
}

// expectPeek checks if the next token is of expected type and advances if so
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	// TODO: Check if peekToken matches type
	// TODO: If yes: call nextToken() and return true
	// TODO: If no: call peekError(t) and return false
	return false
}

// Errors returns the list of parsing errors
func (p *Parser) Errors() []string {
	// TODO: Return the errors slice
	return nil
}

// peekError adds an error when the peek token doesn't match expectations
func (p *Parser) peekError(t lexer.TokenType) {
	// TODO: Create error message: "expected next token to be X, got Y instead"
	// TODO: Append to p.errors
}

// ParseProgram is the entry point for parsing a filter expression
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}

	// Handle empty input
	if p.currentToken.Type == lexer.EOF {
		return program
	}

	program.Expression = p.parseExpression()
	return program
}

// parseExpression is the top-level expression parser
// Grammar: expression = or_expression ;
func (p *Parser) parseExpression() ast.Expression {
	return p.parseOrExpression()
}

// parseOrExpression handles OR operators (lowest precedence)
// Grammar: or_expression = and_expression, {"OR", and_expression} ;
func (p *Parser) parseOrExpression() ast.Expression {
	// INITIAL IMPLEMENTATION: Simple pass-through for testing
	// This allows basic parsing to work while you implement other features
	// Task 7 will replace this with full OR operator handling
	return p.parseAndExpression()

	// TODO (Task 7): Enhance with OR operator loop
	// Get left operand from parseAndExpression()
	// While current token is OR:
	//   - Save the operator token
	//   - Advance to next token
	//   - Get right operand from parseAndExpression()
	//   - Create LogicalExpression node
	//   - Update left to be the new expression (for left-associativity)
	// Return left
}

// parseAndExpression handles AND operators (higher precedence than OR)
// Grammar: and_expression = not_expression, {"AND", not_expression} ;
func (p *Parser) parseAndExpression() ast.Expression {
	// INITIAL IMPLEMENTATION: Simple pass-through for testing
	// This allows basic parsing to work while you implement other features
	// Task 6 will replace this with full AND operator handling
	return p.parseNotExpression()

	// TODO (Task 6): Enhance with AND operator loop
	// Similar pattern to OR expression but:
	//   - Get left operand from parseNotExpression()
	//   - Loop while current token is AND
	//   - Create LogicalExpression nodes
	//   - Build left-associative chain
}

// parseNotExpression handles NOT operator (prefix, highest precedence)
// Grammar: not_expression = ["NOT"], comparison ;
func (p *Parser) parseNotExpression() ast.Expression {
	// INITIAL IMPLEMENTATION: Simple pass-through for testing
	// This allows basic parsing to work while you implement other features
	// Task 5 will replace this with full NOT operator handling
	return p.parseComparison()

	// TODO (Task 5): Enhance with NOT operator handling
	// Check if current token is NOT
	//   - If yes: save token, advance, recursively call parseNotExpression(),
	//     create UnaryExpression node, return it
	//   - If no: return parseComparison()
	// Hint: Recursive call allows "NOT NOT active"
}

// parseComparison handles comparison operators
// Grammar: comparison = value, [comparator, value] ;
func (p *Parser) parseComparison() ast.Expression {
	// INITIAL IMPLEMENTATION: Simple pass-through for testing
	// This allows basic parsing to work while you implement other features
	// Task 4 will replace this with full comparison operator handling
	return p.parseValue()

	// TODO (Task 4): Enhance with comparison operator handling
	// Get left value from parseValue()
	// Check if current token is a comparison operator (use isComparisonOperator)
	//   - If yes: save operator, advance, get right value,
	//     create ComparisonExpression node, return it
	//   - If no: return just the left value (allows bare identifiers)
}

// isComparisonOperator checks if a token type is a comparison operator
func (p *Parser) isComparisonOperator(t lexer.TokenType) bool {
	// TODO (Task 4): Implement comparison operator check
	// Use switch statement to check if t is one of:
	//   EQUALS, NOT_EQUALS, LESS_THAN, LESS_EQUAL, GREATER_THAN, GREATER_EQUAL
	// Return true if it matches, false otherwise
	return false
}

// parseValue dispatches to the appropriate parser based on token type
// Grammar: value = function_call | field | string | number | boolean | null | "(", expression, ")" ;
func (p *Parser) parseValue() ast.Expression {
	// TODO (Task 2): Implement minimal dispatcher
	// Start with just STRING case to get first test passing:
	// switch p.currentToken.Type {
	// case lexer.STRING:
	//     return p.parseString()
	// default:
	//     msg := fmt.Sprintf("unexpected token: %s", p.currentToken.Type)
	//     p.errors = append(p.errors, msg)
	//     return nil
	// }

	// TODO (Task 3): Add remaining literal cases
	//   case NUMBER: return p.parseNumber()
	//   case TRUE, FALSE: return p.parseBoolean()
	//   case NULL: return p.parseNull()
	//   case IDENTIFIER: return p.parseIdentifier()

	// TODO (Task 8): Add parentheses case
	//   case LPAREN: return p.parseGroupedExpression()
	return nil
}

// parseString parses a string literal
func (p *Parser) parseString() ast.Expression {
	// TODO (Task 2): Implement string literal parsing
	// Pattern: Create AST node → Set fields → Advance token → Return
	// lit := &ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}
	// p.nextToken()
	// return lit
	return nil
}

// parseNumber parses a numeric literal
func (p *Parser) parseNumber() ast.Expression {
	// TODO (Task 3): Implement number literal parsing
	// Create NumberLiteral with currentToken
	// Use strconv.ParseFloat(p.currentToken.Literal, 64) to convert
	// If error, append to p.errors and return nil
	// Set num.Value to the parsed value
	// Call nextToken() to advance
	// Return the NumberLiteral
	return nil
}

// parseBoolean parses a boolean literal (true or false)
func (p *Parser) parseBoolean() ast.Expression {
	// TODO (Task 3): Implement boolean literal parsing
	// Create BooleanLiteral with currentToken
	// Set Value to true if currentToken is TRUE, false otherwise
	// Call nextToken() to advance
	// Return the BooleanLiteral
	return nil
}

// parseNull parses a null literal
func (p *Parser) parseNull() ast.Expression {
	// TODO (Task 3): Implement null literal parsing
	// Create NullLiteral with currentToken
	// Call nextToken() to advance
	// Return the NullLiteral
	return nil
}

// parseIdentifier parses a simple identifier
// Module 5 will extend this to handle field traversal and function calls
func (p *Parser) parseIdentifier() ast.Expression {
	// TODO (Task 3): Implement identifier parsing
	// Create Identifier with currentToken and its Literal value
	// Call nextToken() to advance
	// Return the Identifier
	return nil
}

// parseGroupedExpression parses an expression wrapped in parentheses
// Grammar: "(", expression, ")"
func (p *Parser) parseGroupedExpression() ast.Expression {
	// TODO (Task 8): Implement grouped expression parsing
	// Call nextToken() to consume '('
	// Parse inner expression by calling p.parseExpression()
	// Check that currentToken is ')' - if not, add error and return nil
	// Call nextToken() to consume ')'
	// Return the inner expression (no need to wrap in special node)
	return nil
}
