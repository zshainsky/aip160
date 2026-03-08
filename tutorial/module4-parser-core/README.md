# Module 4: Building the Parser (Core)

**Duration**: ~2-3 hours  
**Difficulty**: Intermediate-Advanced

## 🎯 Module Objectives

By the end of this module, you will:
1. Implement the complete parser structure with two-token lookahead
2. Build core parsing functions for expressions, comparisons, and values
3. Handle operator precedence correctly (OR < AND < NOT < Comparison)
4. Parse literals: strings, numbers, booleans, null
5. Implement error handling and recovery
6. Write a recursive descent parser from scratch

## 📋 Prerequisites

Before starting this module, you should have completed:
- ✅ Module 1: Lexer (tokenization)
- ✅ Module 2: AST Design
- ✅ Module 3: EBNF & Parser Design

Make sure you understand:
- How the lexer produces tokens
- The AST node types and their structure
- How grammar rules map to functions
- Why precedence matters in parsing

## 🏗️ What You'll Build

In this module, you'll implement the **core parser** that handles:

### ✅ Supported in Module 4:
- **Literals**: `"hello"`, `42`, `3.14`, `true`, `false`, `null`
- **Identifiers**: `name`, `age`, `status`
- **Comparisons**: `age > 18`, `name = "John"`, `status != "active"`
- **Logical operators**: `a AND b`, `x OR y`, `NOT active`
- **Parentheses**: `(a OR b) AND c`
- **Operator precedence**: `a OR b AND c` → `a OR (b AND c)`

### ⏭️ Deferred to Module 5:
- Field traversal: `user.email.domain`
- Has operator: `list:value`
- Function calls: `now()`, `timestamp(created)`
- Negative numbers: `-42`

## 🎓 Recursive Descent Refresher

Quick reminder of the key concept:

**Each grammar rule = One function**

```ebnf
expression = or_expression ;
```
↓
```go
func (p *Parser) parseExpression() ast.Expression {
    return p.parseOrExpression()
}
```

**Precedence = Call order (higher precedence called deeper)**

```
parseExpression()
  └─ parseOrExpression()      ← Lowest precedence (evaluated last)
      └─ parseAndExpression()   ← Higher precedence
          └─ parseNotExpression() ← Even higher
              └─ parseComparison()  ← Highest (evaluated first)
```

## 🏗️ Parser Structure

### Step 1: The Parser Type

```go
type Parser struct {
    lexer        *lexer.Lexer  // Token provider
    errors       []string       // Collected errors
    
    currentToken lexer.Token    // Current token being processed
    peekToken    lexer.Token    // Next token (for lookahead)
}
```

**Why two tokens?**
To handle two-character operators and make decisions:
- `>=` vs `>`
- `!=` vs `!`
- Is this a function call `name(` or just identifier `name`?

### Step 2: Initialization

```go
func New(l *lexer.Lexer) *Parser {
    p := &Parser{
        lexer:  l,
        errors: []string{},
    }
    
    // Prime the pump: read two tokens
    p.nextToken()
    p.nextToken()
    
    return p
}
```

**Initial state**:
- First call to `nextToken()`: `currentToken = nil`, `peekToken = first token`
- Second call: `currentToken = first token`, `peekToken = second token`

Now we're ready to parse!

### Step 3: Token Management Helpers

```go
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
```

**Pattern**: These helpers make parsing code cleaner:
- Instead of: `if p.currentToken.Type == lexer.NUMBER`
- Write: `if p.currentTokenIs(lexer.NUMBER)`

## � Getting Started

### Starting Your Implementation

This module includes a starter file to help you begin:

**Copy the starter file to begin:**
```bash
cp tutorial/module4-parser-core/parser_starter.go pkg/filter/parser/parser.go
```

The starter file includes:
- ✅ Complete `Parser` struct definition
- ✅ Precedence chain **already implemented** as pass-throughs for testing
- ✅ `ParseProgram()` and `parseExpression()` fully working
- 📝 TODO comments marking what you'll implement in each task

**What the pass-throughs mean:**
Functions like `parseOrExpression()`, `parseAndExpression()`, etc. are already implemented as simple delegations. This lets you test basic parsing immediately! You'll enhance them with full logic in later tasks.

**What you'll implement:**
- Task 1: Foundation functions (`New()`, `nextToken()`, helpers)
- Task 2: First literal (`parseString()`) + minimal dispatcher
- Tasks 3-8: Add more literals, then enhance pass-throughs with full operator logic
- Task 9: Complete with grouped expressions

### Testing as You Go

After each task, run the specified tests:
```bash
# Task 1
go test ./pkg/filter/parser -run TestParserCreation

# Task 2
go test ./pkg/filter/parser -run TestStringLiteral

# And so on...
```

### If You Get Stuck

- 💡 [HINTS.md](HINTS.md) - Complete `parseString()` example and strategic hints
- ✅ [SOLUTION.md](SOLUTION.md) - Full working implementation with design explanations

## �🔨 Implementation Guide

This module is designed to build incrementally. Each task adds functionality that you can test before moving to the next.

### Task 1: Parser Foundation (15 min)

The `Parser` struct is **already implemented** for you! Now implement these core methods:

**Constructor and Token Management:**
- `New()` - Create parser, initialize errors, call `nextToken()` twice
- `nextToken()` - Move peek to current, get next token from lexer
- `currentTokenIs()` - Check if current token matches type
- `peekTokenIs()` - Check if peek token matches type
- `expectPeek()` - Check peek, advance if matches, error if not

**Error Handling:**
- `Errors()` - Return the errors slice
- `peekError()` - Add error message to errors slice

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestParserCreation
```

### Task 2: Build Chain + First Literal (25 min)

Build the precedence chain as **pass-throughs** AND implement your first literal. This gives you immediate testable feedback!

**Part A: Implement the precedence chain (just delegate for now):**
- `ParseProgram()` - Create `&ast.Program{}`, handle EOF (return empty program), set `program.Expression = p.parseExpression()`, return program
- `parseExpression()` - Just `return p.parseOrExpression()`
- `parseOrExpression()` - Just `return p.parseAndExpression()` (Task 8 will enhance with OR loop)
- `parseAndExpression()` - Just `return p.parseNotExpression()` (Task 7 will enhance with AND loop)
- `parseNotExpression()` - Just `return p.parseComparison()` (Task 6 will enhance with NOT handling)
- `parseComparison()` - Just `return p.parseValue()` (Task 5 will enhance with comparisons)

**Part B: Implement your first literal:**
- `parseString()` - Create `&ast.StringLiteral{Token: p.currentToken, Value: p.currentToken.Literal}`, call `p.nextToken()`, return it
- `parseValue()` - Minimal dispatcher with just the STRING case:
  ```go
  switch p.currentToken.Type {
  case lexer.STRING:
      return p.parseString()
  default:
      msg := fmt.Sprintf("unexpected token: %s", p.currentToken.Type)
      p.errors = append(p.errors, msg)
      return nil
  }
  ```

**Why?** Now you have a complete path from `ParseProgram()` down to a literal. You can parse `"hello"` end-to-end!

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestStringLiteral
```

### Task 3: Add Remaining Literals (20 min)

Now add the other literal types. You already have the pattern from `parseString()`!

**Functions to implement:**
- `parseNumber()` - Create `NumberLiteral`, use `strconv.ParseFloat(p.currentToken.Literal, 64)` to convert, handle errors
- `parseBoolean()` - Create `BooleanLiteral`, set `Value` to `true` if token is TRUE, `false` otherwise
- `parseNull()` - Create `NullLiteral` with token
- `parseIdentifier()` - Create `Identifier` with token and literal value

**Pattern for each** (same as parseString):
1. Create the AST node with `currentToken`
2. Extract/convert the value if needed (numbers need `strconv.ParseFloat`)
3. Call `p.nextToken()` to advance
4. Return the node

**Important for `parseNumber()`**: Handle parse errors by appending to `p.errors` and returning `nil`

**Update `parseValue()`**: Add cases for `NUMBER`, `TRUE`/`FALSE`, `NULL`, and `IDENTIFIER` to your switch statement.

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestNumberLiteral
go test ./pkg/filter/parser -run TestBooleanLiteral
go test ./pkg/filter/parser -run TestNullLiteral
go test ./pkg/filter/parser -run TestIdentifier
```

### Task 4: Comparisons (25 min)

**Enhance `parseComparison()`** (replace the pass-through with full logic):
1. Get left value by calling `parseValue()`
2. Check if current token is a comparison operator using `isComparisonOperator()`
3. If yes: save operator, advance, get right value, return `ComparisonExpression`
4. If no: return just the left value (allows bare identifiers like `active`)

**Implement `isComparisonOperator()`**:
Use a switch statement to check if the token is:
- `EQUALS`, `NOT_EQUALS`
- `LESS_THAN`, `LESS_EQUAL`
- `GREATER_THAN`, `GREATER_EQUAL`

Return `true` if it matches, `false` otherwise.

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestSimpleComparison
go test ./pkg/filter/parser -run TestBareIdentifierInExpression
```

### Task 6: NOT Expression (15 min)

**Update `parseNotExpression()`** (replace the pass-through):
1. Check if currentToken is NOT
2. If yes:
   - Save token
   - Call `nextToken()`
   - Get operand by recursively calling `p.parseNotExpression()`
   - Return `UnaryExpression` with operator "NOT" and right = operand
3. If no: return `p.parseComparison()`

### Task 5: NOT Expression (15 min)

**Enhance `parseNotExpression()`** (replace the pass-through with full logic):
1. Check if currentToken is NOT
2. If yes:
   - Save token
   - Call `nextToken()`
   - Get operand by recursively calling `p.parseNotExpression()`
   - Return `UnaryExpression` with operator "NOT" and right = operand
3. If no: return `p.parseComparison()`

**Hint**: Recursive call allows `NOT NOT active`

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestNotExpression
go test ./pkg/filter/parser -run TestMultipleNOTs
```

### Task 6: AND Expression (20 min)

**Enhance `parseAndExpression()`** (replace the pass-through with a loop):
1. Get left operand by calling `p.parseNotExpression()`
2. Loop while `currentToken` is AND:
   - Save operator token
   - Call `nextToken()`
   - Get right operand by calling `p.parseNotExpression()`
   - Create `LogicalExpression` with left, operator "AND", right
   - **Important**: Update `left` to be this new expression
3. Return `left`

This builds a left-associative tree: `a AND b AND c` becomes `((a AND b) AND c)`

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestLogicalAnd
```

### Task 7: OR Expression (20 min)

**Enhance `parseOrExpression()`** (replace the pass-through with a loop):
1. Get left operand by calling `p.parseAndExpression()`
2. Loop while `currentToken` is OR:
   - Save operator token
   - Call `nextToken()`
   - Get right operand by calling `p.parseAndExpression()`
   - Create `LogicalExpression` with left, operator "OR", right
   - **Important**: Update `left` to be this new expression
3. Return `left`

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestLogicalOr
go test ./pkg/filter/parser -run TestOperatorPrecedence
```

### Task 8: Grouped Expressions (15 min)

**Implement `parseGroupedExpression()`**:
1. Call `nextToken()` to consume the opening `(`
2. Parse the inner expression with `parseExpression()` (allows full recursion!)
3. Check that `currentToken` is `)` - if not, add error and return nil
4. Call `nextToken()` to consume the closing `)`
5. Return the inner expression directly

**Update `parseValue()`**: Add the `LPAREN` case to call `parseGroupedExpression()`

**Tests to pass**:
```bash
go test ./pkg/filter/parser -run TestGroupedExpression
go test ./pkg/filter/parser -run TestComplexExpression
```

### Task 9: Final Verification (5 min)

Run all tests to verify everything works together:

```bash
go test ./pkg/filter/parser -v
```

All tests should pass! 🎉

## 🔍 Testing Your Parser

Run all parser tests:

```bash
# Run all tests
go test ./pkg/filter/parser -v

# Run specific test
go test ./pkg/filter/parser -run TestParserCreation -v

# Run with coverage
go test ./pkg/filter/parser -cover
```

### Example Test Case

```go
func TestSimpleComparison(t *testing.T) {
    input := `age > 18`
    
    l := lexer.New(input)
    p := New(l)
    program := p.ParseProgram()
    
    checkParserErrors(t, p)
    
    if program.Expression == nil {
        t.Fatal("program.Expression is nil")
    }
    
    comp, ok := program.Expression.(*ast.ComparisonExpression)
    if !ok {
        t.Fatalf("expected ComparisonExpression, got %T", program.Expression)
    }
    
    if comp.Operator != ">" {
        t.Errorf("expected operator '>', got %s", comp.Operator)
    }
}
```

## 💡 Common Pitfalls

### 1. Forgetting to Advance Tokens

❌ **Wrong**:
```go
func (p *Parser) parseString() ast.Expression {
    return &ast.StringLiteral{
        Token: p.currentToken,
        Value: p.currentToken.Literal,
    }
    // Forgot p.nextToken()!
}
```

✅ **Correct**:
```go
func (p *Parser) parseString() ast.Expression {
    str := &ast.StringLiteral{
        Token: p.currentToken,
        Value: p.currentToken.Literal,
    }
    p.nextToken() // ✓ Always advance!
    return str
}
```

### 2. Wrong Precedence Order

❌ **Wrong** (AND has lower precedence):
```go
func (p *Parser) parseOrExpression() ast.Expression {
    left := p.parseComparison() // Too deep!
    // ...
}
```

✅ **Correct**:
```go
func (p *Parser) parseOrExpression() ast.Expression {
    left := p.parseAndExpression() // ✓ Next level down
    // ...
}
```

### 3. Not Handling Optional Cases

For `comparison = value, [comparator, value]`, the `[]` means optional:

```go
func (p *Parser) parseComparison() ast.Expression {
    left := p.parseValue()
    
    // Don't assume there's always an operator!
    if p.isComparisonOperator(p.currentToken.Type) {
        // ... handle comparison
    }
    
    // Return just the value if no operator
    return left
}
```

## 📊 Tracing Parse Execution

To understand how your parser works, trace this example:

**Input**: `age > 18 AND active`

```
ParseProgram()
└─ parseExpression()
   └─ parseOrExpression()
      └─ parseAndExpression()
         ├─ left: parseNotExpression()
         │  └─ parseComparison()
         │     ├─ left: parseValue() → Identifier("age")
         │     ├─ operator: GREATER_THAN
         │     └─ right: parseValue() → Number(18)
         │     Returns: ComparisonExpression(age > 18)
         │
         ├─ sees AND token, loops
         │
         └─ right: parseNotExpression()
            └─ parseComparison()
               └─ parseValue() → Identifier("active")
               Returns: Identifier("active")
         
         Returns: LogicalExpression(
            left: ComparisonExpression(age > 18),
            operator: "AND",
            right: Identifier("active")
         )
```

## 📝 Implementation Checklist

Before moving to Module 5:

- [ ] Parser struct with currentToken and peekToken
- [ ] New() constructor that primes the pump
- [ ] Token management helpers (nextToken, currentTokenIs, etc.)
- [ ] Error handling (Errors(), peekError())
- [ ] ParseProgram() entry point
- [ ] parseExpression() → delegates to parseOrExpression()
- [ ] parseOrExpression() → handles OR with loops
- [ ] parseAndExpression() → handles AND with loops
- [ ] parseNotExpression() → handles optional NOT prefix
- [ ] parseComparison() → handles comparison operators
- [ ] parseValue() → dispatcher to specific parsers
- [ ] parseString(), parseNumber(), parseBoolean(), parseNull()
- [ ] parseIdentifier() → simple identifier
- [ ] parseGroupedExpression() → handles parentheses
- [ ] All tests pass!

## 🎯 Success Criteria

Your parser should handle:

```go
// Simple comparisons
"age > 18"
"name = \"John\""
"active = true"

// Logical operators
"age > 18 AND status = \"active\""
"premium = true OR trial = true"
"NOT inactive"

// Precedence
"a OR b AND c"           // → a OR (b AND c)
"NOT x AND y"            // → (NOT x) AND y

// Parentheses
"(a OR b) AND c"         // Override precedence
"((a OR b) AND (c OR d))" // Nested parens

// Complex combinations
"age >= 21 AND (status = \"premium\" OR trial = true)"
```

## 📚 Resources

**Next Steps**: [Module 5: Advanced Features](../module5-advanced-features/README.md) (coming soon)

**Need Help?**
- [HINTS.md](HINTS.md) - Nudges in the right direction
- [SOLUTION.md](SOLUTION.md) - Complete working implementation

**Additional Reading**:
- [Operator Precedence Parsing](https://en.wikipedia.org/wiki/Operator-precedence_parser)
- [Recursive Descent](https://www.engr.mun.ca/~theo/Misc/exp_parsing.htm)
- [Writing an Interpreter in Go](https://interpreterbook.com/)

---

**Previous Module**: [Module 3: Parser Design](../module3-parser-design/README.md)  
**Next Module**: [Module 5: Advanced Features](../module5-advanced-features/README.md)

**Happy Parsing! 🚀**
