# Module 2: Abstract Syntax Tree (AST) Design

**Duration**: ~1.5 hours  
**Difficulty**: Intermediate

## 🎯 Module Objectives

By the end of this module, you will:
1. Understand what an Abstract Syntax Tree (AST) is and why it's essential
2. Design Go types to represent filter expressions as a tree structure
3. Learn the Node interface pattern common in AST implementations
4. Implement String() methods for debugging and testing
5. Understand how the AST bridges the lexer and parser

## 📖 What is an Abstract Syntax Tree?

An **Abstract Syntax Tree (AST)** is a tree representation of the structure of source code. Each node in the tree represents a construct in the code.

### The "Abstract" Part

It's called "abstract" because it represents the **logical structure** without syntactic details like whitespace, parentheses positions, or exact formatting.

**Example:**
```
Input: age > 18 AND status = "active"

Tokens (from Lexer):
IDENTIFIER("age"), GREATER_THAN, NUMBER("18"), AND, IDENTIFIER("status"), EQUALS, STRING("active")

AST (tree structure):
           LogicalExpression (AND)
                  /      \
     ComparisonExpr      ComparisonExpr
     (age > 18)          (status = "active")
        /    \              /         \
    Ident   Number      Ident      String
    "age"     18       "status"   "active"
```

## 🌳 Why Do We Need an AST?

The AST sits between the lexer and the evaluator:

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  String  │ ──→ │  Lexer   │ ──→ │  Parser  │ ──→ │   AST    │
│  Input   │     │ (Tokens) │     │          │     │  (Tree)  │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
                                                          │
                                                          ↓
                                                    ┌──────────┐
                                                    │Evaluator │
                                                    │ (Result) │
                                                    └──────────┘
```

### Benefits:
1. **Structure**: Tokens are linear, AST shows relationships and hierarchy
2. **Precedence**: Tree structure naturally encodes operator precedence
3. **Evaluation**: Easy to traverse and evaluate recursively
4. **Transformation**: Can optimize or modify the filter before evaluation
5. **Multiple Backends**: Same AST can be used for different evaluation strategies

## 🧩 AST Node Types for AIP-160

Let's break down the filter language into node types:

### 1. **Literals** (Leaf Nodes)
Values that represent themselves:
```go
StringLiteral    → "John", "active"
NumberLiteral    → 42, 3.14, 2.5e10
BooleanLiteral   → true, false
NullLiteral      → null
```

### 2. **Identifier** (Leaf Node)
Field names:
```go
Identifier → age, name, user, status
```

### 3. **Binary Expressions** (Branch Nodes)
Operations with two operands:
```go
ComparisonExpression → age > 18, name = "John"
LogicalExpression    → expr1 AND expr2, expr1 OR expr2
TraversalExpression  → user.email, address.city
HasExpression        → tags:urgent, roles:admin
```

### 4. **Unary Expressions** (Branch Nodes)
Operations with one operand:
```go
UnaryExpression → NOT active, -5
```

### 5. **Function Calls** (Branch Nodes)
Function invocations:
```go
FunctionCall → timestamp(created_at), duration(start, end)
```

### 6. **Program** (Root Node)
The root of the entire tree:
```go
Program → Contains the top-level expression
```

## 🎨 Visual Tree Examples

### Example 1: Simple Comparison
**Filter:** `age > 18`

```
    ComparisonExpr
        /    \
   Identifier Number
    "age"      18
```

### Example 2: Logical AND
**Filter:** `age > 18 AND status = "active"`

```
         LogicalExpr (AND)
            /         \
   ComparisonExpr   ComparisonExpr
    (age > 18)      (status = "active")
       /    \           /         \
   Ident   Number   Ident      String
   "age"     18    "status"   "active"
```

### Example 3: Nested with OR
**Filter:** `(age > 18 AND status = "active") OR role = "admin"`

```
                LogicalExpr (OR)
                 /              \
        LogicalExpr (AND)    ComparisonExpr
         /         \          (role = "admin")
ComparisonExpr  ComparisonExpr     /        \
 (age > 18)   (status = "active") Ident   String
   /    \         /         \    "role"  "admin"
Ident  Number  Ident     String
"age"    18   "status" "active"
```

### Example 4: Traversal
**Filter:** `user.email = "john@example.com"`

```
        ComparisonExpr
           /        \
  TraversalExpr   String
      /     \    "john@example.com"
  Ident   Ident
  "user" "email"
```

### Example 5: Has Operator
**Filter:** `tags:urgent`

```
    HasExpr
     /    \
  Ident  Ident
 "tags" "urgent"
```

## 💡 Key Concepts

### 1. **The Node Interface**

All AST nodes implement a common `Node` interface:

```go
type Node interface {
    TokenLiteral() string  // For debugging
    String() string        // String representation
}
```

This allows you to treat any node generically.

### 2. **Expression Interface & The Marker Method Pattern**

Expressions are nodes that can be evaluated to produce a value:

```go
type Expression interface {
    Node
    expressionNode()  // Marker method
}
```

#### Understanding the Marker Method

The `expressionNode()` is a **marker method** (also called a "tag method"). It's an empty method that is **never actually called** - but it serves an important purpose!

**Why does it exist?**

It's a **type safety mechanism** in Go. Without the marker method:

```go
// ❌ Without marker - TOO PERMISSIVE
type Expression interface {
    Node  // ANY type implementing Node is automatically an Expression
}
```

With this definition, you couldn't distinguish between different categories of nodes. Every Node would automatically be an Expression!

**With the marker method:**

```go
// ✅ With marker - EXPLICIT MEMBERSHIP
type Expression interface {
    Node
    expressionNode()  // Only types that explicitly add this are Expressions
}
```

Now a type must **explicitly** declare itself as an Expression by implementing the marker:

```go
func (i *Identifier) expressionNode() {}  // Empty - just a badge of membership!
```

**Real-world benefit - Type Safety:**

```go
// This function ONLY accepts Expressions
func Evaluate(expr Expression, data map[string]interface{}) interface{} {
    // Compiler enforces: expr MUST have expressionNode() method
    switch e := expr.(type) {
    case *NumberLiteral:
        return e.Value
    case *ComparisonExpression:
        // ... evaluate comparison
    }
}

// Try passing the wrong type:
var token lexer.Token
Evaluate(token, data)  // ❌ Compiler error! Token doesn't have expressionNode()
```

**In a full language:**

If you were building a complete programming language, you might have:

```go
type Expression interface {
    Node
    expressionNode()  // Things that produce values: 5 + 3, "hello", x > 10
}

type Statement interface {
    Node
    statementNode()  // Things that perform actions: if, return, for
}

// Now you can have type-safe functions:
func evaluateExpr(expr Expression) interface{}  // Only expressions allowed
func executeStmt(stmt Statement) error          // Only statements allowed
```

**Key Takeaway:** The marker method is never called - it's purely a **compile-time membership badge** that says "I'm an Expression" and enforces type safety. It's a standard pattern from well-known parsers and follows Go best practices.

### 3. **String Representation**

The `String()` method is crucial for:
- **Debugging**: See what your AST looks like
- **Testing**: Verify structure easily
- **Error messages**: Show users what was parsed

Convention: Use parentheses to show precedence clearly:
```go
age > 18 AND status = "active"
→ "((age > 18) AND (status = \"active\"))"
```

### 4. **Tree Traversal**

ASTs are naturally recursive. To process them, you recursively visit nodes:

```go
// Example: Evaluating a comparison
func evaluate(node Node) interface{} {
    switch n := node.(type) {
    case *NumberLiteral:
        return n.Value
    case *ComparisonExpression:
        left := evaluate(n.Left)
        right := evaluate(n.Right)
        return compare(left, n.Operator, right)
    // ... handle other node types
    }
}
```

### 5. **Operator Precedence in the Tree**

The tree structure naturally encodes operator precedence:

**Filter:** `a OR b AND c`  
**Meaning:** `a OR (b AND c)` (AND binds tighter)

```
    OR
   /  \
  a   AND
      /  \
     b    c
```

Notice how AND is deeper in the tree - it gets evaluated first!

## 📝 Implementation Task

Your task is to complete the AST node implementations in `pkg/filter/ast/ast.go`.

### What's Already Done:
- ✅ `Node` and `Expression` interfaces defined
- ✅ `Program` struct (root node)
- ✅ `Identifier` struct (fully implemented as example)
- ✅ `StringLiteral` struct (fully implemented as example)

### What You Need to Implement:

For each remaining node type, implement these three methods:

1. **expressionNode()** - Marker method (just empty function body)
2. **TokenLiteral()** - Return the token's literal value
3. **String()** - Return string representation

#### Node Types to Complete:

1. **NumberLiteral**
   - String() should format the number properly

2. **BooleanLiteral**
   - String() should return "true" or "false"

3. **NullLiteral**
   - String() should return "null"

4. **ComparisonExpression**
   - String() should return: `"(<left> <operator> <right>)"`
   - Example: `"(age > 18)"`

5. **LogicalExpression**
   - String() should return: `"(<left> <operator> <right>)"`
   - Example: `"((age > 18) AND (status = \"active\"))"`

6. **UnaryExpression**
   - String() should return: `"(<operator><right>)"`
   - Example: `"(NOT active)"` or `"(-5)"`

7. **TraversalExpression**
   - String() should return: `"<left>.<right>"`
   - Example: `"user.email"`

8. **HasExpression**
   - String() should return: `"<collection>:<member>"`
   - Example: `"tags:urgent"`

9. **FunctionCall**
   - String() should return: `"<function>(<arg1>, <arg2>, ...)"`
   - Example: `"duration(start, end)"`

## 🎯 Test-Driven Approach

Run the tests to guide your implementation:

```bash
cd /Users/zshainky/Projects/aip160
go test ./pkg/filter/ast -v
```

The tests cover:
1. ✅ Identifier (already passes - example implementation)
2. ✅ StringLiteral (already passes - example implementation)  
3. ❌ NumberLiteral (you'll implement)
4. ❌ BooleanLiteral (you'll implement)
5. ❌ NullLiteral (you'll implement)
6. ❌ ComparisonExpression (you'll implement)
7. ❌ LogicalExpression (you'll implement)
8. ❌ UnaryExpression (you'll implement)
9. ❌ TraversalExpression (you'll implement)
10. ❌ HasExpression (you'll implement)
11. ❌ FunctionCall (you'll implement)
12. ❌ Complex nested expressions (you'll implement)

## 🚀 Getting Started

### Step 1: Review the Starter Code

Open `pkg/filter/ast/ast.go` and read through:
- The `Node` and `Expression` interfaces
- The `Identifier` and `StringLiteral` implementations (these are complete)
- The TODO comments for other node types

### Step 2: Understand the Pattern

Each node follows this pattern:

```go
type SomeNode struct {
    Token lexer.Token  // The token this node is based on
    // ... other fields specific to this node type
}

func (s *SomeNode) expressionNode() {}  // Marker method

func (s *SomeNode) TokenLiteral() string {
    return s.Token.Literal
}

func (s *SomeNode) String() string {
    // Return string representation
    // This is where most of your work is!
}
```

### Step 3: Run the Tests

```bash
go test ./pkg/filter/ast -v
```

You'll see which tests are failing. Start with the simpler ones!

### Step 4: Implement Node by Node

Start with the easiest:
1. **NullLiteral** - Just return "null"
2. **BooleanLiteral** - Return "true" or "false" based on Value
3. **NumberLiteral** - Format the number for display
4. Then move to expressions...

### Tip: Use fmt.Sprintf()

For complex String() methods, use `fmt.Sprintf()`:

```go
func (c *ComparisonExpression) String() string {
    return fmt.Sprintf("(%s %s %s)", 
        c.Left.String(), 
        c.Operator, 
        c.Right.String())
}
```

### Tip: Handle Function Arguments

For FunctionCall, you'll need to join arguments:

```go
// Build comma-separated list
var args []string
for _, arg := range f.Arguments {
    args = append(args, arg.String())
}
return fmt.Sprintf("%s(%s)", f.Function, strings.Join(args, ", "))
```

## 💭 Reflection Questions

As you implement, think about:

1. **Why use an interface instead of a concrete type?**  
   *What flexibility does `Expression` interface provide?*

2. **How does the tree structure help with evaluation?**  
   *How would you evaluate a LogicalExpression?*

3. **What's the relationship between tokens and AST nodes?**  
   *Does every token become a node?*

4. **How would you add a new operator to AIP-160?**  
   *What would you need to change?*

5. **Why do we need both left and right children?**  
   *Could we use a different structure?*

## 🎓 Learning Checkpoint

Before moving to the next module, make sure you can:

- [ ] Explain what an AST is and why it's used
- [ ] Identify the node type for any part of a filter expression
- [ ] Draw the tree structure for a given filter
- [ ] Explain the Node and Expression interfaces
- [ ] Understand how tree structure encodes precedence
- [ ] All tests pass! ✅

## 💡 Common Patterns

### Pattern 1: Recursive String Building

Many String() methods recursively call String() on child nodes:

```go
func (l *LogicalExpression) String() string {
    return fmt.Sprintf("(%s %s %s)",
        l.Left.String(),   // Recursive call!
        l.Operator,
        l.Right.String())  // Recursive call!
}
```

### Pattern 2: Type Assertions

When traversing, you'll use type assertions:

```go
switch node := expr.(type) {
case *NumberLiteral:
    // Handle number
case *ComparisonExpression:
    // Handle comparison
}
```

### Pattern 3: Visitor Pattern (Advanced)

For complex operations, you might use the Visitor pattern:

```go
type Visitor interface {
    VisitComparison(*ComparisonExpression)
    VisitLogical(*LogicalExpression)
    // ...
}

func (c *ComparisonExpression) Accept(v Visitor) {
    v.VisitComparison(c)
}
```

We won't implement this now, but it's good to know!

## 📚 Resources

### Implementation Help

**Need hints?** → [HINTS.md](HINTS.md) - Progressive hints for implementing each node type

**Really stuck?** → [SOLUTION.md](SOLUTION.md) - Complete implementations with explanations

### Additional Learning

- [Abstract Syntax Trees](https://en.wikipedia.org/wiki/Abstract_syntax_tree)
- [Visitor Pattern](https://en.wikipedia.org/wiki/Visitor_pattern)
- [Tree Data Structures](https://en.wikipedia.org/wiki/Tree_(data_structure))

---

**Ready to code?** Open `pkg/filter/ast/ast.go` and start implementing!

**Run tests:** `go test ./pkg/filter/ast -v`

**Previous Module**: [Module 1: Lexer](../module1-lexer/README.md)  
**Next Module**: [Module 3: EBNF & Parser Design](../module3-parser-design/README.md) (coming soon)
