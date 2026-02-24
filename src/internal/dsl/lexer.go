package dsl

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenNumber
	TokenAnd
	TokenOr
	TokenNot
	TokenIn
	TokenLParen
	TokenRParen
	TokenOp // >, >=, <, <=, =, !=
	TokenComma
)

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

type Lexer struct {
	input      []rune
	ch         rune
	pos        int
	readPos    int
	fullString string
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: []rune(input)}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.fullString += string(l.ch)
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Pos = l.pos

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	case '(':
		tok.Type = TokenLParen
		tok.Literal = "("
		l.readChar()
	case ')':
		tok.Type = TokenRParen
		tok.Literal = ")"
		l.readChar()
	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
		l.readChar()
	case '\'':
		tok.Type = TokenString
		tok.Literal = l.readString()
	case '>', '<', '=', '!':
		tok.Type = TokenOp
		tok.Literal = l.readOperator()
	default:
		if isDigit(l.ch) {
			tok.Type = TokenNumber
			tok.Literal = l.readNumber()
		} else if isLetter(l.ch) {
			lit := l.readIdentifier()
			upper := strings.ToUpper(lit)
			switch upper {
			case "AND":
				tok.Type = TokenAnd
				tok.Literal = "AND"
			case "OR":
				tok.Type = TokenOr
				tok.Literal = "OR"
			case "NOT":
				tok.Type = TokenNot
				tok.Literal = "NOT"
			case "IN":
				tok.Type = TokenIn
				tok.Literal = "IN"
			default:
				tok.Type = TokenIdent
				tok.Literal = lit
			}
		} else {
			// unknown character
			tok.Type = TokenIdent // treat as garbage ident or error later
			tok.Literal = string(l.ch)
			l.readChar()
		}
	}
	return tok
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	// skip opening quote
	l.readChar()
	start := l.pos
	for l.ch != '\'' && l.ch != 0 {
		l.readChar()
	}
	// extract content
	out := string(l.input[start:l.pos])
	// skip closing quote
	if l.ch == '\'' {
		l.readChar()
	}
	return out
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '.' {
		l.readChar()
	}
	return string(l.input[start:l.pos])
}

func (l *Lexer) readOperator() string {
	start := l.pos
	if l.ch == '>' || l.ch == '<' || l.ch == '!' {
		if l.peekChar() == '=' {
			l.readChar()
			l.readChar()
			return string(l.input[start:l.pos])
		}
	}
	l.readChar()
	return string(l.input[start:l.pos])
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}
