package lexer

// Lexer performs lexical analysis on AIP-160 filter strings
type Lexer struct {
	input        string // the input string being tokenized
	position     int    // current position in input (points to current char)
	readPosition int    // current reading position in input (after current char)
	ch           byte   // current char under examination
}

// New creates a new Lexer instance for the given input string
// TODO: Implement this function
// Initialize the lexer and read the first character
func New(input string) *Lexer {
	// Your implementation here
	return nil
}

// readChar reads the next character and advances the position in the input
// TODO: Implement this method
// This should:
// - Set l.ch to the character at l.readPosition
// - Handle EOF by setting l.ch to 0
// - Advance both position and readPosition
func (l *Lexer) readChar() {
	// Your implementation here
}

// peekChar looks at the next character without advancing the position
// TODO: Implement this method
// Useful for two-character operators like >= or !=
func (l *Lexer) peekChar() byte {
	// Your implementation here
	return 0
}

// NextToken returns the next token from the input
// TODO: Implement this method
// This is the main method of the lexer!
// It should:
// 1. Skip whitespace
// 2. Determine the token type based on the current character
// 3. Read the full token (may be multiple characters)
// 4. Return the Token
func (l *Lexer) NextToken() Token {
	// Your implementation here
	return Token{Type: ILLEGAL, Literal: ""}
}

// skipWhitespace skips over whitespace characters (space, tab, newline, carriage return)
// TODO: Implement this method
func (l *Lexer) skipWhitespace() {
	// Your implementation here
}

// readIdentifier reads an identifier or keyword
// TODO: Implement this method
// Identifiers start with a letter or underscore
// and can contain letters, digits, and underscores
func (l *Lexer) readIdentifier() string {
	// Your implementation here
	return ""
}

// readNumber reads a number (integer or float)
// TODO: Implement this method
// Numbers can be:
// - Integers: 42
// - Floats: 3.14
// - Scientific notation: 2.997e9, 1e-10
func (l *Lexer) readNumber() string {
	// Your implementation here
	return ""
}

// readString reads a string literal delimited by quotes
// TODO: Implement this method
// Strings can be delimited by either " or '
// The delimiter parameter indicates which quote started the string
func (l *Lexer) readString(delimiter byte) string {
	// Your implementation here
	return ""
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
