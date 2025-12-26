package monyet

import "unicode"

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(src string) *Lexer {
	return &Lexer{input: []rune(src)}
}

func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	ch := l.input[l.pos]
	l.pos++
	return ch
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) NextToken() Token {
	ch := l.next()

	// skip whitespace
	for unicode.IsSpace(ch) {
		ch = l.next()
	}

	switch ch {
	case 0:
		return Token{Type: EOF}
	case '+':
		return Token{Type: PLUS, Value: "+"}
	case '-':
		return Token{Type: MINUS, Value: "-"}
	case '*':
		return Token{Type: STAR, Value: "*"}
	case '/':
		if l.peek() == '/' {
			// Ini komentar, abaikan sampai akhir baris
			for {
				ch := l.next()
				if ch == '\n' || ch == 0 {
					break
				}
			}
			return l.NextToken() // Panggil lagi untuk cari token asli berikutnya
		}
		return Token{Type: SLASH, Value: "/"}
	case '=':
		if l.peek() == '=' {
			l.next() // makan = kedua
			return Token{Type: EQ, Value: "=="}
		}
		return Token{Type: ASSIGN, Value: "="}
	case ';':
		return Token{Type: SEMICOLON, Value: ";"}
	case '(':
		return Token{Type: LPAREN, Value: "("}
	case ')':
		return Token{Type: RPAREN, Value: ")"}
	case '$':
		return Token{Type: DOLLAR, Value: "$"}
	case '"':
		return Token{Type: STRING, Value: l.readString()}
	case '>':
		return Token{Type: GT, Value: ">"}
	case '{':
		return Token{Type: LBRACE, Value: "{"}
	case '}':
		return Token{Type: RBRACE, Value: "}"}
	case ',':
		return Token{Type: COMMA, Value: ","}
	case '[':
		return Token{Type: LBRACKET, Value: "["}
	case ']':
		return Token{Type: RBRACKET, Value: "]"}

	}

	if unicode.IsDigit(ch) {
		return Token{Type: NUMBER, Value: l.readNumber(ch)}
	}

	if unicode.IsLetter(ch) || ch == '_' {
		ident := l.readIdent(ch)
		if ident == "echo" {
			return Token{Type: ECHO, Value: ident}
		}
		if ident == "fn" || ident == "function" {
			return Token{Type: FUNCTION, Value: ident}
		}
		if ident == "return" {
			return Token{Type: RETURN, Value: ident}
		}
		if ident == "if" {
			return Token{Type: IF, Value: ident}
		}
		if ident == "else" {
			return Token{Type: ELSE, Value: ident}
		}
		if ident == "serve" {
			return Token{Type: SERVE, Value: ident}
		}

		return Token{Type: IDENT, Value: ident}
	}

	return Token{Type: ILLEGAL, Value: string(ch)}
}

func (l *Lexer) readNumber(start rune) string {
	out := []rune{start}
	for unicode.IsDigit(l.peek()) {
		out = append(out, l.next())
	}
	return string(out)
}

func (l *Lexer) readIdent(start rune) string {
	out := []rune{start}
	for {
		ch := l.peek()
		// Izinkan huruf, angka, dan underscore
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			out = append(out, l.next())
		} else {
			break
		}
	}
	return string(out)
}

func (l *Lexer) readString() string {
	out := []rune{}
	for {
		ch := l.next()
		if ch == '"' || ch == 0 {
			break
		}
		out = append(out, ch)
	}
	return string(out)
}
