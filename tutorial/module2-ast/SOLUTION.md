# Module 2 AST Complete Solution

This document contains the complete implementation of all AST node types with detailed explanations.

## Complete Implementation

Here's the full `ast.go` file with all TODO items completed:

```go
package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

// Node is the base interface that all AST nodes must implement.
type Node interface {
	TokenLiteral() string
	String() string
}

// Expression represents any node that can be evaluated to produce a value.
type Expression interface {
	Node
	expressionNode() // marker method
}

// Program is the root node of the AST.
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
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// StringLiteral represents a string value in quotes.
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (s *StringLiteral) expressionNode()      {}
func (s *StringLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *StringLiteral) String() string       { return "\"" + s.Value + "\"" }

// NumberLiteral represents a numeric value (integer or float).
type NumberLiteral struct {
	Token lexer.Token
	Value float64
}

func (n *NumberLiteral) expressionNode()      {}
func (n *NumberLiteral) TokenLiteral() string { return n.Token.Literal }
func (n *NumberLiteral) String() string       { return fmt.Sprintf("%g", n.Value) }

// BooleanLiteral represents a boolean value.
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (b *BooleanLiteral) expressionNode()      {}
func (b *BooleanLiteral) TokenLiteral() string { return b.Token.Literal }
func (b *BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

// NullLiteral represents a null value.
type NullLiteral struct {
	Token lexer.Token
}

func (n *NullLiteral) expressionNode()      {}
func (n *NullLiteral) TokenLiteral() string { return n.Token.Literal }
func (n *NullLiteral) String() string       { return "null" }

// ComparisonExpression represents a comparison between two expressions.
type ComparisonExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (c *ComparisonExpression) expressionNode()      {}
func (c *ComparisonExpression) TokenLiteral() string { return c.Token.Literal }
func (c *ComparisonExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", c.Left.String(), c.Operator, c.Right.String())
}

// LogicalExpression represents a logical operation between two expressions.
type LogicalExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (l *LogicalExpression) expressionNode()      {}
func (l *LogicalExpression) TokenLiteral() string { return l.Token.Literal }
func (l *LogicalExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", l.Left.String(), l.Operator, l.Right.String())
}

// UnaryExpression represents a unary operation.
type UnaryExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (u *UnaryExpression) expressionNode()      {}
func (u *UnaryExpression) TokenLiteral() string { return u.Token.Literal }
func (u *UnaryExpression) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator, u.Right.String())
}

// TraversalExpression represents field traversal with the dot operator.
type TraversalExpression struct {
	Token lexer.Token
	Left  Expression
	Right Expression
}

func (t *TraversalExpression) expressionNode()      {}
func (t *TraversalExpression) TokenLiteral() string { return t.Token.Literal }
func (t *TraversalExpression) String() string {
	return fmt.Sprintf("%s.%s", t.Left.String(), t.Right.String())
}

// HasExpression represents the has operator for checking collection membership.
type HasExpression struct {
	Token      lexer.Token
	Collection Expression
	Member     Expression
}

func (h *HasExpression) expressionNode()      {}
func (h *HasExpression) TokenLiteral() string { return h.Token.Literal }
func (h *HasExpression) String() string {
	return fmt.Sprintf("%s:%s", h.Collection.String(), h.Member.String())
}

// FunctionCall represents a function call expression.
type FunctionCall struct {
	Token     lexer.Token
	Function  string
	Arguments []Expression
}

func (f *FunctionCall) expressionNode()      {}
func (f *FunctionCall) TokenLiteral() string { return f.Token.Literal }
func (f *FunctionCall) String() string {
	var args []string
	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Function, strings.Join(args, ", "))
}
```

---

## Detailed Explanations

### NumberLiteral

```go
func (n *NumberLiteral) String() string {
	return fmt.Sprintf("%g", n.Value)
}
```

**Why `%g`?**
- The `%g` format specifier intelligently chooses between fixed and exponential format
- It removes trailing zeros: `42.0` → `"42"`
- It uses exponential notation for very large/small numbers: `2.5e+10`
- Perfect for human-readable number representation

**Alternatives:**
- `%f` - always fixed point, includes unnecessary decimals: `42.000000`
- `%e` - always exponential: `4.200000e+01` (too verbose)
- `%v` - works but less control

### BooleanLiteral

```go
func (b *BooleanLiteral) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}
```

**Design Decision:**
- Simple conditional is clearest
- Could use `fmt.Sprintf("%t", b.Value)` but explicit is better
- Matches Go's boolean literal syntax exactly

**Why not just return the token literal?**
- The token might be "TRUE" (uppercase) but we want normalized output
- We're converting based on the parsed boolean value, not the original text

### NullLiteral

```go
func (n *NullLiteral) String() string {
	return "null"
}
```

**Simplest implementation:**
- No state to check
- Always returns the same string
- Could return `n.Token.Literal` but explicit is clearer

### ComparisonExpression

```go
func (c *ComparisonExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", c.Left.String(), c.Operator, c.Right.String())
}
```

**Key Points:**
1. **Parentheses** - Make precedence explicit: `(age > 18)`
2. **Recursive calls** - `c.Left.String()` and `c.Right.String()` recursively build the tree
3. **Spaces around operator** - Readable format: `age > 18` not `age>18`

**Example Tree Walk:**
```
ComparisonExpression(age > 18)
  ├─ Left: Identifier("age").String() → "age"
  ├─ Operator: ">"
  └─ Right: NumberLiteral(18).String() → "18"
Result: "(age > 18)"
```

### LogicalExpression

```go
func (l *LogicalExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", l.Left.String(), l.Operator, l.Right.String())
}
```

**Same pattern as ComparisonExpression:**
- The structure is identical
- Only the semantic meaning differs (logical vs comparison)
- Shows good design: similar concepts use similar implementations

**Nested Example:**
```
LogicalExpression(AND)
  ├─ Left: ComparisonExpression(age > 18).String() → "(age > 18)"
  ├─ Operator: "AND"
  └─ Right: ComparisonExpression(status = "active").String() → "(status = \"active\")"
Result: "((age > 18) AND (status = \"active\"))"
```

Notice the nested parentheses - each level adds its own!

### UnaryExpression

```go
func (u *UnaryExpression) String() string {
	return fmt.Sprintf("(%s%s)", u.Operator, u.Right.String())
}
```

**Differences from binary expressions:**
1. **No left operand** - Only Right is used
2. **No space** - Operator directly precedes operand: `NOT active`, `-5`
3. **Still uses parentheses** - Maintains consistency

**Examples:**
- `NOT active` → `"(NOT active)"`
- `-5` → `"(-5)"`

### TraversalExpression

```go
func (t *TraversalExpression) String() string {
	return fmt.Sprintf("%s.%s", t.Left.String(), t.Right.String())
}
```

**Key Difference:**
- **NO parentheses** - Dot notation is already unambiguous
- Produces natural syntax: `user.email` not `(user.email)`

**Handles Nesting Naturally:**
```
TraversalExpression(.)
  ├─ Left: TraversalExpression(.).String() → "user.address"
  │   ├─ Left: Identifier("user").String() → "user"
  │   └─ Right: Identifier("address").String() → "address"
  └─ Right: Identifier("city").String() → "city"
Result: "user.address.city"
```

The recursive structure naturally creates chained traversals!

### HasExpression

```go
func (h *HasExpression) String() string {
	return fmt.Sprintf("%s:%s", h.Collection.String(), h.Member.String())
}
```

**Similar to TraversalExpression:**
- Different separator (`:` instead of `.`)
- No parentheses needed
- Clean, readable syntax

**Examples:**
- `tags:urgent` → `"tags:urgent"`
- `roles:admin` → `"roles:admin"`

### FunctionCall

```go
func (f *FunctionCall) String() string {
	var args []string
	for _, arg := range f.Arguments {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("%s(%s)", f.Function, strings.Join(args, ", "))
}
```

**Most Complex Implementation:**

**Step 1: Build argument strings**
```go
var args []string
for _, arg := range f.Arguments {
    args = append(args, arg.String())
}
```
- Iterate through all argument expressions
- Call String() on each (recursion!)
- Collect into a slice of strings

**Step 2: Join with commas**
```go
strings.Join(args, ", ")
```
- Creates comma-separated list: `"start, end"`
- Empty slice becomes empty string: `""`

**Step 3: Build final string**
```go
fmt.Sprintf("%s(%s)", f.Function, joinedArgs)
```
- Function name with parentheses
- Arguments inside parentheses

**Examples:**
- `now()` - Empty args: `"now()"`
- `timestamp(created_at)` - One arg: `"timestamp(created_at)"`
- `duration(start, end)` - Two args: `"duration(start, end)"`

---

## Design Patterns Used

### 1. **Interface Segregation**

We have two interfaces:
- `Node` - Common to all AST nodes
- `Expression` - Specific to expression nodes

This follows the **Interface Segregation Principle**: Clients shouldn't depend on interfaces they don't use. If we later add Statement nodes (for a full programming language), they'd implement Node but not Expression.

### 2. **Marker Method Pattern**

```go
type Expression interface {
    Node
    expressionNode()  // Marker method
}
```

The `expressionNode()` method is never called and has no body. It exists solely for **type safety** - only types that explicitly implement it can satisfy the Expression interface.

**Why not just use Node?**
- Makes intent explicit in function signatures
- `func Parse() Expression` vs `func Parse() Node`
- The first makes it clear we're parsing expressions, not potentially other node types

### 3. **Composite Pattern**

The AST is a **Composite** structure:
- **Leaf nodes**: Literals, Identifiers (no children)
- **Composite nodes**: Expressions (have child expressions)

All nodes implement the same interface, allowing uniform treatment.

### 4. **Recursive Composition**

Expression nodes contain other Expressions:
```go
type LogicalExpression struct {
    Left  Expression  // Can be any expression!
    Right Expression  // Can be any expression!
}
```

This allows arbitrary nesting and is the essence of tree structures.

---

## Common Pitfalls & Solutions

### Pitfall 1: Infinite Recursion

**Problem:** Circular references in the AST
```go
// Don't do this!
expr := &LogicalExpression{
    Left: expr,  // Refers to itself!
}
```

**Solution:** The parser (Module 3) will construct the tree correctly. In this module, we just define the types.

### Pitfall 2: Nil Pointer Dereference

**Problem:** Calling methods on nil expressions
```go
var expr Expression  // nil
expr.String()  // Panic!
```

**Solution:** The parser will ensure expressions are properly initialized. Add nil checks in production code:
```go
func (l *LogicalExpression) String() string {
    if l.Left == nil || l.Right == nil {
        return "<invalid>"
    }
    return fmt.Sprintf("(%s %s %s)", l.Left.String(), l.Operator, l.Right.String())
}
```

### Pitfall 3: Forgetting Imports

**Problem:** Using `fmt` or `strings` without importing

**Solution:** Ensure imports at top of file:
```go
import (
    "fmt"
    "strings"
    "github.com/zshainsky/aip160/pkg/filter/lexer"
)
```

### Pitfall 4: Inconsistent Formatting

**Problem:** Some expressions with parens, some without
```go
// Inconsistent
ComparisonExpression: "(age > 18)"
LogicalExpression: "age > 18 AND status = active"  // Missing parens!
```

**Solution:** Follow the conventions:
- Binary expressions with operators: **use parentheses**
- Dot and colon notation: **no parentheses**
- Function calls: **natural function syntax**

---

## Testing Your Implementation

### Run All Tests

```bash
cd /Users/zshainky/Projects/aip160
go test ./pkg/filter/ast -v
```

Expected output (all passing):
```
=== RUN   TestIdentifier
--- PASS: TestIdentifier (0.00s)
=== RUN   TestStringLiteral
--- PASS: TestStringLiteral (0.00s)
=== RUN   TestNumberLiteral
=== RUN   TestNumberLiteral/integer
--- PASS: TestNumberLiteral/integer (0.00s)
=== RUN   TestNumberLiteral/float
--- PASS: TestNumberLiteral/float (0.00s)
=== RUN   TestNumberLiteral/scientific_notation
--- PASS: TestNumberLiteral/scientific_notation (0.00s)
--- PASS: TestNumberLiteral (0.00s)
... (all tests pass)
PASS
ok      github.com/zshainsky/aip160/pkg/filter/ast      0.123s
```

### Test Individual Node Types

```bash
go test ./pkg/filter/ast -v -run TestNumberLiteral
go test ./pkg/filter/ast -v -run TestLogicalExpression
go test ./pkg/filter/ast -v -run TestComplexExpression
```

### Check Test Coverage

```bash
go test ./pkg/filter/ast -cover
```

Should show high coverage (close to 100% for String() methods).

---

## What's Next?

With the AST complete, you now have:
1. ✅ **Lexer** - Converts strings to tokens
2. ✅ **AST** - Defines the tree structure for parsed expressions

Next up:
3. ⏭️ **Parser** - Converts tokens to AST (Modules 3 & 4)
4. ⏭️ **Evaluator** - Executes the AST against data (Module 6)

---

## Reflection Questions - Answers

Let's revisit those reflection questions:

**1. Why use an interface instead of a concrete type?**

Interfaces provide **polymorphism** - we can write functions that work with any Expression without knowing the specific type:

```go
func Evaluate(expr Expression, data map[string]interface{}) interface{} {
    switch e := expr.(type) {
    case *NumberLiteral:
        return e.Value
    case *ComparisonExpression:
        left := Evaluate(e.Left, data)
        right := Evaluate(e.Right, data)
        return compare(left, e.Operator, right)
    // ... handle all types
    }
}
```

**2. How does the tree structure help with evaluation?**

Trees naturally map to **recursive evaluation**:
- Evaluate leaf nodes directly (literals, identifiers)
- Evaluate branch nodes by evaluating children first
- Bottom-up evaluation: leaves → branches → root

**3. What's the relationship between tokens and AST nodes?**

- **Tokens** are flat, sequential: `[age, >, 18, AND, status, =, "active"]`
- **AST** is hierarchical, showing relationships:
  ```
  AND
   ├─ age > 18
   └─ status = "active"
  ```
- Each AST node stores its associated token for error reporting
- Not all tokens become nodes (e.g., whitespace tokens are skipped)

**4. How would you add a new operator to AIP-160?**

1. Add token type in `lexer/token.go`
2. Update lexer to recognize it
3. Add AST node type in `ast/ast.go`
4. Update parser to handle it (Module 3)
5. Update evaluator to execute it (Module 6)

The AST is just one step in the pipeline!

**5. Why do we need both left and right children?**

Most operators are **binary** (two operands). The tree structure makes the relationship explicit:
- **Left child**: First operand
- **Right child**: Second operand
- **Operator**: How to combine them

For **unary** operators (NOT, -), we only use one child.

---

**Congratulations!** You've completed the AST design module. You now understand how to represent parsed code as a tree structure!

**Previous Module**: [Module 1: Lexer](../module1-lexer/README.md)  
**Next Module**: [Module 3: EBNF & Parser Design](../module3-parser-design/README.md) (coming soon)
