package parser

import (
	"fmt"
	"strconv"
	"strings"

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
	// Initialize parser and errors slice
	p := &Parser{
		lexer:  l,
		errors: make([]string, 0),
	}
	// Call nextToken() twice to prime the pump
	p.nextToken()
	p.nextToken()
	return p
}

// nextToken advances both the current and peek tokens
func (p *Parser) nextToken() {
	// Move peekToken to currentToken
	p.currentToken = p.peekToken
	// Get next token from lexer and set as peekToken
	p.peekToken = p.lexer.NextToken()
}

// currentTokenIs checks if the current token matches the given type
func (p *Parser) currentTokenIs(t lexer.TokenType) bool {
	// Compare p.currentToken.Type with t
	return p.currentToken.Type == t
}

// peekTokenIs checks if the peek token matches the given type
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	// Compare p.peekToken.Type with t
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of expected type and advances if so
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	// Check if peekToken matches type
	if p.peekTokenIs(t) {
		// If yes: call nextToken() and return true
		p.nextToken()
		return true
	}
	// If no: call peekError(t) and return false
	p.peekError(t)
	return false
}

// Errors returns the list of parsing errors
func (p *Parser) Errors() []string {
	// Return the errors slice
	return p.errors
}

// peekError adds an error when the peek token doesn't match expectations
func (p *Parser) peekError(t lexer.TokenType) {
	// TODO: Create error message: "expected next token to be X, got Y instead"
	err := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken)
	// TODO: Append to p.errors
	p.errors = append(p.errors, err)
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
	// (Task 7): Enhance with OR operator loop
	// Get left operand from parseAndExpression()
	// While current token is OR:
	//   - Save the operator token
	//   - Advance to next token
	//   - Get right operand from parseAndExpression()
	//   - Create LogicalExpression node
	//   - Update left to be the new expression (for left-associativity)
	// Return left
	left := p.parseAndExpression()
	for p.currentTokenIs(lexer.OR) {
		opToken := p.currentToken
		p.nextToken()
		right := p.parseAndExpression()
		left = &ast.LogicalExpression{
			Token:    opToken,
			Operator: opToken.Literal,
			Left:     left,
			Right:    right,
		}
	}
	return left
}

// parseAndExpression handles AND operators (higher precedence than OR)
// Grammar: and_expression = not_expression, {"AND", not_expression} ;
func (p *Parser) parseAndExpression() ast.Expression {
	// (Task 6): Enhance with AND operator loop
	// Similar pattern to OR expression but:
	//   - Get left operand from parseNotExpression()
	//   - Loop while current token is AND
	//   - Create LogicalExpression nodes
	//   - Build left-associative chain
	left := p.parseNotExpression()
	for p.currentTokenIs(lexer.AND) {
		opToken := p.currentToken
		p.nextToken()
		right := p.parseNotExpression()
		left = &ast.LogicalExpression{
			Token:    opToken,
			Operator: opToken.Literal,
			Left:     left,
			Right:    right,
		}
	}
	return left
}

// parseNotExpression handles NOT operator (prefix, highest precedence)
// Grammar: not_expression = ["NOT"], comparison ;
func (p *Parser) parseNotExpression() ast.Expression {
	// (Task 5): Enhance with NOT operator handling
	// Check if current token is NOT
	//   - If yes: save token, advance, recursively call parseNotExpression(),
	//     create UnaryExpression node, return it
	//   - If no: return parseComparison()
	// Hint: Recursive call allows "NOT NOT active"
	if p.currentTokenIs(lexer.NOT) {
		tok := p.currentToken
		p.nextToken()
		return &ast.UnaryExpression{
			Token:    tok,
			Operator: tok.Literal,
			Right:    p.parseNotExpression(),
		}
	}
	return p.parseComparison()
}

// parseComparison handles comparison operators
// Grammar: comparison = value, [comparator, value] ;
func (p *Parser) parseComparison() ast.Expression {
	// (Task 4): Enhance with comparison operator handling
	// Get left value from parseValue()
	// Check if current token is a comparison operator (use isComparisonOperator)
	//   - If yes: save operator, advance, get right value,
	//     create ComparisonExpression node, return it
	//   - If no: return just the left value (allows bare identifiers)

	// TODO (Module 5): For proper precedence, call parseHasExpression() instead of parseValue()
	// Has operator sits between comparison and value in precedence
	left := p.parseHasExpression()
	if p.isComparisonOperator(p.currentToken.Type) {
		opToken := p.currentToken
		p.nextToken()
		right := p.parseHasExpression()

		return &ast.ComparisonExpression{
			Token:    opToken,
			Left:     left,
			Operator: opToken.Literal,
			Right:    right,
		}
	}
	return left
}

// isComparisonOperator checks if a token type is a comparison operator
func (p *Parser) isComparisonOperator(t lexer.TokenType) bool {
	// (Task 4): Implement comparison operator check
	// Use switch statement to check if t is one of:
	switch t {
	case lexer.EQUALS, lexer.NOT_EQUALS, lexer.LESS_THAN, lexer.LESS_EQUAL, lexer.GREATER_THAN, lexer.GREATER_EQUAL:
		return true
	}
	return false
}

// parseValue dispatches to the appropriate parser based on token type
// Grammar: value = function_call | field | string | number | boolean | null | "(", expression, ")" | "-", number ;
func (p *Parser) parseValue() ast.Expression {
	// (Task 2): Implement minimal dispatcher
	// Start with just STRING case to get first test passing:
	switch p.currentToken.Type {
	case lexer.STRING:
		return p.parseString()
	// (Task 3): Add remaining literal cases
	case lexer.NUMBER:
		return p.parseNumber()
	case lexer.TRUE, lexer.FALSE:
		return p.parseBoolean()
	case lexer.NULL:
		return p.parseNull()
	case lexer.IDENTIFIER:
		return p.parseIdentifier()
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	case lexer.MINUS:
		// Handle unary minus: negative literals or negation (Cycle 7A)
		return p.parseMinus()
	default:
		msg := fmt.Sprintf("unexpected token: %s", p.currentToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
}

// parseString parses a string literal
func (p *Parser) parseString() ast.Expression {
	// (Task 2): Implement string literal parsing
	// Pattern: Create AST node → Set fields → Advance token → Return
	lit := &ast.StringLiteral{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
	p.nextToken()
	return lit
}

// parseNumber parses a numeric literal
func (p *Parser) parseNumber() ast.Expression {
	// (Task 3): Implement number literal parsing
	// Create NumberLiteral with currentToken
	lit := &ast.NumberLiteral{Token: p.currentToken}

	// Use strconv.ParseFloat(p.currentToken.Literal, 64) to convert
	val, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	// If error, append to p.errors and return nil
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as number", p.currentToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	// Set num.Value to the parsed value
	lit.Value = val
	// Call nextToken() to advance
	p.nextToken()
	// Return the NumberLiteral
	return lit
}

// parseBoolean parses a boolean literal (true or false)
func (p *Parser) parseBoolean() ast.Expression {
	// (Task 3): Implement boolean literal parsing
	// Create BooleanLiteral with currentToken
	lit := &ast.BooleanLiteral{Token: p.currentToken, Value: false}
	// Set Value to true if currentToken is TRUE, false otherwise
	if strings.ToLower(p.currentToken.Literal) == "true" {
		lit.Value = true
	}
	// Call nextToken() to advance
	p.nextToken()
	// Return the BooleanLiteral
	return lit
}

// parseNull parses a null literal
func (p *Parser) parseNull() ast.Expression {
	// (Task 3): Implement null literal parsing
	// Create NullLiteral with currentToken
	lit := &ast.NullLiteral{Token: p.currentToken}
	// Call nextToken() to advance
	p.nextToken()
	// Return the NullLiteral
	return lit
}

// parseIdentifier parses a simple identifier
// Module 5 will extend this to handle field traversal and function calls
func (p *Parser) parseIdentifier() ast.Expression {
	// (Task 3): Implement identifier parsing
	// Create Identifier with currentToken and its Literal value
	identToken := p.currentToken
	identLiternal := p.currentToken.Literal
	// Call nextToken() to advance
	p.nextToken()

	// TODO (Module 5 - Task 3): Add function call detection
	// Check if current token (after advancing past identifier) is LPAREN
	// If yes, this is a function call - call parseFunctionCall helper
	// Pass the saved identifier token and value to the helper
	// Function calls bind tightest, so check this BEFORE checking for DOT
	if p.currentTokenIs(lexer.LPAREN) {
		return p.parseFunctionCall(identToken, identLiternal)
	}

	// TODO (Module 5 - Task 1): Add field traversal support
	// After handling function calls, check if current token is DOT
	// If yes, enter loop to build TraversalExpression chain
	var left ast.Expression = &ast.Identifier{
		Token: identToken,
		Value: identLiternal,
	}
	for p.currentTokenIs(lexer.DOT) {
		dotToken := p.currentToken
		p.nextToken()

		// Expect an identifier after the dot
		if !p.currentTokenIs(lexer.IDENTIFIER) {
			p.errors = append(p.errors, fmt.Sprintf("expected identifier after '.', got %s", p.currentToken.Type))
			return left
		}

		right := &ast.Identifier{Token: p.currentToken, Value: p.currentToken.Literal}
		p.nextToken()

		left = &ast.TraversalExpression{
			Token: dotToken,
			Left:  left,
			Right: right,
		}
	}

	// Return the built expression (either simple Identifier or TraversalExpression chain)
	return left
}

// TODO (Module 5 - Task 2): Implement parseHasExpression
// This function handles the has operator (:) for collection membership
// Should be called from parseComparison instead of parseValue
// Pattern similar to parseComparison:
//   - Get left from parseValue()
//   - Check if current token is HAS
//   - If yes: save token, advance, get right from parseValue(), create HasExpression
//   - If no: return left as-is
//
// Remember: both sides can be complex (traversals, etc.)
func (p *Parser) parseHasExpression() ast.Expression {
	left := p.parseValue()
	if p.currentTokenIs(lexer.HAS) {
		hasToken := p.currentToken
		p.nextToken() // consume ':'
		right := p.parseValue()
		return &ast.HasExpression{
			Token:      hasToken,
			Collection: left,
			Member:     right,
		}
	}

	return left
}

// TODO (Module 5 - Task 3): Implement parseFunctionCall
// This helper parses function calls: identifier followed by (arguments)
// Parameters: identToken (saved token), functionName (string)
// Steps:
//   - Advance past LPAREN (current token when called)
//   - Check for empty args (immediate RPAREN)
//   - Parse first argument using parseExpression()
//   - Loop while current is COMMA: advance, parse next expression
//   - Expect RPAREN at end
//   - Return FunctionCall with saved token, function name, and arguments slice
// See HINTS.md for argument parsing pattern

func (p *Parser) parseFunctionCall(identToken lexer.Token, functionName string) ast.Expression {
	p.nextToken() // advanced past LPAREN
	if p.currentTokenIs(lexer.RPAREN) {
		// empty args
		p.nextToken()
		return &ast.FunctionCall{
			Token:     identToken,
			Function:  functionName,
			Arguments: []ast.Expression{},
		}
	}
	arg := p.parseExpression()
	args := []ast.Expression{arg}
	for p.currentTokenIs(lexer.COMMA) {
		p.nextToken()
		nextArg := p.parseExpression()
		args = append(args, nextArg)
	}
	if !p.currentTokenIs(lexer.RPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("expected closing parenthesis but found %s", p.currentToken.Literal))
		return nil
	}
	p.nextToken() // consume RPAREN
	return &ast.FunctionCall{
		Token:     identToken,
		Function:  functionName,
		Arguments: args,
	}
}

// parseGroupedExpression parses an expression wrapped in parentheses
// Grammar: "(", expression, ")"
func (p *Parser) parseGroupedExpression() ast.Expression {
	// (Task 8): Implement grouped expression parsing
	// Call nextToken() to consume '('
	p.nextToken()
	// Parse inner expression by calling p.parseExpression()
	expression := p.parseExpression()
	// Check that currentToken is ')' - if not, add error and return nil
	if !p.currentTokenIs(lexer.RPAREN) {
		p.errors = append(p.errors, fmt.Sprintf("expected closing parenthesis but found %s", p.currentToken.Literal))
		return nil
	}
	// Call nextToken() to consume ')'
	p.nextToken()
	// Return the inner expression (no need to wrap in special node)
	return expression
}

// parseMinus dispatches between negative number literals and negation operators (Cycle 7A)
// The MINUS token can represent two distinct operations per AIP-160:
// 1. Negative number literal: -5, -3.14, -1.5e-3 (when followed by NUMBER)
// 2. Negation operator: -active, -enabled (shorthand for NOT, when followed by expression)
func (p *Parser) parseMinus() ast.Expression {
	minusToken := p.currentToken
	p.nextToken()

	// Dispatch based on what follows the MINUS
	if p.currentTokenIs(lexer.NUMBER) {
		return p.parseNegativeLiteral(minusToken)
	}

	return p.parseNegation(minusToken)
}

// parseNegativeLiteral parses a negative number literal: -5, -3.14, -1.5e-3 (Cycle 7A)
// Returns a UnaryExpression with operator "-" and a NumberLiteral as right child
func (p *Parser) parseNegativeLiteral(minusToken lexer.Token) ast.Expression {
	numLit := p.parseNumber()
	if numLit == nil {
		return nil
	}

	return &ast.UnaryExpression{
		Token:    minusToken,
		Operator: "-",
		Right:    numLit,
	}
}

// parseNegation parses a negation operator: -active, -enabled (Cycle 7A)
// Per AIP-160, - is shorthand for NOT operator
// Returns a UnaryExpression with operator "-" and the negated expression as right child
func (p *Parser) parseNegation(minusToken lexer.Token) ast.Expression {
	right := p.parseValue()
	if right == nil {
		return nil
	}

	return &ast.UnaryExpression{
		Token:    minusToken,
		Operator: "-",
		Right:    right,
	}
}
