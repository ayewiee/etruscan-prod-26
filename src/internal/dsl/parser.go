package dsl

import (
	"etruscan/internal/domain/models"
	"fmt"
	"strconv"
)

type Parser struct {
	l      *Lexer
	cur    Token
	peek   Token
	errors []*models.DSLError
	raw    string
}

func NewParser(input string) *Parser {
	l := NewLexer(input)
	p := &Parser{
		l:      l,
		raw:    input,
		errors: []*models.DSLError{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.cur = p.peek
	p.peek = p.l.NextToken()
}

func (p *Parser) Errors() []*models.DSLError {
	return p.errors
}

func (p *Parser) addError(code models.ErrorCode, msg string, pos int) {
	start := pos - 2
	if start < 0 {
		start = 0
	}

	end := pos + 3
	if end > len(p.raw) {
		end = len(p.raw)
	}

	near := ""
	if start < len(p.raw) {
		near = p.raw[start:end]
	}

	p.errors = append(p.errors, &models.DSLError{
		Code:     code,
		Message:  msg,
		Position: &pos,
		Near:     &near,
	})
}

// Parse is the main entry point
func (p *Parser) Parse() (Node, []*models.DSLError) {
	expr := p.parseExpression()

	// if we haven't reached EOF and haven't crashed yet, it's a trailing token error
	if p.cur.Type != TokenEOF && len(p.errors) == 0 {
		p.addError(models.ErrCodeDSLParseError, "Неожиданный токен в конце выражения", p.cur.Pos)
	}

	if len(p.errors) > 0 {
		return nil, p.errors
	}

	return expr, nil
}

// EBNF: expression = term { "OR" term }
func (p *Parser) parseExpression() Node {
	left := p.parseTerm()

	for p.cur.Type == TokenOr {
		p.nextToken() // consume OR
		right := p.parseTerm()
		left = &BinaryExpression{Left: left, Operator: "OR", Right: right}
	}
	return left
}

// EBNF: term = factor { "AND" factor }
func (p *Parser) parseTerm() Node {
	left := p.parseFactor()

	for p.cur.Type == TokenAnd {
		p.nextToken() // consume AND
		right := p.parseFactor()
		left = &BinaryExpression{Left: left, Operator: "AND", Right: right}
	}
	return left
}

// EBNF: factor = "NOT" factor | comparison | "(" expression ")"
func (p *Parser) parseFactor() Node {
	if p.cur.Type == TokenNot {
		p.nextToken()
		right := p.parseFactor()
		return &UnaryExpression{Right: right}
	}

	if p.cur.Type == TokenLParen {
		p.nextToken()
		// recursively parse expression. the parsing logic inherently handles
		// simplification of ((A)) because it just returns the inner node
		expr := p.parseExpression()
		if p.cur.Type != TokenRParen {
			p.addError(models.ErrCodeDSLParseError, "Ожидалась закрывающая скобка ')'", p.cur.Pos)
			return expr
		}
		p.nextToken()
		return expr
	}

	// it must be a comparison (Field Op Value)
	return p.parseComparison()
}

func (p *Parser) parseComparison() Node {
	// expect field
	if p.cur.Type != TokenIdent {
		p.addError(
			models.ErrCodeDSLParseError,
			fmt.Sprintf("Ожидалось название поля, получено '%s'", p.cur.Literal),
			p.cur.Pos,
		)
		// return dummy to prevent panic
		return &ComparisonExpression{Field: "?", Operator: "=", Value: 0.0}
	}

	field := p.cur.Literal
	p.nextToken()

	// expect operator: standard comparison, IN, or NOT IN
	var op string
	switch p.cur.Type {
	case TokenOp:
		op = p.cur.Literal
		p.nextToken()
	case TokenIn:
		op = "IN"
		p.nextToken()
	case TokenNot:
		// support "NOT IN" as a combined operator
		if p.peek.Type == TokenIn {
			op = "NOT IN"
			p.nextToken() // consume NOT
			p.nextToken() // consume IN
		} else {
			p.addError(models.ErrCodeDSLParseError, "Ожидался оператор (e.g. >, =, !=, IN, NOT IN)", p.cur.Pos)
			return &ComparisonExpression{Field: field, Operator: "?", Value: 0.0}
		}
	default:
		p.addError(models.ErrCodeDSLParseError, "Ожидался оператор (e.g. >, =, !=, IN, NOT IN)", p.cur.Pos)
		return &ComparisonExpression{Field: field, Operator: "?", Value: 0.0}
	}

	// expect value
	var val interface{}

	// IN / NOT IN expect a list of values in parentheses: (v1, v2, ...)
	if op == "IN" || op == "NOT IN" {
		if p.cur.Type != TokenLParen {
			p.addError(models.ErrCodeDSLParseError, fmt.Sprintf("Ожидался '(' после оператора '%s'", op), p.cur.Pos)
			return &ComparisonExpression{Field: field, Operator: op, Value: nil}
		}
		p.nextToken() // consume '('

		var list []interface{}

		// at least one value
		for {
			if p.cur.Type == TokenNumber {
				f, _ := strconv.ParseFloat(p.cur.Literal, 64)
				list = append(list, f)
				p.nextToken()
			} else if p.cur.Type == TokenString {
				list = append(list, p.cur.Literal)
				p.nextToken()
			} else {
				p.addError(models.ErrCodeDSLParseError, fmt.Sprintf("Ожидалось значение в списке после '%s'", op), p.cur.Pos)
				break
			}

			if p.cur.Type == TokenComma {
				p.nextToken()
				continue
			}
			break
		}

		if p.cur.Type != TokenRParen {
			p.addError(models.ErrCodeDSLParseError, "Ожидалась закрывающая скобка ')' в списке значений", p.cur.Pos)
		} else {
			p.nextToken() // consume ')'
		}

		val = list
	} else {
		if p.cur.Type == TokenNumber {
			f, _ := strconv.ParseFloat(p.cur.Literal, 64)
			val = f
			p.nextToken()
		} else if p.cur.Type == TokenString {
			val = p.cur.Literal
			p.nextToken()
		} else {
			p.addError(models.ErrCodeDSLParseError, fmt.Sprintf("Ожидалось значение после '%s'", op), p.cur.Pos)
			val = 0.0
		}
	}

	return &ComparisonExpression{
		Field:    field,
		Operator: op,
		Value:    val,
	}
}
