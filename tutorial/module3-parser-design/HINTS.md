# Module 3 Parser Design Hints

These hints will help you understand EBNF grammar and parser design concepts before implementing in Module 4.

## Understanding EBNF - Guided Practice

### Hint 1: Reading Basic EBNF

When you see EBNF notation, break it down symbol by symbol:

```ebnf
greeting = "Hello", name, ["!"] ;
```

**Read it as**: 
- `greeting =` → "A greeting consists of..."
- `"Hello"` → the literal word "Hello"
- `,` → followed by
- `name` → a name (defined by another rule)
- `["!"]` → optionally followed by an exclamation mark
- `;` → end of rule

**Practice**: Before looking at complex rules, identify each symbol:
- What's literal? (in quotes)
- What's a reference to another rule? (no quotes)
- What's optional? (in `[]`)
- What repeats? (in `{}`)
- What are alternatives? (separated by `|`)

### Hint 2: Decoding Repetition `{}`

```ebnf
list = item, {",", item} ;
```

**Think of it as**:
- Start with one `item`
- Then zero or more times: comma followed by item
- Examples: `a`, `a,b`, `a,b,c,d,e`

**Pattern**: `first, {separator, more}` is how lists are typically written

**Why not just `{item}`?**
Because that would allow an empty list! By requiring at least one item outside the `{}`, we ensure "one or more".

### Hint 3: Understanding Alternatives `|`

```ebnf
comparator = "=" | "!=" | "<" | ">" | "<=" | ">=" ;
```

**Think of it as**: "Pick exactly ONE of these options"

```ebnf
value = string | number | boolean | identifier ;
```

**Think of it as**: "A value can be any one of these types"

**Parsing implication**: Use a `switch` statement!
```go
switch p.currentToken.Type {
case lexer.STRING:
    return p.parseString()
case lexer.NUMBER:
    return p.parseNumber()
case lexer.TRUE, lexer.FALSE:
    return p.parseBoolean()
// ...
}
```

---

## Grammar Structure Hints

### Hint 4: Tracing Through the Grammar

Let's trace `age > 18`:

```
filter
└─ expression (what does this do?)
   └─ or_expression (look for OR? Not found, moves on)
      └─ and_expression (look for AND? Not found, moves on)
         └─ not_expression (look for NOT? Not found, moves on)
            └─ comparison (HERE! Finds: value, >, value)
```

**Practice**: Trace these yourself:
- `name = "John"`
- `a AND b`
- `NOT active`
- `user.email`

### Hint 5: Operator Precedence Mystery Solved

**Question**: Why does this grammar make AND higher precedence than OR?

```ebnf
or_expression = and_expression, {"OR", and_expression} ;
and_expression = not_expression, {"AND", not_expression} ;
```

**Think about it**: 
- To parse an OR expression, you first must parse an AND expression
- The AND expression "consumes" its operators before returning
- So `a OR b AND c` becomes: OR expr sees `a`, then sees `OR`, then asks AND expr for right side
- AND expr takes `b AND c` completely!
- Result: `a OR (b AND c)`

**Key insight**: Deeper in the grammar = tighter binding = higher precedence

### Hint 6: Left vs Right Recursion

**Left-recursive (BAD for recursive descent)**:
```ebnf
expression = expression, "+", term ;  (* Calls itself first! *)
```

**Why bad?** 
```go
func parseExpression() {
    parseExpression()  // Infinite recursion!
    // ...never gets here
}
```

**Right-recursive or iterative (GOOD)**:
```ebnf
expression = term, {"+", term} ;  (* Iteration pattern *)
```

**Why good?**
```go
func parseExpression() {
    result := parseTerm()  // Parse first, no recursion yet
    for /* see plus */ {
        parseTerm()  // Controlled recursion
    }
}
```

---

## Recursive Descent Hints

### Hint 7: Pattern for Sequence (,)

When EBNF shows: `rule = partA, partB, partC ;`

**Translate to**:
```go
func parseRule() *ast.RuleNode {
    a := parsePartA()
    b := parsePartB()
    c := parsePartC()
    return &ast.RuleNode{A: a, B: b, C: c}
}
```

**Just call the functions in order!**

### Hint 8: Pattern for Optional ([])

When EBNF shows: `rule = required, [optional] ;`

**Translate to**:
```go
func parseRule() *ast.RuleNode {
    required := parseRequired()
    
    var optional *ast.OptionalNode
    if p.currentToken.Type == OPTIONAL_TYPE {
        optional = parseOptional()
    }
    
    return &ast.RuleNode{Required: required, Optional: optional}
}
```

**Check if the token is there, if so, parse it!**

### Hint 9: Pattern for Repetition ({})

When EBNF shows: `rule = first, {",", more} ;`

**Translate to**:
```go
func parseRule() *ast.RuleNode {
    items := []ast.Expression{parseFirst()}
    
    for p.currentToken.Type == COMMA {
        p.nextToken()  // consume comma
        items = append(items, parseMore())
    }
    
    return &ast.RuleNode{Items: items}
}
```

**Use a loop! Check for the separator/operator each time**

### Hint 10: Pattern for Alternatives (|)

When EBNF shows: `rule = optionA | optionB | optionC ;`

**Translate to**:
```go
func parseRule() ast.Expression {
    switch p.currentToken.Type {
    case TYPE_A:
        return parseOptionA()
    case TYPE_B:
        return parseOptionB()
    case TYPE_C:
        return parseOptionC()
    default:
        // error: unexpected token
    }
}
```

**Use a switch! Check the current token type**

---

## Parser Structure Hints

### Hint 11: The Two-Token Window

Think of it like reading with your finger and your eyes:

```
Tokens:  [IDENT] [EQUALS] [STRING] [AND] ...
            ↑       ↑
         current   peek
         (finger) (eyes)
```

**Why two tokens?**
- **Current**: What we're processing now
- **Peek**: What comes next (helps make decisions)

**Example decision**:
```go
if p.currentToken.Type == IDENT {
    if p.peekToken.Type == DOT {
        // It's a traversal: name.field
    } else if p.peekToken.Type == LPAREN {
        // It's a function call: name(...)
    } else {
        // It's just an identifier: name
    }
}
```

### Hint 12: nextToken() Pattern

Every parse function should consume the tokens it needs:

```go
func parseComparison() ast.Expression {
    left := parseValue()  // consumes left side tokens
    
    if isComparisonOp(p.currentToken) {
        operator := p.currentToken
        p.nextToken()  // ← IMPORTANT! Move past operator
        right := parseValue()  // consumes right side tokens
        
        return &ast.ComparisonExpression{
            Left: left,
            Operator: operator.Literal,
            Right: right,
        }
    }
    
    return left
}
```

**Rule**: If you use a token, advance past it with `nextToken()`!

### Hint 13: Error Handling Strategy

Don't panic! Collect errors:

```go
type Parser struct {
    // ...
    errors []string
}

func (p *Parser) peekError(expected lexer.TokenType) {
    msg := fmt.Sprintf("expected %s, got %s", expected, p.peekToken.Type)
    p.errors = append(p.errors, msg)
}
```

**Why?**
- Shows multiple errors at once (better user experience)
- Parser can try to continue and find more errors
- Tests can verify error messages

---

## Precedence Hints

### Hint 14: Building Precedence Table

From the grammar, extract precedence (highest to lowest):

```
1. Primary (literals, identifiers, parentheses)
2. NOT (unary)
3. Comparison (=, !=, <, >, <=, >=, :)
4. AND
5. OR
```

**Parsing order**: Reverse! Parse OR first, which calls AND, which calls NOT, etc.

**Why reverse?** 
- Lowest precedence should be at the top initially
- It calls down to higher precedence
- Higher precedence "gets first dibs" on tokens

### Hint 15: Precedence via Grammar Depth

Count how deep each operation is in the grammar:

```
expression (depth 1)
└─ or_expression (depth 2)      ← Shallowest = lowest precedence
   └─ and_expression (depth 3)
      └─ not_expression (depth 4)
         └─ comparison (depth 5) ← Deepest = highest precedence
```

**Mental model**: Think of the grammar as a nested Russian doll!

---

## Common Conceptual Pitfalls

### Pitfall 1: Confusing Grammar Order with Parse Order

**Grammar definition order** (how rules are written) doesn't matter.

What matters is **the rule call chain**:
```
parseExpression() 
  → calls parseOr()
    → calls parseAnd()
      → calls parseNot()
```

### Pitfall 2: Thinking Precedence is "Hardcoded"

Precedence isn't hardcoded with if-statements. It's **structural**:

❌ **Not like this**:
```go
if operator == "OR" {
    precedence = 1
} else if operator == "AND" {
    precedence = 2
}
```

✅ **Like this**:
```go
parseOr() {
    left = parseAnd()  // AND goes deeper first!
    // handle OR
}
```

### Pitfall 3: Forgetting That Grammar is Documentation

The grammar is your **specification**. If you can't parse something, check:
1. Is it valid according to the grammar?
2. If not, should the grammar be updated?
3. If yes, is my parser correctly following the grammar?

---

## Self-Check Questions

Before Module 4, test yourself:

**Q1**: What does this EBNF mean?
```ebnf
arg_list = expression, {",", expression} ;
```

**Q2**: Why is this bad for recursive descent?
```ebnf
expr = expr, "+", term ;
```

**Q3**: In the AIP-160 grammar, which is deeper: AND or NOT?

**Q4**: If you see this token sequence: `IDENT DOT IDENT`, what are you parsing?

**Q5**: What's the purpose of `peekToken`?

---

## Answers to Self-Check

**A1**: An argument list is one expression, followed by zero or more occurrences of (comma then expression). Examples: `x`, `x, y`, `x, y, z`

**A2**: Left recursion! The rule calls itself before doing anything else, causing infinite recursion in a recursive descent parser.

**A3**: NOT is deeper (called by and_expression), so NOT has higher precedence.

**A4**: A traversal expression like `user.email`. You'd call `parseTraversalExpression()` or similar.

**A5**: To look ahead one token and make decisions without committing. Helps distinguish between `name`, `name.field`, and `name()`.

---

**Ready for implementation?** Move on to [Module 4: Parser Implementation](../module4-parser/README.md)

**Need deeper explanation?** Check [SOLUTION.md](SOLUTION.md) for detailed design documentation

**Review concepts?** Go back to [README.md](README.md)
