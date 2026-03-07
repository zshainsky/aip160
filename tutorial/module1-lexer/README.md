# Module 1: Lexer (Tokenization)

**Duration**: ~2 hours  
**Difficulty**: Beginner-Intermediate

## 🎯 Module Objectives

By the end of this module, you will:
1. Understand what tokens are and why they're needed
2. Learn the complete lexing process
3. Implement a lexer that tokenizes AIP-160 filter strings
4. Handle different token types: keywords, operators, literals, identifiers
5. Build the foundation for the parser

## 📖 What is Tokenization?

Tokenization (lexical analysis) is the **first step** in parsing. It converts a raw string into a sequence of meaningful units called **tokens**.

### Example: From String to Tokens

**Input String:**
```
name = "John" AND age > 25
```

**Output Tokens:**
```
IDENTIFIER("name")
EQUALS
STRING("John")
AND
IDENTIFIER("age")
GREATER_THAN
NUMBER(25)
EOF
```

### Visual Representation

```
    Raw Filter String
           ↓
    ┌──────────────┐
    │    LEXER     │  ← You build this!
    └──────────────┘
           ↓
      Token Stream
           ↓
    ┌──────────────┐
    │    PARSER    │  ← Later modules
    └──────────────┘
           ↓
    Abstract Syntax Tree (AST)
```

## 🧩 Understanding Tokens

A **token** is a categorized piece of the input string. Think of it like reading a sentence - your brain automatically groups characters into words, punctuation, etc.

### Token Categories for AIP-160

```
┌─────────────────────────────────────────────────────────┐
│                    TOKEN TYPES                          │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  LITERALS          STRING, NUMBER, TRUE, FALSE, NULL    │
│                    "hello", 42, 3.14, true              │
│                                                         │
│  IDENTIFIERS       field names, function names          │
│                    name, age, user.email                │
│                                                         │
│  OPERATORS         =, !=, <, >, <=, >=, :               │
│                                                         │
│  LOGICAL           AND, OR, NOT, -                      │
│                                                         │
│  DELIMITERS        (, ), .                              │
│                                                         │
│  SPECIAL           *, EOF, ILLEGAL                      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

## 🔍 The Lexing Process

Let's trace through tokenizing a filter step-by-step:

### Example: `age > 18 AND status = "active"`

```
Step 1: Read 'a' → Continue reading identifier
Step 2: Read 'g' → Still identifier
Step 3: Read 'e' → Still identifier
Step 4: Read ' ' → Identifier complete! Token: IDENTIFIER("age")
Step 5: Read ' ' → Skip whitespace
Step 6: Read '>' → Single character operator! Token: GREATER_THAN
Step 7: Read ' ' → Skip whitespace
Step 8: Read '1' → Start number
Step 9: Read '8' → Still number
Step 10: Read ' ' → Number complete! Token: NUMBER(18)
... and so on
```

### State Machine Diagram

The lexer is essentially a state machine:

```
          ┌─────────┐
    ┌────→│  START  │←────┐
    │     └─────────┘     │
    │          │          │
    │     (read char)     │
    │          │          │
    │          ↓          │
    │     ┌───────────────────────┐
    │     │   Determine State     │
    │     └───────────────────────┘
    │          │
    │     ┌────┼────┬────┬────┬────┐
    │     ↓    ↓    ↓    ↓    ↓    ↓
    │   Letter Digit " ' Operator Whitespace
    │     │    │    │    │    │    │
    │     ↓    ↓    ↓    ↓    ↓    │
    │   IDENT NUMBER STRING OP │   │
    │     │    │    │    │    │    │
    └─────┴────┴────┴────┴────┴────┘
         (emit token)
```

## 💡 Key Concepts

### 1. **Lookahead**
Sometimes you need to peek at the next character to decide what token you're building.

Example: When you see `>`, you need to check if the next char is `=` to distinguish:
- `>` → GREATER_THAN
- `>=` → GREATER_EQUAL

### 2. **Whitespace Handling**
In AIP-160, whitespace is mostly insignificant (except inside strings). The lexer handles whitespace differently depending on its state:

**Normal State** (default):
- Whitespace between tokens is **skipped** via `skipWhitespace()`
- Acts as a separator: `age>18` and `age > 18` produce identical tokens

**STRING State** (inside quotes):
- Whitespace is **preserved** as part of the string content
- `"hello world"` → STRING("hello world") with space intact

```go
// Normal state - whitespace skipped
name = "John"
  ↓
IDENTIFIER("name"), EQUALS, STRING("John")

// Inside string state - whitespace preserved
"John Doe"
  ↓
STRING("John Doe")  ← Space kept!
```

This state-dependent handling is why `NextToken()` calls `skipWhitespace()` first, but `readString()` preserves all characters until the closing delimiter.

### 3. **String Literals**
Strings can be delimited by `"` or `'`. You must:
- Track which delimiter started the string
- Continue until finding the matching delimiter
- Handle escape sequences (optional for basic implementation)

### 4. **Keywords vs Identifiers**
Words like `AND`, `OR`, `NOT` are keywords. Everything else is an identifier.

```
AND     → Keyword (token: AND)
and     → Keyword (case-insensitive)
name    → Identifier (token: IDENTIFIER)
userName → Identifier (token: IDENTIFIER)
```

### 5. **Numbers**
Numbers can be integers or floats:
- `42` → INTEGER
- `3.14` → FLOAT
- `2.997e9` → FLOAT (with exponent)

## 📝 Implementation Task

You'll implement two main components:

### 1. Token Definition (`token.go`)
Define:
- Token types (constants)
- Token struct (type and value)
- Helper functions

### 2. Lexer (`lexer.go`)
Implement:
- `Lexer` struct with position tracking
- `NextToken()` method - returns the next token
- Helper methods for reading characters, identifiers, numbers, strings

## 🎯 Test-Driven Approach

You'll make these tests pass:

1. **Basic tokens**: Identifiers, operators, delimiters
2. **String literals**: Double and single quotes
3. **Numbers**: Integers and floats
4. **Keywords**: AND, OR, NOT
5. **Complex expressions**: Multiple tokens together
6. **Edge cases**: Empty strings, unknown characters

## 🚀 Getting Started

### Step 1: Review the Token Types

Open `pkg/filter/lexer/token.go` and review the token type definitions. These are provided for you.

### Step 2: Understand the Lexer Structure

The `Lexer` struct tracks:
```go
type Lexer struct {
    input        string  // The input filter string
    position     int     // Current position in input
    readPosition int     // Next position to read
    ch           byte    // Current character
}
```

### Step 3: Run the Tests

```bash
cd /Users/zshainky/Projects/aip160
go test ./pkg/filter/lexer -v
```

You'll see failing tests. Your job: make them pass!

### Step 4: Implement the Lexer

Key methods to implement:
- `New(input string) *Lexer` - Create a new lexer
- `readChar()` - Advance to next character  
- `NextToken() Token` - Main method! Return the next token
- `readIdentifier() string` - Read an identifier
- `readNumber() string` - Read a number
- `readString(delimiter byte) string` - Read a string literal
- `skipWhitespace()` - Skip spaces, tabs, newlines

## 💭 Reflection Questions

As you implement, think about:

1. **Why separate lexing from parsing?**  
   *Hint: What if you wanted to support multiple input formats?*

2. **How does the lexer handle errors?**  
   *What happens when you encounter an unexpected character?*

3. **Why track position?**  
   *How would this help with error messages later?*

4. **What if AIP-160 added a new operator?**  
   *How easy is it to extend your lexer?*

## 🎓 Learning Checkpoint

Before moving to the next module, make sure you can:

- [ ] Explain what a token is
- [ ] Describe the lexing process step-by-step
- [ ] Identify token types in a filter string
- [ ] Explain why `OR` and `AND` are keywords but `name` is an identifier
- [ ] Handle both `"string"` and `'string'` formats
- [ ] All tests pass! ✅

## ❓ Common Questions & Pitfalls

These are real questions from learners like you! Understanding these will save you debugging time.

<details>
<summary>Q: What is "IDENT" in the state machine diagram?</summary>

**A:** IDENT is short for **IDENTIFIER**. It's abbreviated for space in the diagram.

When the lexer sees a letter or underscore, it enters the IDENT state and reads the complete identifier (e.g., `name`, `user_id`, `age`). After reading, it checks if it's a keyword like `AND` or a regular identifier.
</details>

<details>
<summary>Q: Are digits considered IDENTifiers?</summary>

**A:** **No!** It depends on what the token **starts with**:

- `123abc` → Starts with digit → **NUMBER** state (may be invalid later)
- `abc123` → Starts with letter → **IDENTIFIER** state → Valid identifier ✅
- `_123` → Starts with underscore → **IDENTIFIER** state → Valid identifier ✅

**Rule:** First character determines the token type. IDENTIFIERs can *contain* digits but can't *start* with them.
</details>

<details>
<summary>Q: Is `readPosition` always 1 ahead of `position`?</summary>

**A:** **Yes!** This invariant is maintained after every `readChar()` call:

```go
readPosition = position + 1
```

This lets you:
- Know the **current** character (`position`)
- **Peek** at the next character (`readPosition`) without advancing

Example:
```
Input: "age"
Start: position=0, readPosition=1, ch='a'
After readChar(): position=1, readPosition=2, ch='g'
```
</details>

<details>
<summary>Q: Go strings vs bytes - Why can't I do `ident += l.input[l.position]`?</summary>

**A:** In Go, indexing a string returns a **byte**, not a string:

```go
str := "hello"
str[0]  // Returns byte 'h', not string "h"
```

**Solutions:**
```go
// ❌ Wrong: Can't concatenate byte to string
ident += l.input[l.position]

// ✅ Option 1: Convert byte to string
ident += string(l.input[l.position])

// ✅ Option 2: Use slicing (recommended - faster!)
position := l.position
// ... advance through identifier ...
return l.input[position:l.position]
```
</details>

<details>
<summary>Q: Byte arithmetic mistake - Why is `>=` becoming a weird character?</summary>

**A:** Common bug when building two-character operators:

```go
// ❌ WRONG: Byte arithmetic!
literal := l.ch        // literal is a BYTE (ASCII value)
literal += l.ch        // Adds ASCII values: 62 + 61 = 123 → '{'

// ✅ CORRECT: String concatenation
literal := string(l.ch)      // Convert to string first
l.readChar()
literal += string(l.ch)      // Now it's string concatenation
// Result: ">=" ✅
```

**Remember:** Bytes are uint8 numbers. Adding them does arithmetic, not concatenation!
</details>

<details>
<summary>Q: readString() - Why are quotes included in my string output?</summary>

**A:** You need to advance past the closing quote **after** extracting the string:

```go
// When loop breaks, l.position is AT the closing quote
for {
    l.readChar()
    if l.ch == delimiter || l.ch == 0 {
        break  // ← l.position is AT the quote!
    }
}

// ❌ Wrong: Next token will see the quote again!
return l.input[position:l.position]

// ✅ Correct: Save result, then advance past quote
result := l.input[position:l.position]
l.readChar()  // Move past closing quote
return result
```

This ensures the next `NextToken()` call starts at the right position.
</details>

<details>
<summary>Q: readNumber() - What if the number is malformed like "2e" or "3."?</summary>

**A:** **The lexer doesn't validate!** Its job is to extract tokens, not validate them.

```go
// Just return whatever you read
"2e"   → Token{Type: NUMBER, Literal: "2e"}
"3."   → Token{Type: NUMBER, Literal: "3"}  (dot not consumed)
```

Validation happens later when **parsing or evaluating**:
```go
val, err := strconv.ParseFloat(token.Literal, 64)
if err != nil {
    return fmt.Errorf("invalid number: %s", token.Literal)
}
```

**Principle:** Lexer extracts, Parser validates.
</details>

<details>
<summary>Q: Empty string input - Will `New()` panic?</summary>

**A:** Yes, if you directly access `input[0]`!

```go
// ❌ WRONG: Panics on empty input
func New(input string) *Lexer {
    return &Lexer{
        input: input,
        ch: input[0],  // ← Panic if input is ""
    }
}

// ✅ CORRECT: Let readChar() handle it
func New(input string) *Lexer {
    l := &Lexer{input: input}
    l.readChar()  // Safely sets ch=0 if empty
    return l
}
```
</details>

## 📚 Resources

### Implementation Help

**Need hints?** → [HINTS.md](HINTS.md) - Progressive hints for each function without spoiling the solution

**Really stuck?** → [SOLUTION.md](SOLUTION.md) - Complete implementation with detailed explanations (try HINTS.md first!)

### Additional Learning

- [How Lexers Work](https://en.wikipedia.org/wiki/Lexical_analysis)
- [State Machine Basics](https://en.wikipedia.org/wiki/Finite-state_machine)

---

**Ready to code?** Check out the test file: `../../pkg/filter/lexer/lexer_test.go`

**Need the solution?** See [SOLUTION.md](SOLUTION.md) (try implementing first!)

**Next Module**: [Module 2: AST Design](../module2-ast/README.md) (coming soon)
