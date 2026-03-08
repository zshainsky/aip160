# Module 3: EBNF Grammar & Parser Design

**Duration**: ~2 hours  
**Difficulty**: Intermediate-Advanced

## 🎯 Module Objectives

By the end of this module, you will:
1. Understand EBNF (Extended Backus-Naur Form) notation
2. Read and interpret the AIP-160 filter grammar
3. Learn how recursive descent parsing works
4. Understand operator precedence in parsing
5. Design the parser structure before implementing it
6. Map grammar rules to Go functions

## 📖 What is EBNF?

**EBNF (Extended Backus-Naur Form)** is a notation for formally describing the syntax of a programming language or data format. It's like a blueprint that shows what valid expressions look like.

### Why Use EBNF?

1. **Precise**: No ambiguity about what's valid syntax
2. **Standardized**: Universally understood notation
3. **Blueprint for Parser**: Maps directly to parsing code
4. **Documentation**: Serves as the language specification

### Real-World Analogy

Think of EBNF like a recipe:
- **Recipe**: "Mix flour, add eggs, then bake"
- **EBNF**: "expression = term, ('+' | '-'), term"

Both describe the **steps** and **allowed ingredients**.

## 🔤 Reading EBNF Notation

Before diving into AIP-160's grammar, let's learn to read EBNF:

### Basic Symbols

```
=     "is defined as"
;     "end of rule"
|     "or" (alternatives)
,     "followed by" (sequence)
()    grouping
[]    optional (zero or one time)
{}    repetition (zero or more times)
" "   literal string
```

### Simple Examples

**Example 1: A boolean value**
```ebnf
boolean = "true" | "false" ;
```
**Translation**: A boolean is either the literal string "true" OR "false"

**Example 2: An optional sign with a number**
```ebnf
number = ["-"], digit, {digit} ;
digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
```
**Translation**: 
- A number is an optional "-" sign, followed by one digit, followed by zero or more additional digits
- So valid: `42`, `-5`, `123`

**Example 3: A list of items**
```ebnf
list = item, {",", item} ;
item = letter, {letter} ;
```
**Translation**: 
- A list is one item, followed by zero or more occurrences of (comma then item)
- Valid: `a`, `a,b`, `a,b,c`

### Practice Reading EBNF

```ebnf
greeting = "Hello", [","], name, ["!"] ;
name = letter, {letter | digit} ;
```

**What does this accept?**
- ✅ `Hello John`
- ✅ `Hello, John!`
- ✅ `Hello John!`
- ✅ `Hello user123`
- ❌ `Hello 123` (name can't start with digit)

## 📋 AIP-160 Filter Grammar

Here's the complete EBNF grammar for AIP-160 filters (simplified for core features):

```ebnf
(* Root rule *)
filter = expression ;

(* Expression with operator precedence *)
expression = or_expression ;

or_expression = and_expression, {"OR", and_expression} ;

and_expression = not_expression, {"AND", not_expression} ;

not_expression = ["NOT"], comparison ;

comparison = value, [comparator, value] ;

comparator = "=" | "!=" | "<" | ">" | "<=" | ">=" | ":" ;

(* Values and terminals *)
value = function_call
      | field
      | string
      | number
      | boolean
      | null
      | "(", expression, ")" ;

field = identifier, {".", identifier} ;

function_call = identifier, "(", [arg_list], ")" ;

arg_list = expression, {",", expression} ;

(* Literals *)
identifier = letter, {letter | digit | "_"} ;
string = '"', {character}, '"' | "'", {character}, "'" ;
number = ["-"], digit, {digit}, [".", {digit}], [("e" | "E"), ["-" | "+"], digit, {digit}] ;
boolean = "true" | "false" | "TRUE" | "FALSE" ;
null = "null" | "NULL" ;
```

### Key Observations

1. **Precedence is built into the grammar**:
   ```
   expression → or_expression → and_expression → not_expression → comparison
   ```
   Deeper in the grammar = higher precedence (evaluated first)

2. **Left-recursive patterns avoided**:
   ```ebnf
   or_expression = and_expression, {"OR", and_expression} ;
   ```
   This means: "one and_expression, followed by zero or more (OR then and_expression)"

3. **Parentheses for grouping**:
   ```ebnf
   value = ... | "(", expression, ")" ;
   ```
   Allows overriding precedence

## 🌳 How Grammar Maps to AST

Let's see how a filter expression maps through the grammar to an AST:

### Example: `age > 18 AND status = "active"`

**Step 1: Grammar rule matching**
```
filter
└─ expression
   └─ or_expression
      └─ and_expression (finds "AND"!)
         ├─ not_expression (left side)
         │  └─ comparison: age > 18
         └─ not_expression (right side)
            └─ comparison: status = "active"
```

**Step 2: AST construction**
```
LogicalExpression (AND)
├─ ComparisonExpression (age > 18)
│  ├─ Identifier("age")
│  └─ Number(18)
└─ ComparisonExpression (status = "active")
   ├─ Identifier("status")
   └─ String("active")
```

Notice how the grammar structure directly translates to the tree structure!

## 🔄 Recursive Descent Parsing

**Recursive descent** is a parsing technique where:
1. Each grammar rule becomes a function
2. Functions call each other recursively
3. The call stack naturally builds the AST

### Mapping Rules to Functions

```ebnf
expression = or_expression ;
```
↓
```go
func (p *Parser) parseExpression() ast.Expression {
    return p.parseOrExpression()
}
```

```ebnf
or_expression = and_expression, {"OR", and_expression} ;
```
↓
```go
func (p *Parser) parseOrExpression() ast.Expression {
    left := p.parseAndExpression()
    
    for p.currentToken.Type == lexer.OR {
        operator := p.currentToken
        p.nextToken()
        right := p.parseAndExpression()
        left = &ast.LogicalExpression{
            Left: left,
            Operator: "OR",
            Right: right,
        }
    }
    
    return left
}
```

### The Pattern

Every grammar rule follows this pattern:

1. **Sequence** (`,` in EBNF): Call other parsing functions in order
2. **Alternative** (`|` in EBNF): Use `switch` on token type
3. **Repetition** (`{}` in EBNF): Use `for` loop
4. **Optional** (`[]` in EBNF): Use `if` condition

## ⚖️ Understanding Operator Precedence

Operator precedence determines which operations happen first:

```
a OR b AND c    means    a OR (b AND c)
NOT a AND b     means    (NOT a) AND b
```

### Precedence Levels (highest to lowest)

```
1. Parentheses    ( )
2. Function calls timestamp(x)
3. Traversal      user.email
4. Unary          NOT, -
5. Comparison     =, !=, <, >, <=, >=, :
6. AND
7. OR
```

### Grammar Encodes Precedence

The grammar structure naturally implements precedence:

```
expression
  ↓ (calls)
or_expression      ← Lowest precedence (evaluated last)
  ↓ (calls)
and_expression     ← Higher precedence
  ↓ (calls)
not_expression     ← Even higher
  ↓ (calls)
comparison         ← Highest precedence (evaluated first)
```

**Why does this work?**

When parsing `a OR b AND c`:
1. `parseOrExpression()` is called first
2. It calls `parseAndExpression()` to get `a`
3. It sees `OR` and loops
4. It calls `parseAndExpression()` again for the right side
5. `parseAndExpression()` consumes `b AND c` as a unit!
6. Result: `a OR (b AND c)` ✅

### Visual Example

**Input**: `a OR b AND c OR d`

```
parseOrExpression()
├─ parseAndExpression() → returns 'a'
├─ sees OR
├─ parseAndExpression() → returns 'b AND c' (whole thing!)
├─ sees OR
└─ parseAndExpression() → returns 'd'

Result: ((a) OR (b AND c)) OR (d)
```

The deeper function (`parseAndExpression`) gets to "claim" its operators first!

## 🏗️ Parser Structure Design

Before implementing, let's design the Parser:

### Core Structure

```go
type Parser struct {
    lexer        *lexer.Lexer
    currentToken lexer.Token
    peekToken    lexer.Token
    errors       []string
}
```

**Fields explained:**
- `lexer`: Source of tokens
- `currentToken`: Token we're currently examining
- `peekToken`: Next token (lookahead for decisions)
- `errors`: Collect parsing errors instead of crashing

### Core Methods

```go
// Constructor
func New(l *lexer.Lexer) *Parser

// Token management
func (p *Parser) nextToken()
func (p *Parser) currentTokenIs(t lexer.TokenType) bool
func (p *Parser) peekTokenIs(t lexer.TokenType) bool
func (p *Parser) expectPeek(t lexer.TokenType) bool

// Entry point
func (p *Parser) ParseProgram() *ast.Program

// Grammar rule methods (one per rule)
func (p *Parser) parseExpression() ast.Expression
func (p *Parser) parseOrExpression() ast.Expression
func (p *Parser) parseAndExpression() ast.Expression
func (p *Parser) parseNotExpression() ast.Expression
func (p *Parser) parseComparison() ast.Expression
func (p *Parser) parseValue() ast.Expression
func (p *Parser) parseIdentifier() ast.Expression
func (p *Parser) parseNumber() ast.Expression
func (p *Parser) parseString() ast.Expression
// ... etc

// Error handling
func (p *Parser) Errors() []string
func (p *Parser) peekError(t lexer.TokenType)
```

### The Two-Token Window

The parser maintains a window of two tokens:

```
Tokens: [IDENT, EQUALS, STRING, AND, ...]
         ↑       ↑
    currentToken peekToken
```

**Why peek ahead?**

To make decisions without committing:
```go
// Is this "name" or "name.email" or "name(...)"?
if p.currentTokenIs(lexer.IDENTIFIER) {
    if p.peekTokenIs(lexer.DOT) {
        return p.parseTraversal()
    } else if p.peekTokenIs(lexer.LPAREN) {
        return p.parseFunctionCall()
    } else {
        return p.parseIdentifier()
    }
}
```

## 🎯 Parsing Algorithm Flow

Let's trace parsing `age > 18`:

```
1. ParseProgram()
   └─ calls parseExpression()

2. parseExpression()
   └─ calls parseOrExpression()

3. parseOrExpression()
   └─ calls parseAndExpression()

4. parseAndExpression()
   └─ calls parseNotExpression()

5. parseNotExpression()
   └─ no NOT, calls parseComparison()

6. parseComparison()
   ├─ calls parseValue() → returns Identifier("age")
   ├─ sees GREATER_THAN, saves operator
   ├─ calls parseValue() → returns Number(18)
   └─ returns ComparisonExpression

7. Control returns up the stack
   └─ returns final AST
```

Each function checks if its pattern matches, processes what it can, and returns!

## 💭 Key Concepts

### 1. **Top-Down Parsing**

We start at the top (root rule) and work down to terminals:
```
expression → or_expression → and_expression → ... → identifier
```

### 2. **Recursive Calls = Tree Structure**

When `parseOrExpression()` calls `parseAndExpression()` twice, it creates left and right children!

### 3. **Precedence via Call Order**

Higher precedence = deeper in the call chain = gets to parse first

### 4. **Error Recovery**

Instead of crashing on errors, collect them:
```go
func (p *Parser) peekError(t lexer.TokenType) {
    msg := fmt.Sprintf("expected next token to be %s, got %s instead",
        t, p.peekToken.Type)
    p.errors = append(p.errors, msg)
}
```

## 📝 Design Checklist

Before Module 4 (implementation), make sure you understand:

- [ ] How to read EBNF notation
- [ ] What each AIP-160 grammar rule means
- [ ] How grammar rules map to functions
- [ ] Why precedence is encoded in grammar structure
- [ ] How the two-token window works
- [ ] The relationship between recursion and tree building
- [ ] Why we use recursive descent over other parsing techniques

## 💡 Practice: Reading the Grammar

Try to answer these using the grammar:

**Q1**: Can you have multiple NOT operators?  
`NOT NOT active`

**Q2**: What's the precedence of `:` (has) operator?

**Q3**: Is `name.email.domain` valid?

**Q4**: Can you nest parentheses?  
`((a OR b) AND (c OR d))`

**Answers:**

**A1**: Yes! 
```ebnf
not_expression = ["NOT"], comparison
```
But wait, it only allows one NOT. To allow multiple, the grammar would need to be recursive: `not_expression = ["NOT"], not_expression | comparison`. The current grammar is simplified.

**A2**: Same as other comparison operators (=, !=, etc.) - they're all in the `comparator` rule at the same level.

**A3**: Yes!
```ebnf
field = identifier, {".", identifier}
```
This allows any number of dot-separated identifiers.

**A4**: Yes!
```ebnf
value = ... | "(", expression, ")" ;
```
Since it includes `expression`, and expression can contain more parentheses, they can nest infinitely!

## 🎓 Learning Checkpoint

Before moving to Module 4 (implementation), ensure you can:

- [ ] Read and explain any line of the EBNF grammar
- [ ] Trace how a filter string matches grammar rules
- [ ] Explain why AND has higher precedence than OR
- [ ] Describe how one grammar rule maps to a Go function
- [ ] Understand the purpose of currentToken and peekToken
- [ ] Explain how recursion builds the AST tree

## 📚 Resources

### Implementation Help

**Ready to implement?** → [Module 4: Parser Implementation](../module4-parser-core/README.md)

**Need hints for design?** → [HINTS.md](HINTS.md) - Design patterns and tips

**Want to see the full design?** → [SOLUTION.md](SOLUTION.md) - Complete parser architecture

### Additional Learning

- [Recursive Descent Parsing](https://en.wikipedia.org/wiki/Recursive_descent_parser)
- [EBNF Notation](https://en.wikipedia.org/wiki/Extended_Backus%E2%80%93Naur_form)
- [Operator Precedence](https://en.wikipedia.org/wiki/Order_of_operations)
- [Writing an Interpreter in Go](https://interpreterbook.com/) - Excellent book on this topic!

---

**Previous Module**: [Module 2: AST Design](../module2-ast/README.md)  
**Next Module**: [Module 4: Parser Implementation](../module4-parser-core/README.md)

**Note**: This module is primarily conceptual. Module 4 will be the hands-on implementation where you'll write the actual parser code!
