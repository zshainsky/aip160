# Module 5: Hints for Advanced Features

This file provides strategic hints without giving complete solutions. If you're stuck, read the hints for your current task. If you're still stuck after trying, check SOLUTION.md.

## General Strategy Hints

### Pattern Recognition

**Lookahead Pattern** (used in all three features):
```
Step 1: Parse initial part (e.g., identifier)
Step 2: Check next token (peek)
Step 3: If special token, continue parsing that feature
Step 4: Otherwise, return what you have
```

**Chain Building Pattern** (used in traversal):
```
Step 1: Start with left expression
Step 2: While operator token present:
        - Save operator
        - Advance
        - Parse right expression
        - Build new node with left and right
        - Update left to be the new node
Step 3: Return final left
```

### Where to Modify Code

- **`parseIdentifier()`** - Starts simple, becomes a dispatcher
- **Between parseComparison and parseValue** - Add new precedence levels
- **New functions** - Create helpers for each feature

---

## Task 1: Field Traversal Hints

### Hint 1: Where to Start

The identifier is already parsed by `parseIdentifier()`. After creating the identifier node,  check if there's more to parse.

**Question:** What token indicates there's a field access coming?  
**Answer:** The DOT token (`.`)

### Hint 2: Checking for Traversal

Use the peek token to look ahead without consuming:
- Call the helper function that checks peek token type
- If it's DOT, you need to build a traversal chain
- If not, just return the identifier

### Hint 3: Building the Chain

Think about `a.b.c`:
- Start with `a` (an Identifier)
- See DOT, consume it
- Parse `b` (another Identifier)  
- Create TraversalExpression: `a.b`
- See DOT again, consume it
- Parse `c` (another Identifier)
- Create TraversalExpression: `(a.b).c`

**Pattern:** Each iteration builds a new TraversalExpression using the previous result as the left side.

### Hint 4: The Loop Structure

```
Start: left = Identifier
While NEXT token is DOT:
    Advance past DOT
    Parse right identifier
    Build TraversalExpression(left, right)
    Update left to be the new traversal
Return left
```

### Hint 5: What to Parse on the Right

After the DOT, what can appear?
- Always an identifier (field name)
- NOT another full expression
- Call the function that parses just an identifier

### Hint 6: AST Node Fields

TraversalExpression needs:
- `Token`: The DOT token (save it before advancing!)
- `Left`: The expression before the dot
- `Right`: The identifier after the dot (usually an Identifier node)

### Example Thought Process

For input `user.name`:
1. parseIdentifier creates Identifier("user")
2. Check peek: it's DOT
3. Enter loop
4. Advance past DOT (now current is "name")
5. Parse "name" as Identifier
6. Create TraversalExpression{Left: Identifier("user"), Right: Identifier("name")}
7. Check peek: it's not DOT (maybe EOF or operator)
8. Exit loop, return TraversalExpression

### Common Mistakes to Avoid

❌ **Parsing full expression on right side** - Only parse identifier  
❌ **Not saving DOT token** - Node needs it  
❌ **Not looping** - Won't handle `a.b.c`  
❌ **Breaking existing code** - Remember to return simple identifier when no DOT

---

## Task 2: Has Operator Hints

### Hint 1: Precedence Level

Has operator sits between comparison and traversal:
- **Lower than:** traversal (so `tags:user.role` means `tags:(user.role)`)
- **Higher than:** comparisons (so `tags:urgent = true` means `(tags:urgent) = true`)

**Where to add it:** Create new function called from parseComparison, before it calls parseTraversal.

### Hint 2: Function Structure

Pattern is similar to parseComparison:
```
Parse left side (from next precedence level)
Check current token for HAS operator
If HAS:
    Build HasExpression
Else:
    Return left as-is
```

### Hint 3: What Calls What

Update the call chain:
- `parseComparison` should call `parseHasExpression` (instead of parseValue)
- `parseHasExpression` should call `parseTraversal`
- `parseTraversal` should call `parseValue` or return identifier with checks

Actually, simpler approach:
- Keep `parseComparison` calling `parseValue`
- Inside `parseValue` or after parsing identifier + traversal, check for HAS
- Or create `parseHasExpression` between comparison and value

### Hint 4: Has is Binary

HasExpression is like ComparisonExpression:
- Has a left side (collection)
- Has an operator (`:`)
- Has a right side (member)

Same pattern as parseComparison:
1. Get left
2. Check for operator
3. If found, get right and build node
4. Return result

### Hint 5: Integrating into Parser

**Option A:** Modify parseComparison to check for HAS after getting left  
**Option B:** Create parseHasExpression that parseComparison calls  
**Option C:** Handle HAS in the same place as traversal

Recommended: Option B for clean separation of concerns

### Hint 6: Both Sides Can Be Complex

Left side: Could be traversal (`user.tags:admin`)  
Right side: Could be traversal or string (`tags:user.role`)

Both sides should call the same parsing function that handles traversal.

### Example Thought Process

For input `tags:urgent`:
1. Parse left: Identifier("tags") (potentially with traversal)
2. Current token is HAS (`:`)
3. Save operator, advance
4. Parse right: Identifier("urgent")
5. Create HasExpression{Collection: Identifier("tags"), Member: Identifier("urgent")}
6. Return HasExpression

---

## Task 3: Function Call Hints

### Hint 1: Distinguishing Functions from Identifiers

When you see an identifier, how do you know if it's a function?

**Answer:** Look at the next token!
- If next is `(`, it's a function call
- Otherwise, it's just an identifier/field name

### Hint 2: Where to Check

In `parseIdentifier()`, after creating the identifier node:
1. Check if peek token is LPAREN
2. If yes, this is actually a function call
3. Parse the function call and return FunctionCall node
4. If no, check for DOT (traversal), then return identifier

### Hint 3: Function Call Structure

What you need to parse:
1. Function name (already have it - the identifier)
2. Opening paren `(`
3. Arguments (none, one, or many)
4. Closing paren `)`

### Hint 4: Parsing Arguments

**Empty arguments:** `func()`
- After consuming `(`, check if current is `)`
- If yes, no arguments

**One argument:** `func(x)`
- After consuming `(`, parse one expression
- Check current is `)`

**Multiple arguments:** `func(x, y, z)`
- After consuming `(`, parse first expression
- While current is COMMA:
  - Consume comma
  - Parse next expression
- Expect `)`

###Hint 5: Pseudocode for Arguments

```
arguments = empty list
Advance past (
If current is NOT ):
    Parse expression, add to arguments
    While current is COMMA:
        Advance past comma
        Parse expression, add to arguments
Expect ) token
Return FunctionCall with arguments
```

### Hint 6: What to Parse for Each Argument

Each argument is a full expression! Call `parseExpression()` not `parseIdentifier()`.

Why? Because arguments can be:
- Literals: `func(42, "hello")`
- Fields: `func(user.name, status)`
- Expressions: `func(age > 18, status = "active")`
- Other functions: `func(get_user(), get_role())`

### Hint 7: AST Node Fields

FunctionCall needs:
- `Token`: The identifier token (function name)
- `Function`: The function name as string
- `Arguments`: Slice of Expression (`[]Expression`)

### Example Thought Process

For input `timestamp(created_at)`:
1. parseIdentifier creates Identifier("timestamp")
2. Check peek: it's LPAREN
3. This is a function call!
4. currentToken already has "timestamp"
5. Advance past `(`
6. Current token is "created_at", not `)`
7. Parse expression: Identifier("created_at")
8. Add to arguments list
9. Current token is `)` (from parsing expression)
10. Expect `)` - success
11. Return FunctionCall{Function: "timestamp", Arguments: [Identifier("created_at")]}

For input `has(tags, "urgent")`:
1. Parse identifier "has", see LPAREN
2. Advance past `(`
3. Parse expression: Identifier("tags")
4. Current is COMMA
5. Advance past comma
6. Parse expression: StringLiteral("urgent")
7. Current is `)`
8. Return FunctionCall{Function: "has", Arguments: [Identifier("tags"), StringLiteral("urgent")]}

### Common Mistakes to Avoid

❌ **Parsing identifier for arguments** - Use parseExpression  
❌ **Not handling empty arguments** - Check for `)` immediately after `(`  
❌ **Not consuming comma** - Need to advance past it in loop  
❌ **Wrong token for node** - Use initial identifier token, not current

---

## Task 4: Integration Hints

### Hint 1: Order of Operations

When modifying parseIdentifier, check tokens in this order:
1. First, check for LPAREN (function call)
2. Then, check for DOT (traversal)
3. Return the result

Why this order? Functions bind tightest, then field access.

### Hint 2: Function Names Can Be Traversed?

Should `user.get_name()` be valid?

In this grammar: **No**. Functions are standalone identifiers.  
Traversal returns a different expressions type (TraversalExpression), not an identifier.

### Hint 3: Nested Features

These should work:
- `user.tags:admin` - Traversal, then has
- `user.permissions:write AND status = "active"` - Has in logical expression
- `timestamp(user.created_at)` - Function with traversal as argument
- `tags:get_priority()` - Has with function call as member

### Hint 4: Testing Strategy

Test each feature individually first:
1. Simple traversal: `a.b`
2. Simple has: `a:b`
3. Simple function: `f(x)`

Then test combinations:
1. Nested traversal: `a.b.c.d`
2. Traversal + has: `a.b:c`
3. Function + traversal: `f(a.b)`
4. Everything: `f(a.b.c):value AND status = "active"`

### Hint 5: Preserving Module 4

Make sure existing tests still pass!
- Simple comparisons: `age > 18`
- Logical operators: `a AND b OR c`
- Grouped expressions: `(a OR b) AND c`

If these break, you changed something fundamental in the parse chain.

---

##Debugging Tips

### Print Current State

Add temporary debug prints:
```go
fmt.Fprintf(os.Stderr, "DEBUG parseIdentifier: current=%v, peek=%v\n", 
    p.currentToken, p.peekToken)
```

Run with `-v` flag to see output.

### Check Token Position

After parsing, where is currentToken?
- Should be on the first token AFTER what you parsed
- Example: After parsing `user.name`, current should be the operator/EOF, not `name`

### Draw AST Trees

For complex expressions, draw the tree:
```
user.tags:admin
       :
      / \
     .   admin
    / \
  user tags
```

Make sure your code builds this structure.

### Small Steps

Don't implement all three features at once!
1. Get traversal working alone
2. Add has operator
3. Add function calls
4. Test combinations

---

## Still Stuck?

1. Re-read the README section for your task
2. Look at similar patterns in existing code (parseAndExpression, parseComparison)
3. Check SOLUTION.md for the complete implementation
4. Remember: Tests are your friend - they tell you exactly what should work

Good luck! 🚀
