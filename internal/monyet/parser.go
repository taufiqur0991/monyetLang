package monyet

import (
	"fmt"
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
	fmt.Printf("Parsing statement, token saat ini: %s (%s)\n", p.cur.Type, p.cur.Value)
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
		p.next() // makan $
		name := p.cur.Value
		p.next() // makan nama identitas

		// 1. Inisialisasi node awal sebagai variabel
		var node Node = Variable{Name: name}

		// 2. Cek apakah ada bracket (IndexAccess)
		// Kita pakai FOR supaya bisa handle nested: $a["b"]["c"]
		for p.cur.Type == LBRACKET {
			p.next() // makan [
			index := p.parseExpr()
			if p.cur.Type != RBRACKET {
				panic("Kurang ] di index access")
			}
			p.next() // makan ]
			node = IndexAccess{Left: node, Index: index}
		}

		// 3. SEKARANG CEK ASSIGNMENT
		if p.cur.Type == ASSIGN {
			p.next() // makan =
			val := p.parseExpr()

			// Jika node adalah Variable biasa
			if v, ok := node.(Variable); ok {
				return Assign{Name: v.Name, Value: val}
			}

			// Jika node adalah IndexAccess, kita butuh tipe AST baru (AssignIndex)
			// Tapi untuk sementara, agar tidak error, kita return node biasa atau handle error
			return Assign{Name: name, Value: val}
		}

		// Jika bukan assignment (misal cuma manggil $_GET["nama"] di baris baru)
		return node
	case IDENT:
		name := p.cur.Value
		p.next()

		var node Node = Variable{Name: name}

		// --- TAMBAHKAN INI UNTUK MENANGANI _GET["nama"] ---
		for p.cur.Type == LBRACKET {
			p.next() // makan [
			index := p.parseExpr()
			if p.cur.Type != RBRACKET {
				panic("Expected ]")
			}
			p.next() // makan ]
			node = IndexAccess{Left: node, Index: index}
		}
		// ------------------------------------------------

		// Cek jika ini fungsi call: hello()
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

		// Cek jika ini assignment ke variabel tanpa $: _GET["a"] = 1
		if p.cur.Type == ASSIGN {
			p.next()
			return Assign{Name: name, Value: p.parseExpr()}
		}

		return node
	case SERVE:
		return p.parseServe()
	default:
		// Jika parser bingung, dia akan lapor
		if p.cur.Type != EOF && p.cur.Type != SEMICOLON && p.cur.Type != RBRACE {
			fmt.Printf("DEBUG: Token = %s, Value = %s\n", p.cur.Type, p.cur.Value)
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
	var node Node

	// 1. Ambil Node Dasar
	if tok.Type == DOLLAR {
		p.next()
		name := p.cur.Value
		p.next()
		node = Variable{Name: name}
	} else if tok.Type == NUMBER {
		p.next()
		v, _ := strconv.Atoi(tok.Value)
		node = Number{Value: v}
	} else if tok.Type == STRING {
		p.next()
		node = String{Value: tok.Value}
	} else if tok.Type == IDENT {
		name := tok.Value
		p.next()
		if p.cur.Type == LPAREN {
			p.next() // (
			args := []Node{}
			for p.cur.Type != RPAREN {
				args = append(args, p.parseExpr())
				if p.cur.Type == COMMA {
					p.next()
				}
			}
			p.next() // )
			node = Call{Name: name, Args: args}
		} else {
			node = Variable{Name: name}
		}
	}

	for p.cur.Type == LBRACKET {
		p.next() // makan [
		index := p.parseExpr()
		if p.cur.Type != RBRACKET {
			panic("Expected ]")
		}
		p.next() // makan ]
		node = IndexAccess{Left: node, Index: index}
	}

	return node
}

func (p *Parser) parseExpr() Node {
	return p.parseLogical()
}

func (p *Parser) parseComparison() Node {
	left := p.parseAdditive()
	for p.cur.Type == GT || p.cur.Type == EQ {
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
	if p.cur.Type == RBRACE {
		p.next() // makan }
	}

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
			if p.cur.Type == RBRACE {
				p.next() // makan }
			}
		} else if p.cur.Type == IF {
			// Support 'else if' secara rekursif
			elseBody = append(elseBody, p.parseIf())
		}
	}

	return If{Condition: cond, Then: thenBody, Else: elseBody}
}

func (p *Parser) parseServe() Node {
	p.next() // makan 'serve'
	if p.cur.Type != LPAREN {
		panic("Kurang ( di serve")
	}
	p.next()

	port := p.parseExpr()

	if p.cur.Type != COMMA {
		panic("Kurang koma setelah port di serve")
	}
	p.next()

	handlerName := p.cur.Value // Ambil nama fungsi
	p.next()

	if p.cur.Type != RPAREN {
		panic("Kurang ) di serve")
	}
	p.next()

	return Serve{Port: port, Handler: handlerName}
}

func (p *Parser) parseLogical() Node {
	left := p.parseComparison() // Logical membungkus comparison

	for p.cur.Type == AND || p.cur.Type == OR {
		op := p.cur.Type
		p.next()
		right := p.parseComparison()
		left = Binary{Left: left, Op: op, Right: right}
	}
	return left
}
