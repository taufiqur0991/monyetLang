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
	LBRACE = "{"
	RBRACE = "}"
	COMMA  = ","
	IF     = "IF"
	ELSE   = "ELSE"

	FUNCTION = "FUNCTION"
	RETURN   = "RETURN"
	SERVE    = "SERVE"
)

type Token struct {
	Type  TokenType
	Value string
}

var keywords = map[string]TokenType{
	"echo":     ECHO,
	"function": FUNCTION,
	"return":   RETURN,
}
