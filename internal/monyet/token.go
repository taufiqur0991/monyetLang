package monyet

type TokenType string

const (
	RENDER  = "RENDER"
	INCLUDE = "INCLUDE"
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

	ECHO        = "ECHO"
	STRING      = "STRING"
	GT          = ">"
	LBRACE      = "{"
	RBRACE      = "}"
	COMMA       = ","
	IF          = "IF"
	ELSE        = "ELSE"
	EQ          = "=="
	LBRACKET    = "["
	RBRACKET    = "]"
	FUNCTION    = "FUNCTION"
	RETURN      = "RETURN"
	SERVE       = "SERVE"
	AND         = "&&"
	OR          = "||"
	JSON_ENCODE = "JSON_ENCODE"
	JSON_DECODE = "JSON_DECODE"
	ARROW       = "=>"
)

type Token struct {
	Type  TokenType
	Value string
}

// var keywords = map[string]TokenType{
// 	"echo":     ECHO,
// 	"function": FUNCTION,
// 	"return":   RETURN,
// 	"include":  INCLUDE,
// }
