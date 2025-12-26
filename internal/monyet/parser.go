package monyet

import (
	"strconv"
)

type Parser struct {
	l   *Lexer
	cur Token
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.next()
	return p
}

func (p *Parser) next() {
	p.cur = p.l.NextToken()
}

func (p *Parser) Parse() *Program {
	prog := &Program{}

	for p.cur.Type != EOF {
		// fmt.Printf("TOKEN: %+v\n", p.cur)

		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}

		if p.cur.Type == SEMICOLON {
			p.next()
		} else if stmt == nil {
			p.next()
		}
	}

	return prog
}

func (p *Parser) parseStatement() Node {
	switch p.cur.Type {
	case IF:
		return p.parseIf()
	case FUNCTION:
		return p.parseFunction()
	case RETURN:
		p.next()
		return Return{Value: p.parseExpr()}
	case ECHO:
		p.next()
		return Echo{Value: p.parseExpr()}
	case DOLLAR:
		// Ini pasti Assignment: $a = ...
		p.next() // makan $
		name := p.cur.Value
		p.next() // makan identitas
		if p.cur.Type == ASSIGN {
			p.next() // makan =
			return Assign{Name: name, Value: p.parseExpr()}
		}
		return Variable{Name: name}
	case IDENT:
		// Ini bisa jadi Call: cekAngka(15)
		name := p.cur.Value
		p.next()
		if p.cur.Type == LPAREN {
			p.next() // makan (
			args := []Node{}
			for p.cur.Type != RPAREN && p.cur.Type != EOF {
				args = append(args, p.parseExpr())
				if p.cur.Type == COMMA {
					p.next()
				}
			}
			p.next() // makan )
			return Call{Name: name, Args: args}
		}
		return Variable{Name: name}
	case SERVE:
		p.next() // makan 'serve'
		p.next() // makan '('
		port := p.parseExpr()
		p.next() // makan ','
		handlerName := p.cur.Value
		p.next()
		p.next() // makan ')'
		return Serve{Port: port, Handler: handlerName}
	}

	return nil
}

func (p *Parser) parseTerm() Node {
	left := p.parseFactor()
	for p.cur.Type == STAR || p.cur.Type == SLASH {
		op := p.cur.Type
		p.next()
		right := p.parseFactor()
		left = Binary{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseFactor() Node {
	tok := p.cur

	if tok.Type == DOLLAR {
		p.next()
		name := p.cur.Value
		p.next()
		return Variable{Name: name}
	}

	if tok.Type == NUMBER {
		p.next()
		v, _ := strconv.Atoi(tok.Value)
		return Number{Value: v}
	}

	if tok.Type == STRING {
		p.next()
		return String{Value: tok.Value}
	}

	if tok.Type == IDENT {
		name := tok.Value
		p.next()

		if p.cur.Type == LPAREN {
			p.next()
			args := []Node{}

			for p.cur.Type != RPAREN {
				args = append(args, p.parseExpr())
				if p.cur.Type == COMMA {
					p.next()
				}
			}

			p.next() // )

			return Call{Name: name, Args: args}
		}

		return Variable{Name: name}
	}

	return nil
}

func (p *Parser) parseExpr() Node {
	return p.parseComparison()
}

func (p *Parser) parseComparison() Node {
	left := p.parseAdditive()
	for p.cur.Type == GT {
		op := p.cur.Type
		p.next()
		right := p.parseAdditive()
		left = Binary{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAdditive() Node {
	left := p.parseTerm()
	for p.cur.Type == PLUS || p.cur.Type == MINUS {
		op := p.cur.Type
		p.next()
		right := p.parseTerm()
		left = Binary{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseFunction() Node {
	p.next() // makan 'function' atau 'fn'

	name := p.cur.Value
	p.next() // makan nama fungsi

	if p.cur.Type != LPAREN {
		panic("Expected ( after function name")
	}
	p.next() // makan '('

	params := []string{}
	for p.cur.Type != RPAREN && p.cur.Type != EOF {
		// Jika parameter menggunakan $, sesuaikan di sini
		if p.cur.Type == DOLLAR {
			p.next() // skip $
		}
		params = append(params, p.cur.Value)
		p.next()
		if p.cur.Type == COMMA {
			p.next()
		}
	}

	if p.cur.Type != RPAREN {
		panic("Expected ) after parameters")
	}
	p.next() // makan ')'

	if p.cur.Type != LBRACE {
		panic("Expected { before function body")
	}
	p.next() // makan '{'

	body := []Node{}
	for p.cur.Type != RBRACE && p.cur.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			body = append(body, stmt)
		}
		if p.cur.Type == SEMICOLON {
			p.next()
		}
	}

	if p.cur.Type != RBRACE {
		panic("Expected } after function body")
	}
	p.next() // makan '}'

	return Function{Name: name, Params: params, Body: body}
}

func (p *Parser) parseIf() Node {
	p.next() // makan 'if'

	if p.cur.Type != LPAREN {
		panic("Expected ( after if")
	}
	p.next() // makan '('
	cond := p.parseExpr()
	if p.cur.Type != RPAREN {
		panic("Expected ) after condition")
	}
	p.next() // makan ')'

	if p.cur.Type != LBRACE {
		panic("Expected { after if condition")
	}
	p.next() // makan '{'

	thenBody := []Node{}
	for p.cur.Type != RBRACE && p.cur.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			thenBody = append(thenBody, stmt)
		}
		if p.cur.Type == SEMICOLON {
			p.next()
		}
	}
	p.next() // makan '}' (RBRACE)

	var elseBody []Node
	if p.cur.Type == ELSE {
		p.next() // makan 'else'

		// Cek apakah ada '{' (untuk else) atau langsung 'if' (untuk else if)
		if p.cur.Type == LBRACE {
			p.next() // makan '{'
			for p.cur.Type != RBRACE && p.cur.Type != EOF {
				stmt := p.parseStatement()
				if stmt != nil {
					elseBody = append(elseBody, stmt)
				}
				if p.cur.Type == SEMICOLON {
					p.next()
				}
			}
			p.next() // makan '}'
		} else if p.cur.Type == IF {
			// Support 'else if' secara rekursif
			elseBody = append(elseBody, p.parseIf())
		}
	}

	return If{Condition: cond, Then: thenBody, Else: elseBody}
}
