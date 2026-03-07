# Module 1: Lexer - Implementation Hints

Need a nudge in the right direction? These hints will help you implement the lexer without giving away the complete solution.

## 🎯 General Approach

The lexer follows a simple pattern:
1. Read character by character
2. Determine what kind of token is starting
3. Read the complete token
4. Return it and move to the next

Take it one function at a time and test frequently!

---

## Hint 1: Implementing `New()`

The `New()` function needs to:
- Create a Lexer instance with the input string
- Initialize the lexer's state
- Read the first character

```go
func New(input string) *Lexer {
    l := &Lexer{input: input}
    l.readChar()  // Initialize by reading first character
    return l
}
```

**Why call `readChar()` instead of setting fields directly?**  
It handles the empty input case safely and sets up the position tracking correctly.

---

## Hint 2: Implementing `readChar()`

This method advances through the input one character at a time.

**Key steps:**
1. Check if we've reached the end of input
2. If yes → set `ch` to 0 (EOF marker)
3. If no → read the character at `readPosition`
4. Update `position` to catch up to `readPosition`
5. Advance `readPosition` by 1

```go
func (l *Lexer) readChar() {
    if l.readPosition >= len(l.input) {
        l.ch = 0  // EOF
    } else {
        l.ch = l.input[l.readPosition]
    }
    l.position = l.readPosition
    l.readPosition++
}
```

---

## Hint 3: Implementing `peekChar()`

Similar to `readChar()` but **doesn't advance** the position.

```go
func (l *Lexer) peekChar() byte {
    if l.readPosition >= len(l.input) {
        return 0  // EOF
    }
    return l.input[l.readPosition]
}
```

**Important:** Don't modify `l.ch` in peek! It's read-only.

---

## Hint 4: Implementing `skipWhitespace()`

Keep calling `readChar()` while the current character is whitespace.

Whitespace characters to skip:
- Space: `' '`
- Tab: `'\t'`
- Newline: `'\n'`
- Carriage return: `'\r'`

```go
func (l *Lexer) skipWhitespace() {
    for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
        l.readChar()
    }
}
```

---

## Hint 5: Implementing `readIdentifier()`

Save the starting position, then keep advancing while you see valid identifier characters.

**Valid identifier characters:**
- Must start with letter or underscore
- Can contain letters, digits, or underscores

```go
func (l *Lexer) readIdentifier() string {
    position := l.position
    
    // Keep reading while letter or digit
    for isLetter(l.ch) || isDigit(l.ch) {
        l.readChar()
    }
    
    // Extract the identifier using slice
    return l.input[position:l.position]
}
```

**Note:** Don't call `readChar()` after returning - the helper functions handle advancement!

---

## Hint 6: Implementing `readNumber()`

Numbers have three optional parts: integer, decimal, exponent.

**Structure:**
```
Integer part: 42
             ↓
Decimal part: .14 (optional, must have digits after '.')
             ↓
Exponent: e9 or E-10 (optional, can have +/- sign)
```

**Implementation approach:**
```go
func (l *Lexer) readNumber() string {
    position := l.position
    
    // 1. Read integer part (digits)
    for isDigit(l.ch) {
        l.readChar()
    }
    
    // 2. Check for decimal part
    if l.ch == '.' && isDigit(l.peekChar()) {
        l.readChar()  // consume '.'
        for isDigit(l.ch) {
            l.readChar()
        }
    }
    
    // 3. Check for exponent part
    if l.ch == 'e' || l.ch == 'E' {
        l.readChar()  // consume 'e' or 'E'
        
        // Optional sign
        if l.ch == '+' || l.ch == '-' {
            l.readChar()
        }
        
        // Exponent digits
        for isDigit(l.ch) {
            l.readChar()
        }
    }
    
    return l.input[position:l.position]
}
```

---

## Hint 7: Implementing `readString()`

Read until you find the matching closing delimiter or EOF.

**Approach:**
```go
func (l *Lexer) readString(delimiter byte) string {
    position := l.position + 1  // Start after opening quote
    
    // Read until closing delimiter or EOF
    for {
        l.readChar()
        if l.ch == delimiter || l.ch == 0 {
            break
        }
    }
    
    // Save result before advancing
    result := l.input[position:l.position]
    
    // Advance past closing delimiter
    l.readChar()
    
    return result
}
```

**Why save before the final `readChar()`?**  
Because `l.position` points at the closing quote. After we advance, it points past it, and we'd include the quote in our slice!

---

## Hint 8: Implementing `NextToken()` - Structure

Use a switch statement to handle different character types:

```go
func (l *Lexer) NextToken() Token {
    l.skipWhitespace()
    
    var tok Token
    
    switch l.ch {
    case 0:
        tok = Token{Type: EOF, Literal: ""}
        l.readChar()
        return tok
        
    case '(':
        tok = Token{Type: LPAREN, Literal: string(l.ch)}
        l.readChar()
        return tok
        
    // ... more single-char operators
    
    case '>':
        // Two-character operator logic here
        
    // ... more cases
        
    default:
        // Handle identifiers, numbers, strings
    }
    
    return tok
}
```

---

## Hint 9: Two-Character Operators

Use `peekChar()` to look ahead:

```go
case '>':
    if l.peekChar() == '=' {
        // It's >=
        ch := l.ch
        l.readChar()  // Move to '='
        tok = Token{Type: GREATER_EQUAL, Literal: string(ch) + string(l.ch)}
    } else {
        // It's just >
        tok = Token{Type: GREATER_THAN, Literal: string(l.ch)}
    }
    l.readChar()  // Advance for next token
    return tok
```

**Key points:**
- Save first character before advancing
- Use string concatenation: `string(ch) + string(l.ch)`
- Don't forget final `readChar()` to advance past the operator!

---

## Hint 10: Handling Identifiers and Keywords

When you encounter a letter, read the full identifier then check if it's a keyword:

```go
if isLetter(l.ch) {
    literal := l.readIdentifier()
    return Token{Type: LookupIdentifier(literal), Literal: literal}
}
```

**Why `LookupIdentifier()`?**  
It checks if the identifier is a keyword (AND, OR, NOT, etc.) and returns the appropriate token type. If not a keyword, it returns IDENTIFIER.

---

## 🐛 Debugging Tips

### Print Current State
```go
fmt.Printf("ch='%c' position=%d readPosition=%d\n", l.ch, l.position, l.readPosition)
```

### Test Small Inputs First
Start with simple cases:
- `"="` - single operator
- `"age"` - simple identifier
- `"42"` - simple number
- `">"` then `">="` - test two-char operators

### Run Tests Incrementally
Don't try to implement everything at once! Implement one function, test it, then move to the next.

---

## 🎯 Common Mistakes to Avoid

1. **Forgetting to advance** - Every token needs to be consumed
2. **Double advancing** - Helper functions (readIdentifier, readNumber, readString) already advance
3. **Byte arithmetic** - Convert to string before concatenating: `string(ch)`
4. **Not checking EOF** - Always handle `readPosition >= len(input)`
5. **Modifying state in peek** - `peekChar()` should only return, not modify `l.ch`

---

**Still stuck?** Check [SOLUTION.md](SOLUTION.md) for the complete implementation with detailed explanations.

**Ready to test?** Run `go test ./pkg/filter/lexer -v` and see how you're doing!
