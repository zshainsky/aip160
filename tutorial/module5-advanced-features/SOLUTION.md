# Module 5: Solution - Advanced Parser Features

This document contains the complete solution for Module 5. Try to implement it yourself first using the README and HINTS before looking here!

## Overview of Changes

Module 5 extends the parser with three features:
1. Field traversal (`.` operator)
2. Has operator (`:`)
3. Function calls

These modifications happen in a few key places in your existing parser.go file.

---

## Complete Implementation

### Modified Functions

#### 1. parseIdentifier() - Extended for Traversal and Function Calls

The `parseIdentifier()` function now does more than just create an Identifier node. It checks for DOT (traversal) and LPAREN (function call).

```go
// parseIdentifier parses an identifier, field traversal, or function call
func (p *Parser) parseIdentifier() ast.Expression {
	// Start with the identifier token
	identToken := p.currentToken
	identValue := p.currentToken.Literal
	
	// Advance past the identifier
	p.nextToken()
	
	// Check if this is a function call
	if p.currentTokenIs(lexer.LPAREN) {
		return p.parseFunctionCall(identToken, identValue)
	}
	
	// Create base identifier
	expr := &ast.Identifier{
		Token: identToken,
		Value: identValue,
	}
	
	// Check for field traversal (dot notation)
	for p.currentTokenIs(lexer.DOT) {
		dotToken := p.currentToken
		p.nextToken() // consume the DOT
		
		// After DOT, we expect an identifier
		if !p.currentTokenIs(lexer.IDENTIFIER) {
			msg := fmt.Sprintf("expected identifier after '.', got %s", p.currentToken.Type)
			p.errors = append(p.errors, msg)
			return nil
		}
		
		// Create identifier for the right side
		right := &ast.Identifier{
			Token: p.currentToken,
			Value: p.currentToken.Literal,
		}
		
		p.nextToken() // consume the identifier
		
		// Build traversal expression
		expr = &ast.TraversalExpression{
			Token: dotToken,
			Left:  expr,
			Right: right,
		}
	}
	
	return expr
}
```

**Key Points:**
- Save the initial identifier token before advancing
- Check for LPAREN first (functions bind tightest) 
- Use a loop for traversal to handle `a.b.c.d`
- Each iteration builds a new TraversalExpression with previous result as left
- After DOT, expect an IDENTIFIER token

---

#### 2. parseFunctionCall() - New Function

This is a new helper function for parsing function calls with arguments.

```go
// parseFunctionCall parses a function call with arguments
// Called when we've seen: IDENTIFIER LPAREN
func (p *Parser) parseFunctionCall(token lexer.Token, functionName string) ast.Expression {
	// We're currently on LPAREN, advance past it
	p.nextToken()
	
	// Parse arguments
	arguments := []ast.Expression{}
	
	// Check for empty argument list
	if p.currentTokenIs(lexer.RPAREN) {
		p.nextToken() // consume closing paren
		return &ast.FunctionCall{
			Token:     token,
			Function:  functionName,
			Arguments: arguments,
		}
	}
	
	// Parse first argument
	arg := p.parseExpression()
	if arg != nil {
		arguments = append(arguments, arg)
	}
	
	// Parse remaining arguments (comma-separated)
	for p.currentTokenIs(lexer.COMMA) {
		p.nextToken() // consume comma
		arg := p.parseExpression()
		if arg != nil {
			arguments = append(arguments, arg)
		}
	}
	
	// Expect closing parenthesis
	if !p.currentTokenIs(lexer.RPAREN) {
		msg := fmt.Sprintf("expected ')' in function call, got %s", p.currentToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	
	p.nextToken() // consume closing paren
	
	return &ast.FunctionCall{
		Token:     token,
		Function:  functionName,
		Arguments: arguments,
	}
}
```

**Key Points:**
- Receives the saved identifier token and name
- When called, current token is LPAREN
- Advance past LPARENto start parsing arguments
- Handle empty argument list first (immediate RPAREN)
- Parse first argument, then loop for comma-separated rest
- Each argument is a full expression (use parseExpression)
- Expect RPAREN at end

---

#### 3. parseComparison() - Modified to Call parseHasExpression

Update parseComparison to call the new has expression parser:

```go
// parseComparison handles comparison operators
// Grammar: comparison = has_expression, [comparator, has_expression] ;
func (p *Parser) parseComparison() ast.Expression {
	left := p.parseHasExpression()
	
	if p.isComparisonOperator(p.currentToken.Type) {
		opToken := p.currentToken
		p.nextToken()
		right := p.parseHasExpression()
		
		return &ast.ComparisonExpression{
			Token:    opToken,
			Left:     left,
			Operator: opToken.Literal,
			Right:    right,
		}
	}
	
	return left
}
```

**Key Points:**
- Changed from calling `parseValue()` to calling `parseHasExpression()`
- This inserts has operator into the precedence chain
- Everything else stays the same

---

#### 4. parseHasExpression() - New Function

This is a new function that handles the has operator (`:`).

```go
// parseHasExpression handles has operator for collection membership
// Grammar: has_expression = primary, [":", primary] ;
func (p *Parser) parseHasExpression() ast.Expression {
	left := p.parseValue()
	
	if p.currentTokenIs(lexer.HAS) {
		hasToken := p.currentToken
		p.nextToken() // consume ':'
		right := p.parseValue()
		
		return &ast.HasExpression{
			Token:      hasToken,
			Collection: left,
			Member:     right,
		}
	}
	
	return left
}
```

**Key Points:**
- Pattern similar to parseComparison
- Get left from parseValue (which handles identifiers with traversal)
- Check for HAS token (`:`)
- Parse right side (also from parseValue)
- Return HasExpression or just left if no HAS operator

---

## Precedence Chain Summary

The complete precedence chain after Module 5:

```
parseExpression
  ↓
parseOrExpression (OR - lowest precedence)
  ↓
parseAndExpression (AND)
  ↓
parseNotExpression (NOT)
  ↓
parseComparison (=, !=, <, >, <=, >=)
  ↓
parseHasExpression (:)
  ↓
parseValue (dispatches to specific parsers)
  ↓
parseIdentifier (which handles ., (, or just returns Identifier)
  OR
parseString / parseNumber / parseBoolean / parseNull / parseGroupedExpression
```

---

## Example Parse Traces

### Example 1: `user.name = "John"`

1. parseExpression → parseOrExpression → parseAndExpression → parseNotExpression → parseComparison
2. parseComparison calls parseHasExpression for left
3. parseHasExpression calls parseValue
4. parseValue calls parseIdentifier
5. parseIdentifier creates Identifier("user")
6. parseIdentifier sees DOT, loops:
   - Creates Identifier("name")
   - Builds TraversalExpression{Left: Identifier("user"), Right: Identifier("name")}
7. Returns TraversalExpression to parseHasExpression
8. No HAS operator, returns to parseComparison
9. parseComparison sees EQUALS, parses right side: StringLiteral("John")
10. Returns ComparisonExpression

**AST:**
```
    =
   / \
  .   "John"
 / \
user name
```

### Example 2: `tags:urgent`

1. ... → parseComparison → parseHasExpression
2. parseHasExpression calls parseValue → parseIdentifier
3. parseIdentifier creates Identifier("tags"), no DOT or LPAREN
4. Returns Identifier("tags") to parseHasExpression
5. parseHasExpression sees HAS (`:`)
6. Parses right side: Identifier("urgent")
7. Returns HasExpression{Collection: Identifier("tags"), Member: Identifier("urgent")}

**AST:**
```
     :
    / \
 tags  urgent
```

### Example 3: `timestamp(created_at) > 1000`

1. ... → parseComparison → parseHasExpression → parseValue → parseIdentifier
2. parseIdentifier gets "timestamp", sees LPAREN
3. Calls parseFunctionCall:
   - Advances past LPAREN
   - Parses argument: Identifier("created_at")
   - Consumes RPAREN
   - Returns FunctionCall{Function: "timestamp", Arguments: [Identifier("created_at")]}
4. Back in parseComparison, sees GREATER_THAN
5. Parses right: NumberLiteral(1000)
6. Returns ComparisonExpression

**AST:**
```
        >
       / \
  timestamp()  1000
      |
  created_at
```

### Example 4: `user.tags:admin AND status = "active"`

1. ... → parseAndExpression
2. Left side: ... → parseComparison → parseHasExpression → parseValue
3. parseValue → parseIdentifier gets "user"
4. Sees DOT, builds Traversal: `user.tags`
5. Back in parseHasExpression, sees HAS
6. Right side: Identifier("admin")
7. Returns HasExpression
8. parseComparison sees no comparison, returns HasExpression
9. parseAndExpression sees AND token
10. Right side: parseValue → ... → ComparisonExpression: `status = "active"`
11. Returns LogicalExpression

**AST:**
```
          AND
         /   \
        :     =
       / \   / \
      .   admin status "active"
     / \
  user tags
```

---

## Testing Your Implementation

Run tests individually to verify each feature:

```bash
# Test traversal
go test ./pkg/filter/parser -run TestFieldTraversal -v

# Test has operator
go test ./pkg/filter/parser -run TestHasOperator -v

# Test function calls
go test ./pkg/filter/parser -run TestFunctionCall -v

# Test complex combinations
go test ./pkg/filter/parser -run TestComplex -v

# Run all parser tests
go test ./pkg/filter/parser -v
```

All tests should pass, including the tests from Module 4.

---

## Common Implementation Mistakes

### Mistake 1: Not Saving Tokens

❌ **Wrong:**
```go
p.nextToken()
return &ast.TraversalExpression{
	Token: p.currentToken, // This is the WRONG token now!
	...
}
```

✅ **Correct:**
```go
dotToken := p.currentToken
p.nextToken()
return &ast.TraversalExpression{
	Token: dotToken, // Saved token
	...
}
```

### Mistake 2: Wrong Precedence Integration

❌ **Wrong:** Calling parseValue from parseComparison  
✅ **Correct:** Calling parseHasExpression from parseComparison

This ensures has operator has correct precedence.

### Mistake 3: Parsing Arguments Wrong

❌ **Wrong:** Using parseIdentifier for arguments  
✅ **Correct:** Using parseExpression for arguments

Arguments can be any expression, not just identifiers.

### Mistake 4: Not Handling Current Token Position

After parseIdentifier with traversal `user.name`, current token should be the NEXT token (operator/EOF), not still on "name".

Each parse function advances to the first token it doesn't consume.

---

## Design Decisions Explained

### Why Traverse in parseIdentifier?

**Decision:** Handle traversal in parseIdentifier rather than a separate function.

**Rationale:**
- Traversal always starts with an identifier
- Cleaner to check for DOT right after parsing identifier
- Avoids extra precedence level

**Alternative:** Could have parseTraversal as separate function between parseHasExpression and parseValue. Both approaches work.

### Why Loop for Traversal?

**Decision:** Use while loop to handle multiple dots.

**Rationale:**
- `a.b.c` needs multiple TraversalExpressions
- Loop builds left-associative tree naturally
- Same pattern as parseAndExpression

### Why Check LPAREN in parseIdentifier?

**Decision:** Detect function calls immediately after parsing identifier.

**Rationale:**
- Function call looks like identifier followed by `(`
- Need lookahead to distinguish `func` from `func()`
- Can't wait until later - need to consume arguments now

### Why parseExpression for Arguments?

**Decision:** Use parseExpression, not parseValue or parseIdentifier.

**Rationale:**
- Arguments can be complex: `func(age > 18, status = "active")`
- parseExpression handles full expressions including operators
- Gives maximum flexibility

---

## Verifying Your Solution

Checklist:
- [ ] Simple traversal works: `a.b`
- [ ] Nested traversal works: `a.b.c.d`
- [ ] Simple has works: `tags:urgent`
- [ ] Traversal + has works: `user.tags:admin`
- [ ] Function with no args works: `now()`
- [ ] Function with one arg works: `timestamp(created_at)`
- [ ] Function with multiple args works: `has(tags, "urgent")`
- [ ] Function with complex args works: `func(age > 18)`
- [ ] Complex combinations work: `user.permissions:write AND func(status) = "ok"`
- [ ] All Module 4 tests still pass

---

## Next Steps

After completing Module 5, you have a fully-featured parser that can handle:
- Literals (strings, numbers, booleans, null)
- Identifiers and field traversal
- Comparison operators
- Logical operators (AND, OR, NOT)
- Has operator for collections
- Function calls
- Grouping with parentheses

**Next Module:** Module 6 will implement the evaluation engine that takes your AST and runs it against real data!

---

**Tutorial Home:** [Back to Module 5](README.md)
