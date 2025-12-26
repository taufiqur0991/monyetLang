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

	// echo expr;
	if p.cur.Type == ECHO {
		p.next() // makan ECHO
		expr := p.parseExpr()
		return Echo{Value: expr}
	}

	// assignment: $a = expr;
	if p.cur.Type == IDENT {
		name := p.cur.Value
		p.next()

		if p.cur.Type == ASSIGN {
			p.next()
			val := p.parseExpr()
			return Assign{Name: name, Value: val}
		}
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
