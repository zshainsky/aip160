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
In AIP-160, whitespace is mostly insignificant (except inside strings). Skip it!

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

## ⚡ Hints

<details>
<summary>Hint 1: Starting the NextToken() method</summary>

Begin by skipping whitespace and checking for EOF:
```go
func (l *Lexer) NextToken() Token {
    l.skipWhitespace()
    
    if l.ch == 0 {
        return Token{Type: EOF, Literal: ""}
    }
    
    // Then check character type and handle accordingly
}
```
</details>

<details>
<summary>Hint 2: Handling multi-character operators</summary>

Use `peekChar()` to look ahead without advancing:
```go
if l.ch == '>' {
    if l.peekChar() == '=' {
        // It's >=
        ch := l.ch
        l.readChar()
        return Token{Type: GREATER_EQUAL, Literal: string(ch) + string(l.ch)}
    }
    // It's just >
    return Token{Type: GREATER_THAN, Literal: string(l.ch)}
}
```
</details>

<details>
<summary>Hint 3: Keywords vs Identifiers</summary>

Create a keyword map:
```go
var keywords = map[string]TokenType{
    "AND": AND,
    "OR":  OR,
    "NOT": NOT,
    // ... etc
}

func lookupIdentifier(ident string) TokenType {
    if tok, ok := keywords[strings.ToUpper(ident)]; ok {
        return tok
    }
    return IDENTIFIER
}
```
</details>

## 📚 Additional Resources

- [How Lexers Work](https://en.wikipedia.org/wiki/Lexical_analysis)
- [State Machine Basics](https://en.wikipedia.org/wiki/Finite-state_machine)

---

**Ready to code?** Check out the test file: `../../pkg/filter/lexer/lexer_test.go`

**Need the solution?** See [SOLUTION.md](SOLUTION.md) (try implementing first!)

**Next Module**: [Module 2: AST Design](../module2-ast/README.md) (coming soon)
