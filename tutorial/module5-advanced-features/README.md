# Module 5: Advanced Parser Features

**Duration:** ~2 hours  
**Difficulty:** Intermediate to Advanced

## � Quick Start

**Before you begin:** This module builds on Module 4. Your `parser.go` file contains TODO comments marking where to add code.

**Recommended workflow:**

1. **Start with Task 1 (Field Traversal)** - Modify `parseIdentifier()` to handle the `.` operator
   ```bash
   go test ./pkg/filter/parser -run TestFieldTraversal -v
   ```
   
2. **Then Task 2 (Has Operator)** - Create `parseHasExpression()` and update precedence chain
   ```bash
   go test ./pkg/filter/parser -run TestHasOperator -v
   ```
   
3. **Then Task 3 (Function Calls)** - Add function detection to `parseIdentifier()` and create `parseFunctionCall()`
   ```bash
   go test ./pkg/filter/parser -run TestFunctionCall -v
   ```
   
4. **Finally, verify everything works together:**
   ```bash
   go test ./pkg/filter/parser -v
   ```

**Where to write code:** Search for "TODO (Module 5" in `pkg/filter/parser/parser.go` - there are 5 TODO comments guiding you.

---

## �📚 Overview

In this module, you'll extend the parser to handle three advanced AIP-160 features:
1. **Field Traversal** - Navigate through nested structures (`user.name`, `address.city.zipcode`)
2. **Has Operator** - Check collection membership (`tags:urgent`, `roles:admin`)
3. **Function Calls** - Invoke custom functions (`timestamp(created_at)`, `duration(start, end)`)

These features make the filter language much more powerful for real-world use cases.

## 🎯 Learning Objectives

By completing this module, you will:

- Understand how to parse binary operators with different semantics (`.` for traversal, `:` for has)
- Learn recursive parsing patterns for nested structures
- Implement function call parsing with argument lists
- Handle operator precedence in complex expressions
- Extend an existing parser without breaking existing functionality

## 📖 Background

### Field Traversal (`.` operator)

The dot operator allows navigating through nested message fields, maps, and structs.

**Examples:**
```
user.name = "John"           # Access name field of user
address.city.zipcode > 10000 # Navigate through multiple levels
author.email:*               # Combine traversal with has operator
```

**Grammar:**
```ebnf
traversal = identifier, {".", identifier} ;
```

**Key Insight:** A traversal is just a chain of identifiers connected by dots. The left side is always evaluated first, then each subsequent field is accessed. This is **left-associative**.

**Tree Structure for `user.name.first`:**
```
      .
     / \
    .   first
   / \
user  name
```

### Has Operator (`:`)

The has operator checks if a value exists in a collection or map.

**Examples:**
```
tags:urgent                  # Check if 'urgent' is in tags array
roles:"admin"                # Check if roles contains "admin"  
labels:*                     # Check if labels map has any entries
user.permissions:write       # Combine with traversal
```

**Grammar:**
```ebnf
has_expression = expression, ":", expression ;
```

**Key Insight:** The has operator has **higher precedence** than comparisons but **lower than** traversal. So `tags:urgent OR roles:admin` parses as `(tags:urgent) OR (roles:admin)`, not `tags:(urgent OR roles:admin)`.

**Precedence Order (highest to lowest):**
```
1. Function calls, literals, identifiers, grouping
2. Field traversal (.)
3. Has operator (:)
4. Comparison operators (=, !=, <, >, <=, >=)
5. NOT
6. AND
7. OR
```

### Function Calls

Functions allow custom operations like type conversion, timestamp parsing, or custom validation logic.

**Examples:**
```
timestamp(created_at) > timestamp("2024-01-01")
has(tags, "urgent")
duration(start, end) < 3600
```

**Grammar:**
```ebnf
function_call = identifier, "(", [expression, {",", expression}], ")" ;
```

**Key Insight:** Function calls look like identifiers followed by `(`. You need to **look ahead** to determine if an identifier is a function call or just a field name.

## 🔧 Implementation Strategy

### Where to Add Code

You'll be modifying the existing parser code, not creating new files. The key functions to update:

1. **`parseValue()`** - Entry point for parsing values
   - After parsing identifier, check for `.`, `:`, or `(`
   - Dispatch to appropriate handling logic

2. **`parseIdentifier()`** - Currently handles simple identifiers
   - Extend to check peek token for special operators
   - Build traversal chains, has expressions, or function calls

3. **New helper functions** you'll create:
   - `parseTraversal()` - Build traversal expression chains
   - `parseHasExpression()` - Parse has operator
   - `parseFunctionCall()` - Parse function calls with arguments

### Operator Precedence Integration

The precedence chain needs to be updated:

**Current chain (Module 4):**
```
parseExpression
  → parseOrExpression
    → parseAndExpression
      → parseNotExpression
        → parseComparison
          → parseValue (literals, identifiers, grouping)
```

**Updated chain (Module 5):**
```
parseExpression
  → parseOrExpression
    → parseAndExpression
      → parseNotExpression
        → parseComparison
          → parseHas
            → parseTraversal
              → parseValue (literals, identifiers, function calls, grouping)
```

## 📝 Tasks

### ⭐ Start Here: Task 1 - Field Traversal (`.` operator)

**Goal:** Parse expressions like `user.name`, `a.b.c.d`

**Where to add code:** Find the TODO comment in `parseIdentifier()` function (around line 310 in `parser.go`)

**What to implement:**
1. After creating the initial identifier, check if the next token is DOT
2. If yes, enter a loop: consume DOT, expect IDENT, build TraversalExpression
3. Continue looping while more DOTs are found
4. Return the completed traversal (or just the identifier if no DOT)

**Run this test to check your work:**
```bash
go test ./pkg/filter/parser -run TestFieldTraversal -v
```

**Success criteria:** You should see 4 tests pass:
- Simple traversal: `user.name`
- Multi-level: `a.b.c`  
- Traversal in comparison: `user.name = "John"`
- No breaking simple identifiers: `user` still works

**Need help?** Check HINTS.md for the loop pattern and example code structure.

---

### Task 2: Has Operator (`:`)

**Goal:** Parse expressions like `tags:urgent`, `roles:"admin"`

**Where to add code:** 
1. Find the TODO for `parseHasExpression()` (around line 320)
2. Find the TODO in `parseComparison()` to update the precedence chain (around line 180)

**What to implement:**
1. Create new function `parseHasExpression()` similar to `parseComparison()`
2. Inside it: get left side from `parseValue()`, check for HAS token, get right side
3. Update `parseComparison()` to call `parseHasExpression()` instead of `parseValue()`

**Run this test to check your work:**
```bash
go test ./pkg/filter/parser -run TestHasOperator -v
```

**Success criteria:** You should see 3 tests pass:
- Simple has: `tags:urgent`
- Has with traversal: `user.roles:admin`
- Has precedence: `tags:a OR tags:b` parses as `(tags:a) OR (tags:b)Add a new function `parseHasExpression()` that sits between parseComparison and parseTraversal
- After getting a left expression (from parseTraversal), check if current token is HAS
- If yes, parse the right side and create a HasExpression node
- Update the precedence chain to call parseHasExpression from parseComparison
Need help?** Check HINTS.md for the pattern structure and precedence explanation.

---

### Task 3: Function Calls

**Goal:** Parse expressions like `timestamp(created_at)`, `has(tags, "urgent")`

**Where to add code:**
1. Find the TODO for function detection in `parseIdentifier()` (around line 307)
2. Find the TODO for `parseFunctionCall()` helper (around line 335)

**What to implement:**
1. In `parseIdentifier()`: After saving identifier, check if peek token is LPAREN
2. If yes, call `parseFunctionCall()` helper (passing identifier token and name)
3. Create `parseFunctionCall()` helper that:
   - Consumes `(`
   - Parses comma-separated expression list
   - Consumes `)`
   - Returns FunctionCall node

**Run this test to check your work:**
```bash
go test ./pkg/filter/parser -run TestFunctionCall -v
```

**Success criteria:** You should see 4 tests pass:
- Simple function: `timestamp(created_at)`
- Multiple args: `has(tags, "urgent")`
- No args: `now()`
**Need help?** Check HINTS.md for the argument parsing loop pattern.

---

### Task 4: Integration & Complex Expressions

**Goal:** Ensure all features work together correctly

**Where to test:** No new code needed - just verify everything integrates properly

**What should work:**
- Combining features: `user.permissions:write AND status = "active"`
- Nested traversals: `a.b.c:value`
- Functions with traversal: `timestamp(user.created_at) > timestamp("2024-01-01")`
- Functions in has: `tags:get_priority()`

**Run full test suite:**
```bash
go test ./pkg/filter/parser -v
```

**Success criteria:** All 30 tests pass (17 from Module 4 + 13 from Module 5)

---

## 🧪 Test-Driven Development

**Recommended workflow for each task:**

1. **Red** - Run the test, watch it fail
   ```bash
   go test ./pkg/filter/parser -run TestFieldTraversal -v
   ```

2. **Green** - Implement just enough code to make it pass
   - Find the TODO comment in parser.go
   - Add your implementation
   - Refer to HINTS.md if stuck

3. **Verify** - Run the test again
   ```bash
   go test ./pkg/filter/parser -run TestFieldTraversal -v
   ```

4. **Refactor** - Clean up code while tests still pass

5. **Move to next task** - Repeat for Task 2, then Task 3
## 🧪 Test-Driven Development

Each task has associated tests. Follow this workflow:

1. **Run test** - See it fail
2. **Implement** - Write code to make it pass
3. **Verify** - Run test again, see it pass
4. **Refactor** - Clean up code while tests still pass

## ❓ Questions to Consider

1. **Why is traversal left-associative?**  
   Think about: `user.profile.name` - which part evaluates first?

2. **Why does has have higher precedence than comparison?**  
   Compare: `tags:urgent = true` vs `tags:(urgent = true)`

3. **How do you distinguish a function call from an identifier?**  
   What token helps you decide?

4. **Can you have a function inside a traversal?**  
   Is `get_user().name` valid in this grammar?

5. **What happens with `user.name.first.last`?**  
   Draw the AST tree structure.

## 🚀 Getting Started

**Your starting point:** `pkg/filter/parser/parser.go` (with Module 4 complete)

**Step 1:** Find the TODOs
```bash
grep -n "TODO (Module 5" pkg/filter/parser/parser.go
```

You should see 5 TODO comments:
- Line ~180: Update parseComparison to call parseHasExpression
- Line ~307: Add function call detection in parseIdentifier  
- Line ~313: Add field traversal loop in parseIdentifier
- Line ~320: Implement parseHasExpression function
- Line ~335: Implement parseFunctionCall function

**Step 2:** Start with Task 1 (scroll up to the Tasks section)

**Step 3:** Follow the test-driven workflow for each task

## 📚 Resources

- **HINTS.md** - Strategic hints and examples for each task
- **SOLUTION.md** - Complete reference implementation with explanations
- **Module 4** - Review your parseIdentifier and parseComparison patterns

## ✅ Completion Criteria

Module 5 is complete when:

- [ ] **Task 1 complete:** All traversal tests pass (TestFieldTraversal)
- [ ] **Task 2 complete:** All has operator tests pass (TestHasOperator)  
- [ ] **Task 3 complete:** All function call tests pass (TestFunctionCall)
- [ ] **Task 4 complete:** Complex integration tests pass (TestComplexNestedExpressions)
- [ ] **No regressions:** All Module 4 tests still pass (TestModule4StillWorks)
- [ ] **Full suite:** `go test ./pkg/filter/parser -v` shows 30/30 tests passing

**Final check:**
```bash
go test ./pkg/filter/parser -v
```

Expected output: `ok` with all tests passing

## 🎓 Key Takeaways

After completing this module, you should understand:

1. **Operator precedence** - How to integrate new operators into an existing precedence chain
2. **Lookahead parsing** - Using peek tokens to distinguish between constructs
3. **Recursive parsing** - Building nested structures like traversal chains
4. **List parsing** - Handling comma-separated argument lists
5. **AST composition** - How complex expressions are built from simpler nodes

---

**Next:** Module 6 - Evaluation Engine

**Previous:** [Module 4 - Parser Core](../module4-parser-core/README.md)

**Tutorial Home:** [Back to Tutorial Index](../README.md)
