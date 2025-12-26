package monyet

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENT  = "IDENT"
	NUMBER = "NUMBER"

	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"
	STAR   = "*"
	SLASH  = "/"

	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"

	DOLLAR = "$"

	ECHO   = "ECHO"
	STRING = "STRING"
	GT     = ">"
)

type Token struct {
	Type  TokenType
	Value string
}
