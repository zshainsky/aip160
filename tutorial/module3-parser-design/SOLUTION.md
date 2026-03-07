# Module 3 Complete Parser Design

This document provides the complete parser architecture and design patterns. Module 4 will implement this design.

## Complete Parser Structure

```go
package parser

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

type Parser struct {
	lexer        *lexer.Lexer
	errors       []string
	
	currentToken lexer.Token
	peekToken    lexer.Token
}

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

// Token management
func (p *Parser) nextToken() {
	p.currentToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) currentTokenIs(t lexer.TokenType) bool {
	return p.currentToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// Error handling
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// Entry point
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Expression = p.parseExpression()
	return program
}
```

## Grammar-to-Code Mapping

### Complete Function Signatures

```go
// Expression hierarchy (precedence handling)
func (p *Parser) parseExpression() ast.Expression
func (p *Parser) parseOrExpression() ast.Expression
func (p *Parser) parseAndExpression() ast.Expression
func (p *Parser) parseNotExpression() ast.Expression
func (p *Parser) parseComparison() ast.Expression

// Values and terminals
func (p *Parser) parseValue() ast.Expression
func (p *Parser) parseIdentifierOrField() ast.Expression
func (p *Parser) parseFunctionCall(name *ast.Identifier) ast.Expression
func (p *Parser) parseTraversal(left ast.Expression) ast.Expression
func (p *Parser) parseGroupedExpression() ast.Expression
func (p *Parser) parseString() ast.Expression
func (p *Parser) parseNumber() ast.Expression
func (p *Parser) parseBoolean() ast.Expression
func (p *Parser) parseNull() ast.Expression

// Helpers
func (p *Parser) isComparisonOperator(t lexer.TokenType) bool
```

## Detailed Function Designs

### 1. parseExpression() - Entry Point

```ebnf
expression = or_expression ;
```

```go
func (p *Parser) parseExpression() ast.Expression {
	return p.parseOrExpression()
}
```

**Design**: Simple delegation. This allows future expansion if we add expression-level features.

### 2. parseOrExpression() - Lowest Precedence

```ebnf
or_expression = and_expression, {"OR", and_expression} ;
```

```go
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
```

**Design Pattern**: 
- Start with higher-precedence expression (AND)
- Loop while seeing OR operators
- Build left-associative tree: `((a OR b) OR c)`
- Update `left` each iteration to build the chain

**Why left-associative?**
```
a OR b OR c
→ (a OR b) <- left
   OR c    <- combine with left again
→ ((a OR b) OR c)
```

### 3. parseAndExpression() - Higher Precedence

```ebnf
and_expression = not_expression, {"AND", not_expression} ;
```

```go
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
```

**Design**: Identical pattern to OR, but calls NOT expression (next level down).

**Precedence magic**: By calling `parseNotExpression()`, AND "gets first dibs" on tokens:
```
a OR b AND c
→ parseOr() sees 'a', then OR
→ parseOr() calls parseAnd() for right side
→ parseAnd() takes 'b AND c' completely!
→ Result: a OR (b AND c) ✅
```

### 4. parseNotExpression() - Unary Operator

```ebnf
not_expression = ["NOT"], comparison ;
```

```go
func (p *Parser) parseNotExpression() ast.Expression {
	if p.currentTokenIs(lexer.NOT) {
		token := p.currentToken
		p.nextToken()
		right := p.parseNotExpression() // Recursive! Allows multiple NOTs
		
		return &ast.UnaryExpression{
			Token:    token,
			Operator: "NOT",
			Right:    right,
		}
	}
	
	return p.parseComparison()
}
```

**Design Notes**:
- Optional prefix operator
- Recursive call allows `NOT NOT x` (though unusual)
- Falls through to comparison if no NOT

**Alternative design (single NOT only)**:
```go
if p.currentTokenIs(lexer.NOT) {
	// ...
	right := p.parseComparison() // Non-recursive
}
```

### 5. parseComparison() - The Meat

```ebnf
comparison = value, [comparator, value] ;
comparator = "=" | "!=" | "<" | ">" | "<=" | ">=" | ":" ;
```

```go
func (p *Parser) parseComparison() ast.Expression {
	left := p.parseValue()
	
	if p.isComparisonOperator(p.currentToken.Type) {
		operator := p.currentToken
		p.nextToken()
		right := p.parseValue()
		
		// Handle HAS operator specially
		if operator.Type == lexer.HAS {
			return &ast.HasExpression{
				Token:      operator,
				Collection: left,
				Member:     right,
			}
		}
		
		return &ast.ComparisonExpression{
			Token:    operator,
			Left:     left,
			Operator: operator.Literal,
			Right:    right,
		}
	}
	
	// No comparison operator, just return the value
	return left
}

func (p *Parser) isComparisonOperator(t lexer.TokenType) bool {
	switch t {
	case lexer.EQUALS, lexer.NOT_EQUALS,
		lexer.LESS_THAN, lexer.LESS_EQUAL,
		lexer.GREATER_THAN, lexer.GREATER_EQUAL,
		lexer.HAS:
		return true
	default:
		return false
	}
}
```

**Design Decisions**:
- Optional comparison (might just be a value)
- Special handling for HAS (`:`) - different AST node
- Falls back to returning just the value if no operator

**Why allow bare values?**
For filters like `active` (truthy check) or inside function calls: `matches(name)`

### 6. parseValue() - The Dispatcher

```ebnf
value = function_call
      | field
      | string
      | number
      | boolean
      | null
      | "(", expression, ")" ;
```

```go
func (p *Parser) parseValue() ast.Expression {
	switch p.currentToken.Type {
	case lexer.IDENTIFIER:
		return p.parseIdentifierOrField()
	case lexer.STRING:
		return p.parseString()
	case lexer.NUMBER:
		return p.parseNumber()
	case lexer.TRUE, lexer.FALSE:
		return p.parseBoolean()
	case lexer.NULL:
		return p.parseNull()
	case lexer.LPAREN:
		return p.parseGroupedExpression()
	case lexer.MINUS:
		return p.parseNegation()
	default:
		p.errors = append(p.errors, 
			fmt.Sprintf("unexpected token %s", p.currentToken.Type))
		return nil
	}
}
```

**Design**: Big dispatcher based on current token type.

**Why IDENTIFIER goes to special function?**
Because `name` could be:
- Just an identifier: `name`
- A field traversal: `name.email`
- A function call: `name()`

We need to look ahead!

### 7. parseIdentifierOrField() - Lookahead Decision

```go
func (p *Parser) parseIdentifierOrField() ast.Expression {
	ident := &ast.Identifier{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
	
	p.nextToken() // Move past identifier
	
	// Check what comes after
	if p.currentTokenIs(lexer.DOT) {
		return p.parseTraversal(ident)
	}
	
	if p.currentTokenIs(lexer.LPAREN) {
		return p.parseFunctionCall(ident)
	}
	
	// Just a simple identifier
	return ident
}
```

**Design Pattern**: Read identifier, then look at current token to decide:
- `.` → It's a traversal
- `(` → It's a function call
- Anything else → Just an identifier

### 8. parseTraversal() - Chaining Fields

```ebnf
field = identifier, {".", identifier} ;
```

```go
func (p *Parser) parseTraversal(left ast.Expression) ast.Expression {
	for p.currentTokenIs(lexer.DOT) {
		token := p.currentToken
		p.nextToken()
		
		if !p.currentTokenIs(lexer.IDENTIFIER) {
			p.errors = append(p.errors, "expected identifier after '.'")
			return left
		}
		
		right := &ast.Identifier{
			Token: p.currentToken,
			Value: p.currentToken.Literal,
		}
		
		p.nextToken()
		
		left = &ast.TraversalExpression{
			Token: token,
			Left:  left,
			Right: right,
		}
	}
	
	return left
}
```

**Design Pattern**:
- Loop while seeing dots
- Build left-associative tree
- Each iteration wraps the previous result

**Example**: `user.address.city`
```
Iteration 1: user.address
    TraversalExpr
    ├─ user
    └─ address

Iteration 2: (user.address).city
    TraversalExpr
    ├─ TraversalExpr
    │  ├─ user
    │  └─ address
    └─ city
```

### 9. parseFunctionCall() - Argument Lists

```ebnf
function_call = identifier, "(", [arg_list], ")" ;
arg_list = expression, {",", expression} ;
```

```go
func (p *Parser) parseFunctionCall(name *ast.Identifier) ast.Expression {
	fn := &ast.FunctionCall{
		Token:     name.Token,
		Function:  name.Value,
		Arguments: []ast.Expression{},
	}
	
	p.nextToken() // consume '('
	
	// Empty argument list?
	if p.currentTokenIs(lexer.RPAREN) {
		p.nextToken()
		return fn
	}
	
	// Parse first argument
	fn.Arguments = append(fn.Arguments, p.parseExpression())
	
	// Parse remaining arguments
	for p.currentTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		fn.Arguments = append(fn.Arguments, p.parseExpression())
	}
	
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	
	return fn
}
```

**Design Pattern**:
- Handle empty case first
- Parse first argument
- Loop for remaining comma-separated arguments
- Each argument is a full expression (allows nesting!)

**Examples**:
- `now()` - Empty args
- `timestamp(created_at)` - One arg
- `compare(age, 18)` - Multiple args
- `func(a OR b, c AND d)` - Complex args!

### 10. parseGroupedExpression() - Parentheses

```ebnf
"(", expression, ")"
```

```go
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken() // consume '('
	
	expr := p.parseExpression()
	
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	
	return expr
}
```

**Design**: 
- Consume `(`
- Parse inner expression (allows full recursion!)
- Expect `)`
- Return the inner expression directly (parentheses don't need AST node)

**Why no ParenthesizedExpression node?**
Parentheses are for precedence control during parsing. Once we have the AST, the tree structure already represents the precedence!

### 11. Literal Parsing - Simple Cases

```go
func (p *Parser) parseString() ast.Expression {
	str := &ast.StringLiteral{
		Token: p.currentToken,
		Value: p.currentToken.Literal,
	}
	p.nextToken()
	return str
}

func (p *Parser) parseNumber() ast.Expression {
	num := &ast.NumberLiteral{
		Token: p.currentToken,
	}
	
	// Parse the string to float64
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

func (p *Parser) parseBoolean() ast.Expression {
	b := &ast.BooleanLiteral{
		Token: p.currentToken,
		Value: p.currentTokenIs(lexer.TRUE),
	}
	p.nextToken()
	return b
}

func (p *Parser) parseNull() ast.Expression {
	n := &ast.NullLiteral{
		Token: p.currentToken,
	}
	p.nextToken()
	return n
}
```

**Pattern**: Create AST node, optionally convert value, advance token, return.

## Complete Parse Flow Example

Let's trace parsing: `age > 18 AND status = "active"`

```
1. ParseProgram()
   └─ parseExpression()

2. parseExpression()
   └─ parseOrExpression()

3. parseOrExpression()
   ├─ parseAndExpression()  ← Enters here
   │
   4. parseAndExpression()
      ├─ parseNotExpression()  ← For left side
      │
      5. parseNotExpression()
         ├─ (no NOT, skip)
         └─ parseComparison()
         
         6. parseComparison()
            ├─ parseValue() → Identifier("age")
            ├─ sees GREATER_THAN
            ├─ parseValue() → Number(18)
            └─ returns ComparisonExpression(age > 18)
      
      ← Returns ComparisonExpression to parseAndExpression
      
      7. parseAndExpression()
         ├─ left = ComparisonExpression(age > 18)
         ├─ sees AND token!
         ├─ loops: operator = AND
         └─ calls parseNotExpression() for right side
         
         8. parseNotExpression()
            └─ parseComparison()
            
            9. parseComparison()
               ├─ parseValue() → Identifier("status")
               ├─ sees EQUALS
               ├─ parseValue() → String("active")
               └─ returns ComparisonExpression(status = "active")
         
         10. parseAndExpression() builds LogicalExpression
             ├─ left: ComparisonExpression(age > 18)
             ├─ operator: "AND"
             └─ right: ComparisonExpression(status = "active")
         
         11. Loop check: current token not AND, exits loop
         
← Returns to parseOrExpression

12. parseOrExpression()
    ├─ Has full LogicalExpression(AND)
    ├─ Checks: current token is OR? No (it's EOF)
    ├─ Exits loop
    └─ returns LogicalExpression

Final AST:
LogicalExpression (AND)
├─ ComparisonExpression (age > 18)
│  ├─ Identifier("age")
│  └─ Number(18)
└─ ComparisonExpression (status = "active")
   ├─ Identifier("status")
   └─ String("active")
```

## Design Principles Summary

### 1. **Separation of Concerns**

Each function handles ONE grammar rule:
- `parseOrExpression()` handles OR logic
- `parseAndExpression()` handles AND logic
- `parseComparison()` handles comparisons
- etc.

### 2. **Recursive Structure Mirrors Grammar**

Grammar structure:
```
expression → or → and → not → comparison → value
```

Function calls:
```go
parseExpression() → parseOr() → parseAnd() → parseNot() → parseComparison() → parseValue()
```

### 3. **Precedence Through Call Depth**

Deeper functions get to parse first:
- `parseValue()` is deepest → highest precedence
- `parseOrExpression()` is shallowest → lowest precedence

### 4. **Lookahead for Disambiguation**

Use `peekToken` to decide between alternatives:
- `name` vs `name.field` vs `name()`
- Checked with `peekTokenIs()`

### 5. **Error Recovery**

Don't panic! Collect errors and try to continue:
```go
p.errors = append(p.errors, errorMessage)
return nil // or best-effort result
```

### 6. **Left Associativity**

Binary operators build left-associative trees:
```
a OR b OR c → ((a OR b) OR c)
```

Implemented by updating `left` in the loop:
```go
for seesOperator {
    left = buildNode(left, operator, right)
}
```

## Common Patterns Recap

### Pattern 1: Simple Rule
```ebnf
rule = other_rule ;
```
```go
func parseRule() {
    return parseOtherRule()
}
```

### Pattern 2: Sequence
```ebnf
rule = a, b, c ;
```
```go
func parseRule() {
    a := parseA()
    b := parseB()
    c := parseC()
    return combine(a, b, c)
}
```

### Pattern 3: Optional
```ebnf
rule = required, [optional] ;
```
```go
func parseRule() {
    req := parseRequired()
    if currentTokenMatches {
        opt := parseOptional()
    }
    return build(req, opt)
}
```

### Pattern 4: Repetition (Zero or More)
```ebnf
rule = first, {",", more} ;
```
```go
func parseRule() {
    result := parseFirst()
    for currentTokenIs(COMMA) {
        nextToken()
        more := parseMore()
        result = combine(result, more)
    }
    return result
}
```

### Pattern 5: Alternatives
```ebnf
rule = optionA | optionB | optionC ;
```
```go
func parseRule() {
    switch currentToken.Type {
    case TYPE_A: return parseA()
    case TYPE_B: return parseB()
    case TYPE_C: return parseC()
    }
}
```

---

## What's Next?

**Module 4** will implement this complete design! You'll:
1. Create the Parser struct
2. Implement all parsing functions
3. Write comprehensive tests
4. Handle edge cases and errors

The design is complete. Now it's time to code it!

---

**Previous Module**: [Module 2: AST Design](../module2-ast/README.md)  
**Next Module**: [Module 4: Parser Implementation](../module4-parser/README.md) (coming soon)

**Review concepts?** Go back to [README.md](README.md) or [HINTS.md](HINTS.md)
