# Module 2 AST Implementation Hints

These hints will guide you through implementing the AST node types without giving away the complete solution. Try each hint level before moving to the next!

## General Pattern

All node types follow the same three-method pattern:

```go
func (n *NodeType) expressionNode() {}           // Empty - just a marker

func (n *NodeType) TokenLiteral() string {
    return n.Token.Literal                       // Return token's literal
}

func (n *NodeType) String() string {
    // This is where the work is!
    // Return a string representation of the node
}
```

Focus your effort on the `String()` method for each type.

---

## Hint 1: NumberLiteral

**Goal:** Represent numbers in string form

**Level 1 - What to return:**
- The String() method should return the number as a string
- Use `fmt.Sprintf()` to format the float64 value

**Level 2 - Format specifier:**
- Use `%g` format specifier for numbers
- This automatically chooses between fixed and scientific notation
- Example: `fmt.Sprintf("%g", 42.0)` → `"42"`

**Level 3 - Implementation pattern:**
```go
func (n *NumberLiteral) String() string {
    return fmt.Sprintf("%g", n.Value)
}
```

---

## Hint 2: BooleanLiteral

**Goal:** Represent true/false values

**Level 1 - What to check:**
- Look at the `Value` field (it's a bool)
- Return "true" or "false" accordingly

**Level 2 - Simple conditional:**
```go
if n.Value {
    return "true"
} else {
    return "false"
}
```

**Level 3 - Even simpler:**
- Go's `fmt` package can format booleans directly with `%t`
- Or use `fmt.Sprintf("%v", n.Value)` which works for any type

---

## Hint 3: NullLiteral

**Goal:** Represent null values

**Level 1 - The simplest one:**
- Just return the string `"null"`
- No fields to check, no formatting needed

**Level 2 - One-liner:**
```go
func (n *NullLiteral) String() string {
    return "null"
}
```

---

## Hint 4: ComparisonExpression

**Goal:** Represent comparisons like `age > 18`

**Level 1 - What components do you have?**
- Left expression (e.g., `age`)
- Operator (e.g., `>`)
- Right expression (e.g., `18`)

**Level 2 - How to get strings from child nodes:**
- Call `.String()` on `c.Left` and `c.Right`
- These are Expression interfaces, so they have the String() method

**Level 3 - Format with parentheses:**
- Use format: `"(<left> <operator> <right>)"`
- Example: `"(age > 18)"`

**Level 4 - Implementation pattern:**
```go
func (c *ComparisonExpression) String() string {
    return fmt.Sprintf("(%s %s %s)", 
        c.Left.String(), 
        c.Operator, 
        c.Right.String())
}
```

---

## Hint 5: LogicalExpression

**Goal:** Represent logical operations like `expr1 AND expr2`

**Level 1 - Similar to ComparisonExpression:**
- Same structure: Left, Operator, Right
- Same approach: recursively call String() on children

**Level 2 - Same pattern:**
```go
return fmt.Sprintf("(%s %s %s)", 
    l.Left.String(), 
    l.Operator, 
    l.Right.String())
```

**Note:** The format is identical to ComparisonExpression! The difference is in what the fields contain, not how we format them.

---

## Hint 6: UnaryExpression

**Goal:** Represent unary operations like `NOT active` or `-5`

**Level 1 - What's different about unary?**
- Only ONE operand (the Right field)
- Operator comes BEFORE the operand
- Still use parentheses

**Level 2 - Format pattern:**
- Use format: `"(<operator><right>)"`
- Examples: `"(NOT active)"`, `"(-5)"`

**Level 3 - Implementation:**
```go
func (u *UnaryExpression) String() string {
    return fmt.Sprintf("(%s%s)", 
        u.Operator, 
        u.Right.String())
}
```

---

## Hint 7: TraversalExpression

**Goal:** Represent field traversal like `user.email`

**Level 1 - What's the syntax?**
- Dot notation: `<left>.<right>`
- NO parentheses for this one!
- Example: `user.email`, not `(user.email)`

**Level 2 - Both sides are expressions:**
- Call String() on both Left and Right
- Join with a dot: `"."`

**Level 3 - Implementation:**
```go
func (t *TraversalExpression) String() string {
    return fmt.Sprintf("%s.%s", 
        t.Left.String(), 
        t.Right.String())
}
```

**Note:** This naturally handles nested traversals like `user.address.city` because Left can be another TraversalExpression!

---

## Hint 8: HasExpression

**Goal:** Represent the has operator like `tags:urgent`

**Level 1 - What's the syntax?**
- Colon notation: `<collection>:<member>`
- NO parentheses
- Example: `tags:urgent`

**Level 2 - Similar to TraversalExpression:**
- Two parts joined by a symbol
- Just use `:` instead of `.`

**Level 3 - Implementation:**
```go
func (h *HasExpression) String() string {
    return fmt.Sprintf("%s:%s", 
        h.Collection.String(), 
        h.Member.String())
}
```

---

## Hint 9: FunctionCall

**Goal:** Represent function calls like `timestamp(created_at)` or `duration(start, end)`

**Level 1 - What components?**
- Function name (string)
- List of arguments ([]Expression)
- Format: `name(arg1, arg2, ...)`

**Level 2 - Handle the argument list:**
- Need to convert each Expression to a string
- Join them with `", "` (comma and space)
- Use `strings.Join()` for joining

**Level 3 - Build argument strings:**
```go
var args []string
for _, arg := range f.Arguments {
    args = append(args, arg.String())
}
```

**Level 4 - Join and format:**
```go
argStr := strings.Join(args, ", ")
return fmt.Sprintf("%s(%s)", f.Function, argStr)
```

**Level 5 - Complete implementation:**
```go
func (f *FunctionCall) String() string {
    var args []string
    for _, arg := range f.Arguments {
        args = append(args, arg.String())
    }
    return fmt.Sprintf("%s(%s)", f.Function, strings.Join(args, ", "))
}
```

**Edge case:** Empty argument list results in `"name()"` which is correct!

---

## Implementation Order Suggestion

Start with the easiest and build up:

1. **NullLiteral** - Just return "null" (warmup)
2. **BooleanLiteral** - Simple conditional
3. **NumberLiteral** - Basic formatting
4. **ComparisonExpression** - First binary expression
5. **LogicalExpression** - Same pattern as comparison
6. **UnaryExpression** - Different format, one operand
7. **TraversalExpression** - Different separator, no parens
8. **HasExpression** - Similar to traversal
9. **FunctionCall** - Most complex, needs loop

---

## Debugging Tips

### Tip 1: Run Individual Tests

Run one test at a time to focus:
```bash
go test ./pkg/filter/ast -v -run TestNumberLiteral
go test ./pkg/filter/ast -v -run TestBooleanLiteral
```

### Tip 2: Print Your Output

Add temporary print statements to see what you're generating:
```go
func (n *NumberLiteral) String() string {
    result := fmt.Sprintf("%g", n.Value)
    fmt.Println("NumberLiteral.String():", result)  // Debug
    return result
}
```

### Tip 3: Check the Test Expected Values

Look at the test file to see exactly what format is expected:
```go
// In ast_test.go
expected := "(age > 18)"
if expr.String() != expected {
    t.Errorf("String() = %q, want %q", expr.String(), expected)
}
```

### Tip 4: Parentheses Are Important

Notice which expressions use parentheses and which don't:
- **WITH parens:** ComparisonExpression, LogicalExpression, UnaryExpression
- **WITHOUT parens:** TraversalExpression, HasExpression, FunctionCall

### Tip 5: Recursive Thinking

Remember that String() methods often call String() on child nodes. This is recursion! The base cases are the leaf nodes (literals and identifiers).

---

## Common Mistakes

### Mistake 1: Forgetting to Call String()

```go
// ❌ WRONG - missing .String()
fmt.Sprintf("(%s > %s)", c.Left, c.Right)

// ✅ CORRECT
fmt.Sprintf("(%s > %s)", c.Left.String(), c.Right.String())
```

### Mistake 2: Wrong Format Specifier for Numbers

```go
// ❌ WRONG - might have trailing zeros
fmt.Sprintf("%f", 42.0)  // → "42.000000"

// ✅ CORRECT - smart formatting
fmt.Sprintf("%g", 42.0)  // → "42"
```

### Mistake 3: Forgetting Parentheses

```go
// ❌ WRONG - hard to see precedence
return fmt.Sprintf("%s AND %s", l.Left.String(), l.Right.String())
// → "age > 18 AND status = active"  (ambiguous!)

// ✅ CORRECT - clear precedence
return fmt.Sprintf("(%s AND %s)", l.Left.String(), l.Right.String())
// → "(age > 18 AND status = active)"  (clear!)
```

### Mistake 4: Extra Space in Function Calls

```go
// ❌ WRONG - space after comma
strings.Join(args, " , ")  // → "start , end"

// ✅ CORRECT - comma-space
strings.Join(args, ", ")   // → "start, end"
```

---

## Testing Your Implementation

Once you've implemented all the methods:

```bash
# Run all AST tests
go test ./pkg/filter/ast -v

# Run tests with coverage
go test ./pkg/filter/ast -v -cover

# Run a specific test
go test ./pkg/filter/ast -v -run TestComplexExpression
```

All tests should pass before moving to Module 3!

---

**Still stuck?** Check [SOLUTION.md](SOLUTION.md) for complete implementations with detailed explanations.

**Need more context?** Review [README.md](README.md) for conceptual understanding.
