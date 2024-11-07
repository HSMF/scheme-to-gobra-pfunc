package parser

import (
	"errors"
	"fmt"
	"strings"
)

type Span struct {
	start_line int
	start_col  int
	end_line   int
	end_col    int
}

type TokenKind int

const (
	OpenParen TokenKind = iota
	CloseParen
	Ident
)

type Token struct {
	src  string
	kind TokenKind
	span Span
}

func NewToken(src string, kind TokenKind, line, col int) Token {
	return Token{src, kind, Span{line, col, line, col + len(src)}}
}

func IsSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func IsNewline(ch byte) bool {
	return ch == '\n'
}

func Tokenize(s string) []Token {
	res := make([]Token, 0)
	line := 1
	col := 0
	for i := 0; i < len(s); i++ {
		col += 1
		if IsSpace(s[i]) {
			continue
		}

		if IsNewline(s[i]) {
			line += 1
			col = 0
			continue
		}

		if s[i] == ';' {
			end := strings.IndexFunc(s[i:], func(r rune) bool {
				return IsNewline(byte(r))
			})
			if end == -1 {
				i = len(s)
			} else {
				i = end
			}
			continue
		}

		if s[i] == '(' {
			res = append(res, NewToken("(", OpenParen, line, col))
			continue
		}
		if s[i] == ')' {
			res = append(res, NewToken(")", CloseParen, line, col))
			continue
		}

		end := strings.IndexFunc(s[i:], func(r rune) bool {
			return IsSpace(byte(r)) || IsNewline(byte(r)) || r == '(' || r == ')' || r == ';'
		})
		if end == -1 {
			res = append(res, NewToken(s[i:], Ident, line, col))
			break
		}
		res = append(res, NewToken(s[i:i+end], Ident, line, col))
		col += end
		i += end - 1
	}

	return res
}

func Parse(s string) ([]AstNode, error) {
	tokens := Tokenize(s)
	fmt.Printf("tokens: %v\n", tokens)

	res := make([]AstNode, 0)

	for len(tokens) > 0 {
		first := tokens[0]
		tokens = tokens[1:]
		switch first.kind {
		case OpenParen:
			node, rest := parseSExpr(tokens)
			if node == nil {
				return nil, errors.New("parse error, mismatched (")
			}
			res = append(res, node)
			tokens = rest
		case CloseParen:
			return nil, errors.New("parse error, mismatched )")
		case Ident:
			res = append(res, Atom{first.src})

		}

	}

	return res, nil
}

func parseSExpr(tokens []Token) (AstNode, []Token) {
	children := make([]AstNode, 0)

	for len(tokens) > 0 {
		first := tokens[0]
		tokens = tokens[1:]
		switch first.kind {
		case OpenParen:
			child, toks := parseSExpr(tokens)
			if child == nil {
				return nil, nil
			}
			tokens = toks
			children = append(children, child)
		case CloseParen:
			return Node{Children: children}, tokens
		case Ident:
			children = append(children, Atom{first.src})
		}
	}

	return nil, nil
}
