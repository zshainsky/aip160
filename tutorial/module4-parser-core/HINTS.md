# Module 4 Hints

Stuck on implementation? Here are strategic hints to help you think through the problems.

## General Hints

### 🔍 Understanding the Two-Token Window

**Problem**: Why do we need both `currentToken` and `peekToken`?

**Hint**: Consider parsing `>=`. When you see `>`, you need to check if the next character is `=` before deciding if it's GREATER_THAN or GREATER_EQUAL.

**Also**: When parsing `name(`, you need to look ahead to see if `(` follows to decide if it's a function call or just an identifier.

### 🔄 The nextToken() Pattern

**Every parsing function should follow this pattern**:

1. Do something with `currentToken`
2. Call `nextToken()` when you're done with current token
3. Return the result

**Common mistake**: Forgetting to call `nextToken()` at the end!

## Task-Specific Hints

### Task 1: Parser Foundation

**Hint 1**: The `New()` constructor should call `nextToken()` twice. Why?
- After first call: `currentToken` is the first real token
- After second call: Both tokens are properly positioned

**Hint 2**: For `expectPeek()`:
```go
// Check if peek token matches expected type
// If yes: advance and return true
// If no: record error and return false
```

### Task 2: Literal Parsing

**Complete Solution for parseString()** (use as reference for other literals):
```go
func (p *Parser) parseString() ast.Expression {
    str := &ast.StringLiteral{
        Token: p.currentToken,
        Value: p.currentToken.Literal,
    }
    p.nextToken()
    return str
}
```

**Pattern**: Create AST node → Set fields → Advance token → Return

**Hint for parseNumber()**:
```go
// Need to convert string to float64
import "strconv"

num := &ast.NumberLiteral{
    Token: p.currentToken,
}

value, err := strconv.ParseFloat(p.currentToken.Literal, 64)
if err != nil {
    // Record error in p.errors with fmt.Sprintf
    return nil
}
num.Value = value

// Don't forget to call nextToken() before returning!
```

**Hint for parseBoolean()**:
```go
// Value is: true if token is TRUE, false if token is FALSE
Value: p.currentTokenIs(lexer.TRUE)
```

### Task 3: Expression Entry Point

**Hint**: These are simple! Each one just delegates to the next level:

```go
func (p *Parser) ParseProgram() *ast.Program {
    // Create Program node
    // Call parseExpression() to get the Expression
    // Return program
}

func (p *Parser) parseExpression() ast.Expression {
    // Just return parseOrExpression()
}
```

### Task 4: Logical Operators

**Hint for parseOrExpression()**:

Think about parsing `a OR b OR c`:

```
Step 1: left = parseAndExpression() → gets 'a'
Step 2: See OR token
Step 3: right = parseAndExpression() → gets 'b'
Step 4: left = LogicalExpression(a OR b)  ← Update left!
Step 5: See OR token again
Step 6: right = parseAndExpression() → gets 'c'
Step 7: left = LogicalExpression((a OR b) OR c)  ← Update left again!
Step 8: No more OR, return left
```

**Pattern**:
```go
left := /* get first operand */

for /* while operator matches */ {
    operator := p.currentToken
    p.nextToken()
    right := /* get next operand */
    
    left = &ast.LogicalExpression{
        Left: left,  // ← Previous result becomes left
        Operator: "...",
        Right: right,
    }
}

return left
```

**Hint for parseNotExpression()**:

NOT is optional and prefix:

```go
if /* current token is NOT */ {
    // Save token
    // Advance
    // Parse what comes after (the operand)
    // Return UnaryExpression
}

// No NOT found - just return next level
return p.parseComparison()
```

### Task 5: Comparisons

**Hint**: Comparison operator is also optional! Consider parsing just `active`:

```go
func (p *Parser) parseComparison() ast.Expression {
    left := p.parseValue()  // Gets "active"
    
    // Check if there's a comparison operator
    if p.isComparisonOperator(p.currentToken.Type) {
        // Build ComparisonExpression
    }
    
    // No operator - just return the value itself
    return left
}
```

**Why allow this?** For truthy checks: `active AND premium`

### Task 6: Value Dispatcher

**Hint**: This is a big switch statement:

```go
func (p *Parser) parseValue() ast.Expression {
    switch p.currentToken.Type {
    case lexer.STRING:
        return p.parseString()
    case lexer.NUMBER:
        return p.parseNumber()
    // ... more cases
    default:
        // Handle unexpected token
        p.errors = append(p.errors, ...)
        return nil
    }
}
```

**Remember**: Each case returns directly. No need for `p.nextToken()` here - the individual parse functions handle that!

### Task 7: Grouped Expressions

**Hint for parseGroupedExpression()**:

When you see `(`:
1. Consume the `(` by calling `nextToken()`
2. Parse whatever expression is inside: `parseExpression()` (allows full recursion!)
3. Expect and consume `)` using `expectPeek(lexer.RPAREN)`
4. Return the inner expression (no need to wrap in special node)

**Example trace for `(age > 18)`**:
```
1. See LPAREN
2. nextToken() → now at 'age'
3. parseExpression() → returns ComparisonExpression(age > 18)
4. expectPeek(RPAREN) → advances past ')'
5. Return the ComparisonExpression directly
```

## Debugging Tips

### Print Tokens

Add this to see token flow:

```go
func (p *Parser) nextToken() {
    fmt.Printf("ADVANCE: %s → %s\n", 
        p.currentToken.Type, p.peekToken.Type)
    p.currentToken = p.peekToken
    p.peekToken = p.lexer.NextToken()
}
```

### Print AST

Use the String() methods:

```go
program := p.ParseProgram()
fmt.Println(program.String())
```

### Check Errors

Always check parser errors in tests:

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

## Common Errors and Solutions

### Error: "expected next token to be ), got EOF"

**Cause**: Missing closing parenthesis in input, or parser didn't advance properly.

**Fix**: Check that every `parseGroupedExpression()` consumes the opening `(` before calling `parseExpression()`.

### Error: Wrong precedence (test expects different tree)

**Cause**: Functions calling each other in wrong order.

**Fix**: Remember the precedence chain:
```
parseOrExpression
  ↓ calls
parseAndExpression
  ↓ calls
parseNotExpression
  ↓ calls
parseComparison
  ↓ calls
parseValue
```

Each function should only call the **next level down**, not skip levels!

### Error: Infinite loop in parser

**Cause**: Forgot to call `nextToken()` somewhere, so parser keeps seeing same token.

**Fix**: Every path through your parsing functions should advance tokens. Add debug prints to see where it's stuck.

### Error: nil pointer dereference

**Cause**: Trying to use a token or node that's nil.

**Fix**: Check for nil before accessing:
```go
if program.Expression == nil {
    t.Fatal("program.Expression is nil")
}
```

## Testing Strategy

### Start Small

Test each function individually:

```bash
go test ./pkg/filter/parser -run TestLiteralParsing
```

### Build Up

Once literals work, test combinations:

```bash
go test ./pkg/filter/parser -run TestComparisons
go test ./pkg/filter/parser -run TestLogicalOperators
```

### Test Edge Cases

Don't forget:
- Empty input
- Bare identifiers (no comparison)
- Nested parentheses
- Multiple NOTs
- Long chains of ORs/ANDs

## Questions to Ask Yourself

When stuck, ask:

1. **What is my current token?** (print it!)
2. **What token do I expect next?** (peek at it!)
3. **Have I advanced past the current token?** (call nextToken!)
4. **Am I calling the right precedence level?** (check the chain!)
5. **What AST node should this create?** (look at ast.go!)

## Still Stuck?

If you've been stuck for more than 20 minutes on one issue:

1. Check [SOLUTION.md](SOLUTION.md) for the complete implementation
2. Compare your code with the solution
3. Understand **why** the solution works
4. Try implementing it yourself again

Remember: The goal is to **learn**, not to struggle endlessly. Looking at solutions is part of learning!

---

**Ready to see the complete code?** → [SOLUTION.md](SOLUTION.md)
