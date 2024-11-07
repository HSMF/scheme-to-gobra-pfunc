package compiler

import (
	"fmt"
	"strings"
)

type Arg struct {
	name string
	typ  string
}

func (a Arg) String() string {
	return a.name + " " + a.typ
}

type PureFunc struct {
	name   string
	retTyp string
	args   []Arg
	ret    Expr
	req    []Expr
	ens    []Expr
	pres   []Expr
}

func (e PureFunc) String() string {
	args := make([]string, len(e.args))

	for i, arg := range e.args {
		args[i] = arg.String()
	}

	res := ""
	fmt.Printf("e.req: %v\n", e.req)
	for _, r := range e.req {
		res += fmt.Sprintf("requires %s\n", r.String())
	}
	for _, r := range e.pres {
		res += fmt.Sprintf("preserves %s\n", r.String())
	}
	for _, r := range e.ens {
		res += fmt.Sprintf("ensures %s\n", r.String())
	}
	res += "decreases _\n"

	return res + fmt.Sprintf("pure func %s(%s) %s {\n\treturn %s\n}", e.name, strings.Join(args, ", "), e.retTyp, e.ret.String())
}

type Expr interface{ String() string }

type Atom struct {
	src string
}

func (a Atom) String() string {
	return a.src
}

type UnaryOp struct {
	e  Expr
	op string
}

func (e UnaryOp) String() string {
	return fmt.Sprintf("(%s %s)", e.op, e.e.String())
}

type BinOp struct {
	left  Expr
	right Expr
	op    string
}

func (e BinOp) String() string {
	return fmt.Sprintf("(%s %s %s)", e.left.String(), e.op, e.right.String())
}

type TernaryCond struct {
	cond      Expr
	then      Expr
	otherwise Expr
}

func (e TernaryCond) String() string {
	return fmt.Sprintf("(%s ? %s : %s)", e.cond.String(), e.then.String(), e.otherwise.String())
}

type SeqSlice struct {
	seq   Expr
	start *Expr
	end   *Expr
}

func (e SeqSlice) String() string {
	end := ""
	start := ""
	if e.start != nil {
		start = (*e.start).String()
	}
	if e.end != nil {
		end = (*e.end).String()
	}
	return fmt.Sprintf("(%s[%s:%s])", e.seq.String(), start, end)
}

type SeqLiteral struct {
	typ   string
	elems []Expr
}

func (e SeqLiteral) String() string {
	res := fmt.Sprintf("seq[%s] {", e.typ)

	if len(e.elems) == 0 {
		return res + "}"
	}

	res += e.elems[0].String()

	for _, elem := range e.elems[1:] {
		res += ", "
		res += elem.String()
	}

	return res + "}"
}

type Call struct {
	fn   Expr
	args []Expr
}

func (e Call) String() string {
	res := fmt.Sprintf("%s (", e.fn.String())

	if len(e.args) == 0 {
		return res + ")"
	}

	res += e.args[0].String()

	for _, elem := range e.args[1:] {
		res += ", "
		res += elem.String()
	}

	return res + ")"
}

func Len(e Expr) Expr {
	return Call{Atom{"len"}, []Expr{e}}
}
