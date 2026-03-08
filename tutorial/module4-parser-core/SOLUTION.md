# Module 4 Complete Solution

This document contains the complete, working implementation for Module 4: Parser Core.

## Complete Implementation

### parser.go

```go
package parser

import (
	"fmt"
	"strconv"

	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// Parser performs parsing of filter expressions
type Parser struct {
	lexer  *lexer.Lexer
	errors []string

	currentToken lexer.Token
	peekToken    lexer.Token
}

// New creates a new Parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:  l,
		errors: []string{},
	}

	// Read two tokens to initialize current and peek
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances both the current and peek tokens
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// currentTokenIs checks if the current token matches the given type
func (p *Parser) currentTokenIs(t lexer.TokenType) bool {
	return p.currentToken.Type == t
}

// peekTokenIs checks if the peek token matches the given type
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of expected type and advances if so
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// Errors returns the list of parsing errors
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError adds an error when the peek token doesn't match expectations
func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// ParseProgram is the entry point for parsing a filter expression
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	
	// Skip any leading whitespace tokens if your lexer doesn't handle them
	if p.currentTokenIs(lexer.EOF) {
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
	left := p.parseAndExpression()

	for p.currentTokenIs(lexer.OR) {
		operator := p.currentToken
		p.nextToken()
		right := p.parseAndExpression()

		left = &ast.LogicalExpression{
			Token:    operator,
			Left:     left,
			Operator: "OR",
			Right:    right,
		}
	}

	return left
}

// parseAndExpression handles AND operators (higher precedence than OR)
// Grammar: and_expression = not_expression, {"AND", not_expression} ;
func (p *Parser) parseAndExpression() ast.Expression {
	left := p.parseNotExpression()

	for p.currentTokenIs(lexer.AND) {
		operator := p.currentToken
		p.nextToken()
		right := p.parseNotExpression()

		left = &ast.LogicalExpression{
			Token:    operator,
			Left:     left,
			Operator: "AND",
			Right:    right,
		}
	}

	return left
}

// parseNotExpression handles NOT operator (prefix, highest precedence)
// Grammar: not_expression = ["NOT"], comparison ;
func (p *Parser) parseNotExpression() ast.Expression {
	if p.currentTokenIs(lexer.NOT) {
		token := p.currentToken
		p.nextToken()
		
		// For Module 4, we'll allow only one NOT by calling parseComparison()
		// For multiple NOTs, we'd recursively call parseNotExpression()
		right := p.parseComparison()

		return &ast.UnaryExpression{
			Token:    token,
			Operator: "NOT",
			Right:    right,
		}
	}

	return p.parseComparison()
}

// parseComparison handles comparison operators
// Grammar: comparison = value, [comparator, value] ;
func (p *Parser) parseComparison() ast.Expression {
	left := p.parseValue()

	if p.isComparisonOperator(p.currentToken.Type) {
		operator := p.currentToken
		p.nextToken()
		right := p.parseValue()

		return &ast.ComparisonExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}

	// No comparison operator - return just the value
	return left
}

// isComparisonOperator checks if a token type is a comparison operator
func (p *Parser) isComparisonOperator(t lexer.TokenType) bool {
	switch t {
	case lexer.EQUALS, lexer.NOT_EQUALS,
		lexer.LESS_THAN, lexer.LESS_EQUAL,
		lexer.GREATER_THAN, lexer.GREATER_EQUAL:
		return true
	default:
		return false
	}
}

// parseValue dispatches to the appropriate parser based on token type
// Grammar: value = function_call | field | string | number | boolean | null | "(", expression, ")" ;
func (p *Parser) parseValue() ast.Expression {
	switch p.currentToken.Type {
	case lexer.STRING:
		return p.parseString()
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
	default:
		p.errors = append(p.errors,
			fmt.Sprintf("unexpected token %s in value position", p.currentToken.Type))
		return nil
	}
}

// parseString parses a string literal
func (p *Parser) parseString() ast.Expression {
	str := &ast.StringLiteral{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
	p.nextToken()
	return str
}

// parseNumber parses a numeric literal
func (p *Parser) parseNumber() ast.Expression {
	num := &ast.NumberLiteral{
		Token: p.currentToken,
	}

	value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
	if err != nil {
		p.errors = append(p.errors,
			fmt.Sprintf("could not parse %q as number", p.currentToken.Literal))
		return nil
	}
	num.Value = value

	p.nextToken()
	return num
}

// parseBoolean parses a boolean literal (true or false)
func (p *Parser) parseBoolean() ast.Expression {
	b := &ast.BooleanLiteral{
		Token: p.currentToken,
		Value: p.currentTokenIs(lexer.TRUE),
	}
	p.nextToken()
	return b
}

// parseNull parses a null literal
func (p *Parser) parseNull() ast.Expression {
	n := &ast.NullLiteral{
		Token: p.currentToken,
	}
	p.nextToken()
	return n
}

// parseIdentifier parses a simple identifier
// Module 5 will extend this to handle field traversal and function calls
func (p *Parser) parseIdentifier() ast.Expression {
	ident := &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
	p.nextToken()
	return ident
}

// parseGroupedExpression parses an expression wrapped in parentheses
// Grammar: "(", expression, ")"
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // consume '('

	expr := p.parseExpression()

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return expr
}
```

## Key Design Decisions

### 1. Two-Token Lookahead

**Why?**
- Need to distinguish `>` from `>=` 
- Need to check if identifier is followed by `(` for function calls (Module 5)
- Makes parsing logic cleaner

**How it works**:
```
Input: age > 18

Initial state after New():
  currentToken = "age" (IDENTIFIER)
  peekToken = ">" (GREATER_THAN)

After parseValue() processes "age":
  currentToken = ">" (GREATER_THAN)
  peekToken = "18" (NUMBER)
```

### 2. Error Collection vs. Panic

**We collect errors instead of panicking**:

```go
p.errors = append(p.errors, "error message")
return nil
```

**Benefits**:
- Report multiple errors at once
- User-friendly error messages
- Parser continues (can catch more errors)

**Tests should check errors**:
```go
if len(p.Errors()) > 0 {
    t.Fatalf("parser errors: %v", p.Errors())
}
```

### 3. Precedence Through Call Chain

**Magic!** By having each function call the next level:

```
parseOrExpression()
  → parseAndExpression()
    → parseNotExpression()
      → parseComparison()
        → parseValue()
```

**We automatically get correct precedence**:

For `a OR b AND c`:
1. parseOrExpression() gets 'a'
2. Sees OR
3. Calls parseAndExpression() for right side
4. parseAndExpression() **grabs both 'b AND c'** before returning!
5. Result: `a OR (b AND c)` ✅

### 4. Left-Associativity

**Our loop pattern builds left-associative trees**:

```go
for p.currentTokenIs(lexer.OR) {
    // ...
    left = &ast.LogicalExpression{
        Left: left,  // Previous result becomes left child
        // ...
    }
}
```

**For `a OR b OR c`**:
```
Iteration 1: left = (a OR b)
Iteration 2: left = ((a OR b) OR c)
```

This matches how most languages evaluate: left-to-right.

### 5. Optional Comparison

**Why allow comparison to be optional?**

```go
func (p *Parser) parseComparison() ast.Expression {
    left := p.parseValue()
    
    if p.isComparisonOperator(...) {
        // Build comparison
    }
    
    return left  // Return just the value if no operator
}
```

**Enables**: 
- Bare identifiers: `active AND premium`
- Truthy checks: if identifier evaluates to true
- Cleaner nesting in complex expressions

### 6. No Parentheses AST Node

**Parentheses guide parsing but don't need their own AST node**:

```go
func (p *Parser) parseGroupedExpression() ast.Expression {
    p.nextToken() // skip '('
    expr := p.parseExpression()
    p.expectPeek(lexer.RPAREN) // skip ')'
    return expr  // Return inner expression directly
}
```

**Why?** The tree structure already encodes the precedence!

`(a OR b) AND c` produces:
```
LogicalExpression(AND)
├─ LogicalExpression(OR)  ← Parentheses forced this to be a subtree
│  ├─ a
│  └─ b
└─ c
```

## Testing Strategy

### Test Structure

```go
func TestParseComparison(t *testing.T) {
    tests := []struct {
        input    string
        expected string  // Expected AST String() output
    }{
        {"age > 18", "(age > 18)"},
        {"name = \"John\"", "(name = \"John\")"},
        {"active != false", "(active != false)"},
    }
    
    for _, tt := range tests {
        l := lexer.New(tt.input)
        p := New(l)
        program := p.ParseProgram()
        
        checkParserErrors(t, p)
        
        if program.String() != tt.expected {
            t.Errorf("expected %q, got %q", tt.expected, program.String())
        }
    }
}
```

### Helper Function

```go
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
```

## Common Test Cases

### Literals
```go
"\"hello\""          → StringLiteral("hello")
"42"                 → NumberLiteral(42)
"3.14"               → NumberLiteral(3.14)
"true"               → BooleanLiteral(true)
"false"              → BooleanLiteral(false)
"null"               → NullLiteral
```

### Comparisons
```go
"age > 18"           → ComparisonExpression(age > 18)
"name = \"John\""    → ComparisonExpression(name = "John")
"active != false"    → ComparisonExpression(active != false)
```

### Logical Operators
```go
"a AND b"            → LogicalExpression(a AND b)
"a OR b"             → LogicalExpression(a OR b)
"NOT active"         → UnaryExpression(NOT active)
```

### Precedence
```go
"a OR b AND c"       → LogicalExpression(a OR (b AND c))
"a AND b OR c"       → LogicalExpression((a AND b) OR c)
"NOT a AND b"        → LogicalExpression((NOT a) AND b)
```

### Parentheses
```go
"(a OR b) AND c"     → LogicalExpression((a OR b) AND c)
"a AND (b OR c)"     → LogicalExpression(a AND (b OR c))
"((a))"              → Just 'a' (parentheses don't create nodes)
```

### Complex
```go
"age >= 21 AND (status = \"premium\" OR trial = true)"
→ LogicalExpression(
    (age >= 21)
    AND
    ((status = "premium") OR (trial = true))
  )
```

## Debugging Tips

### 1. Print Token Flow

```go
func (p *Parser) nextToken() {
    fmt.Printf("ADVANCE: %s → %s\n", p.currentToken.Literal, p.peekToken.Literal)
    p.currentToken = p.peekToken
    p.peekToken = p.lexer.NextToken()
}
```

### 2. Print AST Structure

```go
program := p.ParseProgram()
fmt.Printf("AST: %s\n", program.String())
```

### 3. Print Call Stack

Add to each parser function:

```go
func (p *Parser) parseOrExpression() ast.Expression {
    fmt.Println("→ parseOrExpression")
    defer fmt.Println("← parseOrExpression")
    // ... rest of function
}
```

### 4. Check Errors Immediately

```go
program := p.ParseProgram()
if len(p.Errors()) > 0 {
    for _, err := range p.Errors() {
        fmt.Println("ERROR:", err)
    }
}
```

## What's Next?

Module 5 will extend this parser to support:

1. **Field Traversal**: `user.email.domain`
2. **Has Operator**: `tags:urgent`
3. **Function Calls**: `now()`, `timestamp(created)`
4. **Negative Numbers**: `-42`, `-3.14`
5. **Wildcards**: `name = "*John*"`

The foundation you've built here will make those additions straightforward!

---

**Congratulations!** You now have a working recursive descent parser that correctly handles operator precedence. 🎉
