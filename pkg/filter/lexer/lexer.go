package lexer

// Lexer performs lexical analysis on AIP-160 filter strings
type Lexer struct {
	input        string // the input string being tokenized
	position     int    // current position in input (points to current char)
	readPosition int    // current reading position in input (after current char)
	ch           byte   // current char under examination
}

// New creates a new Lexer instance for the given input string
// Initialize the lexer and read the first character
func New(input string) *Lexer {
	// Your implementation here
	l := &Lexer{input: input}
	l.readChar()
	return l

}

// readChar reads the next character and advances the position in the input
// This should:
// - Set l.ch to the character at l.readPosition
// - Handle EOF by setting l.ch to 0
// - Advance both position and readPosition
func (l *Lexer) readChar() {
	// end of input
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// peekChar looks at the next character without advancing the position
// Useful for two-character operators like >= or !=
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
// This is the main method of the lexer!
// It should:
// 1. Skip whitespace
// 2. Determine the token type based on the current character
// 3. Read the full token (may be multiple characters)
// 4. Return the Token
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok *Token
	switch l.ch {
	// end of string
	case 0:
		tok = &Token{Type: EOF, Literal: ""}
	// single char operators
	case '(':
		tok = &Token{Type: LPAREN, Literal: string(l.ch)}
	case ')':
		tok = &Token{Type: RPAREN, Literal: string(l.ch)}
	case ',':
		tok = &Token{Type: COMMA, Literal: string(l.ch)}
	case '.':
		tok = &Token{Type: DOT, Literal: string(l.ch)}
	case ':':
		tok = &Token{Type: HAS, Literal: string(l.ch)}
	case '*':
		tok = &Token{Type: ASTERISK, Literal: string(l.ch)}
	case '-':
		tok = &Token{Type: MINUS, Literal: string(l.ch)}
	case '=':
		tok = &Token{Type: EQUALS, Literal: string(l.ch)}
	// potentially two char operators
	case '>':
		literal := string(l.ch)
		if l.peekChar() == '=' {
			l.readChar()
			literal += string(l.ch)
			tok = &Token{Type: GREATER_EQUAL, Literal: literal}
		} else {
			tok = &Token{Type: GREATER_THAN, Literal: literal}
		}
	case '<':
		literal := string(l.ch)
		if l.peekChar() == '=' {
			l.readChar()
			literal += string(l.ch)
			tok = &Token{Type: LESS_EQUAL, Literal: literal}
		} else {
			tok = &Token{Type: LESS_THAN, Literal: literal}
		}
	case '!':
		literal := string(l.ch)
		if l.peekChar() == '=' {
			l.readChar()
			literal += string(l.ch)
			tok = &Token{Type: NOT_EQUALS, Literal: literal}
		} else {
			return Token{Type: ILLEGAL, Literal: ""}
		}
	}
	if tok != nil {
		l.readChar()
		return *tok
	}

	// first char is letter then we have an IDENTIFIER
	if isLetter(l.ch) {
		identifier := l.readIdentifier()
		return Token{Type: LookupIdentifier(identifier), Literal: identifier}
	}

	// first char is digit we have a NUBMER
	if isDigit(l.ch) {
		return Token{Type: NUMBER, Literal: l.readNumber()}
	}

	// first char is a ' or " we have a STRING
	if l.ch == '\'' {
		return Token{Type: STRING, Literal: l.readString('\'')}
	}
	if l.ch == '"' {
		return Token{Type: STRING, Literal: l.readString('"')}
	}

	return Token{Type: ILLEGAL, Literal: ""}
}

// skipWhitespace skips over whitespace characters (space, tab, newline, carriage return)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier or keyword
// Identifiers start with a letter or underscore
// and can contain letters, digits, and underscores
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()

	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float)
// TODO: Implement this method
// Numbers can be:
// - Integers: 42
// - Floats: 3.14
// - Scientific notation: 2.997e9, 1e-10
func (l *Lexer) readNumber() string {
	position := l.position
	// Read digits
	for isDigit(l.ch) {
		l.readChar()
	}
	// try to read decimal
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	// try to read exponent
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

// readString reads a string literal delimited by quotes
// Strings can be delimited by either " or '
// The delimiter parameter indicates which quote started the string
func (l *Lexer) readString(delimiter byte) string {
	position := l.position + 1

	for {
		l.readChar()
		if l.ch == delimiter || l.ch == 0 {
			break
		}
	}
	res := l.input[position:l.position]
	// Advanced past closign quote
	l.readChar()
	return res
}

// isLetter checks if a byte is a letter or underscore
// Helper function for readIdentifier
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if a byte is a digit
// Helper function for readNumber
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
