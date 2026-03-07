# Module 1: Lexer - Solution Guide

⚠️ **Try implementing on your own first!** Only refer to this guide if you're stuck.

## 🎯 Solution Strategy

The lexer implementation follows this pattern:

1. **Initialize the lexer** with the input string
2. **Read character by character** tracking position
3. **Identify token type** based on current character
4. **Read full token** (may span multiple characters)
5. **Return the token** and move to next

## � Common Pitfalls (Learn from Others!)

Before diving into hints, know these common mistakes that trip up learners:

### 1. **Byte Arithmetic vs String Concatenation**
```go
// ❌ WRONG: Adds ASCII values!
literal := l.ch        // byte: 62 for '>'
literal += l.ch        // 62 + 61 = 123 = '{'

// ✅ CORRECT: String concatenation
literal := string(l.ch)
literal += string(l.ch)
```

### 2. **Empty Input Panic**
```go
// ❌ WRONG: Panics if input is ""
ch: input[0]

// ✅ CORRECT: Use readChar() to handle it
l := &Lexer{input: input}
l.readChar()  // Sets ch=0 if empty
```

### 3. **Not Advancing Past Closing Quote**
```go
// ❌ WRONG: Next token sees quote again!
return l.input[position:l.position]

// ✅ CORRECT: Advance past closing delimiter
result := l.input[position:l.position]
l.readChar()  // Move to next character
return result
```

### 4. **peekChar() Modifying State**
```go
// ❌ WRONG: Peek shouldn't change l.ch!
func (l *Lexer) peekChar() byte {
    if l.readPosition >= len(l.input) {
        l.ch = 0  // ❌ Modifies state!
    }
    return l.input[l.readPosition]
}

// ✅ CORRECT: Just look, don't touch
func (l *Lexer) peekChar() byte {
    if l.readPosition >= len(l.input) {
        return 0  // ✅ Return without modifying
    }
    return l.input[l.readPosition]
}
```

### 5. **Forgetting LookupIdentifier() for Keywords**
```go
// ❌ WRONG: "AND" becomes IDENTIFIER
return Token{Type: IDENTIFIER, Literal: l.readIdentifier()}

// ✅ CORRECT: Check if it's a keyword
identifier := l.readIdentifier()
return Token{Type: LookupIdentifier(identifier), Literal: identifier}
```

## �💭 Progressive Hints

### Hint 1: Implementing `New()`

<details>
<summary>Click to expand</summary>

The `New()` function should:
1. Create a Lexer instance with the input
2. Initialize positions to 0
3. Read the first character

```go
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Read first character
	return l
}
```

</details>

### Hint 2: Implementing `readChar()`

<details>
<summary>Click to expand</summary>

Handle two cases:
1. If we've reached the end → set `ch` to 0 (EOF marker)
2. Otherwise → read the character at `readPosition`

Then advance both positions.

```go
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}
```

</details>

### Hint 3: Implementing `peekChar()`

<details>
<summary>Click to expand</summary>

Similar to `readChar()` but DON'T advance positions:

```go
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}
```

</details>

### Hint 4: Implementing `skipWhitespace()`

<details>
<summary>Click to expand</summary>

Keep reading while current character is whitespace:

```go
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}
```

</details>

### Hint 5: Implementing `readIdentifier()`

<details>
<summary>Click to expand</summary>

Save starting position, then keep reading while characters are letters/digits/underscores:

```go
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}
```

</details>

### Hint 6: Implementing `readNumber()`

<details>
<summary>Click to expand</summary>

This is more complex because we support:
- Integers: `42`
- Floats: `3.14`
- Scientific notation: `2.997e9`, `1e-10`

```go
func (l *Lexer) readNumber() string {
	position := l.position
	
	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}
	
	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	
	// Check for scientific notation (e or E)
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar() // consume 'e' or 'E'
		
		// Optional + or - sign
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

</details>

### Hint 7: Implementing `readString()`

<details>
<summary>Click to expand</summary>

Read until we find the matching delimiter:

```go
func (l *Lexer) readString(delimiter byte) string {
	position := l.position + 1 // Skip opening quote
	
	for {
		l.readChar()
		
		// Found closing delimiter or EOF
		if l.ch == delimiter || l.ch == 0 {
			break
		}
		
		// Optional: Handle escape sequences
		// if l.ch == '\\' {
		//     l.readChar() // Skip escaped character
		// }
	}
	
	return l.input[position:l.position]
}
```

</details>

### Hint 8: Implementing `NextToken()` - Structure

<details>
<summary>Click to expand</summary>

Use a big switch statement on the current character:

```go
func (l *Lexer) NextToken() Token {
	var tok Token
	
	l.skipWhitespace()
	
	switch l.ch {
	case '=':
		tok = Token{Type: EQUALS, Literal: string(l.ch)}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch)}
		}
	// ... more cases
	case 0:
		tok = Token{Type: EOF, Literal: ""}
	default:
		// Handle identifiers, keywords, numbers
	}
	
	l.readChar()
	return tok
}
```

</details>

## ✅ Complete Solution

<details>
<summary>Click to see the full implementation</summary>

```go
package lexer

// Lexer performs lexical analysis on AIP-160 filter strings
type Lexer struct {
	input        string // the input string being tokenized
	position     int    // current position in input (points to current char)
	readPosition int    // current reading position in input (after current char)
	ch           byte   // current char under examination
}

// New creates a new Lexer instance for the given input string
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Initialize by reading the first character
	return l
}

// readChar reads the next character and advances the position in the input
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII code for "NUL" (end of file)
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// peekChar looks at the next character without advancing the position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = Token{Type: EQUALS, Literal: string(l.ch)}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQUALS, Literal: string(ch) + string(l.ch)}
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch)}
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LESS_EQUAL, Literal: string(ch) + string(l.ch)}
		} else {
			tok = Token{Type: LESS_THAN, Literal: string(l.ch)}
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GREATER_EQUAL, Literal: string(ch) + string(l.ch)}
		} else {
			tok = Token{Type: GREATER_THAN, Literal: string(l.ch)}
		}
	case ':':
		tok = Token{Type: HAS, Literal: string(l.ch)}
	case '.':
		tok = Token{Type: DOT, Literal: string(l.ch)}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch)}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch)}
	case '*':
		tok = Token{Type: ASTERISK, Literal: string(l.ch)}
	case '-':
		tok = Token{Type: MINUS, Literal: string(l.ch)}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString('"')
	case '\'':
		tok.Type = STRING
		tok.Literal = l.readString('\'')
	case 0:
		tok.Literal = ""
		tok.Type = EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdentifier(tok.Literal)
			return tok // Early return - readIdentifier already advanced
		} else if isDigit(l.ch) {
			tok.Type = NUMBER
			tok.Literal = l.readNumber()
			return tok // Early return - readNumber already advanced
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace skips over whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float with optional scientific notation)
func (l *Lexer) readNumber() string {
	position := l.position
	
	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}
	
	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	
	// Check for scientific notation (e or E)
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar() // consume 'e' or 'E'
		
		// Optional + or - sign
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

// readString reads a string literal delimited by quotes
func (l *Lexer) readString(delimiter byte) string {
	position := l.position + 1 // Position after opening quote
	
	for {
		l.readChar()
		
		if l.ch == delimiter || l.ch == 0 {
			break
		}
		
		// Optional: Handle escape sequences for advanced implementation
		// if l.ch == '\\' {
		//     l.readChar() // Skip the next character (escaped)
		// }
	}
	
	return l.input[position:l.position]
}

// isLetter checks if a byte is a letter or underscore
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if a byte is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
```

</details>

## 🧪 Testing Your Implementation

Run the tests:
```bash
go test ./pkg/filter/lexer -v
```

All tests should pass! If not, check:

1. **Are you handling EOF correctly?** (ch == 0)
2. **Are you advancing positions properly?** (position and readPosition)
3. **Are you returning early from NextToken() for identifiers and numbers?** (They already call readChar())
4. **Are you handling two-character operators?** (!=, <=, >=)

## 🎓 Understanding the Solution

### Key Design Decisions

1. **Why track both `position` and `readPosition`?**
   - `position` points to the current character being examined
   - `readPosition` points to the next character to read
   - This allows `peekChar()` to look ahead without changing state

2. **Why return early for identifiers and numbers?**
   - `readIdentifier()` and `readNumber()` already advance past the entire token
   - If we called `readChar()` again after, we'd skip a character!

3. **Why use byte (0) for EOF?**
   - It's a sentinel value not found in normal text
   - Simple to check: `if l.ch == 0`

4. **Why is whitespace handling important?**
   - In AIP-160, whitespace is mostly insignificant
   - Skipping it simplifies the rest of the lexer

## 🚀 Next Steps

Congratulations! You've built a working lexer. You now understand:
- ✅ How tokenization works
- ✅ State machines and position tracking
- ✅ Handling different token types
- ✅ Looking ahead (for multi-character operators)

**Ready for Module 2?** → Build the Abstract Syntax Tree!

## 🤔 Challenge Exercises

Want to extend your lexer? Try adding:

1. **Escape sequences in strings**: Handle `\"`, `\'`, `\\`, `\n`
2. **Comments**: Skip `// comment` and `/* multi-line */`
3. **Duration literals**: Recognize `20s`, `1.5s` as special tokens
4. **Better error reporting**: Track line and column numbers
5. **Unicode support**: Handle UTF-8 identifiers

<details>
<summary>Bonus: Escape Sequences Implementation</summary>

```go
func (l *Lexer) readString(delimiter byte) string {
	var result strings.Builder
	l.readChar() // Skip opening quote
	
	for l.ch != delimiter && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // Skip backslash
			switch l.ch {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '"':
				result.WriteByte('"')
			case '\'':
				result.WriteByte('\'')
			default:
				result.WriteByte(l.ch)
			}
		} else {
			result.WriteByte(l.ch)
		}
		l.readChar()
	}
	
	return result.String()
}
```

</details>

