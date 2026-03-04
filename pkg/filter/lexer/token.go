package lexer

import "strings"

// TokenType represents the type of token
type TokenType string

// Token represents a lexical token with its type and literal value
type Token struct {
	Type    TokenType
	Literal string
}

// Token type constants
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL" // Unknown token
	EOF     TokenType = "EOF"     // End of file
	
	// Identifiers and literals
	IDENTIFIER TokenType = "IDENTIFIER" // field names: name, age, user_id
	STRING     TokenType = "STRING"     // "hello", 'world'
	NUMBER     TokenType = "NUMBER"     // 42, 3.14, 2.997e9
	
	// Operators
	EQUALS          TokenType = "="  // =
	NOT_EQUALS      TokenType = "!=" // !=
	LESS_THAN       TokenType = "<"  // <
	LESS_EQUAL      TokenType = "<=" // <=
	GREATER_THAN    TokenType = ">"  // >
	GREATER_EQUAL   TokenType = ">=" // >=
	HAS             TokenType = ":"  // : (has operator)
	
	// Logical operators
	AND TokenType = "AND" // AND
	OR  TokenType = "OR"  // OR
	NOT TokenType = "NOT" // NOT
	
	// Delimiters
	DOT        TokenType = "."  // . (traversal)
	LPAREN     TokenType = "("  // (
	RPAREN     TokenType = ")"  // )
	ASTERISK   TokenType = "*"  // * (wildcard)
	MINUS      TokenType = "-"  // - (negation operator)
	
	// Keywords/Literals
	TRUE  TokenType = "TRUE"  // true
	FALSE TokenType = "FALSE" // false
	NULL  TokenType = "NULL"  // null
)

// keywords maps keyword strings to their token types
// Note: AIP-160 keywords are case-insensitive
var keywords = map[string]TokenType{
	"AND":   AND,
	"OR":    OR,
	"NOT":   NOT,
	"TRUE":  TRUE,
	"FALSE": FALSE,
	"NULL":  NULL,
}

// LookupIdentifier checks if an identifier is a keyword
// Returns the keyword's TokenType if it is, otherwise returns IDENTIFIER
func LookupIdentifier(ident string) TokenType {
	// Keywords are case-insensitive in AIP-160
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return IDENTIFIER
}

// String returns a string representation of the token (for debugging)
func (t Token) String() string {
	return string(t.Type) + "(" + t.Literal + ")"
}
