package parser

import (
	"strings"
)

type AstNode interface {
	String() string
	stringPretty(indent int) (string string, wrapsLine bool)

	Label() string
}

type Node struct {
	Children []AstNode
}

func (n Node) Label() string {
	if len(n.Children) == 0 {
		return ""
	}

	switch f := n.Children[0].(type) {
	case Atom:
		return f.ident
	}
	return ""
}

func (n Node) String() string {
	s, _ := n.stringPretty(1)
	return s
}

func (n Node) stringPretty(indent int) (string, bool) {
	if len(n.Children) == 0 {
		return "()", false
	}

	children := make([]string, 0, len(n.Children))
	childWidth := len(n.Children) - 1

	shouldWrap := false
	for _, child := range n.Children {
		childString, wraps := child.stringPretty(indent + 1)
		children = append(children, childString)
		shouldWrap = shouldWrap || wraps
		childWidth += len(childString)
	}

	builder := strings.Builder{}
	separator := func() {
		builder.WriteByte(' ')
	}

	shouldWrap = shouldWrap || childWidth >= 30
	if shouldWrap {
		separator = func() {
			builder.WriteByte('\n')
			for range indent {
				builder.WriteString("  ")
			}
		}
	}

	builder.WriteString("(")
	builder.WriteString(children[0])

	for _, child := range children[1:] {
		separator()
		builder.WriteString(child)
	}
	builder.WriteByte(')')

	return builder.String(), shouldWrap
}

type Atom struct {
	ident string
}

func (a Atom) String() string {
	res, _ := a.stringPretty(0)
	return res
}

func (a Atom) stringPretty(int) (string, bool) {
	return a.ident, false
}

func (a Atom) Label() string {
	return ""
}
