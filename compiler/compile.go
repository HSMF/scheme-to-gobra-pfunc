package compiler

import (
	"fmt"
	"strings"

	"github.com/HSMF/scheme-to-gobra-pfunc/parser"
)

var compileExprHandlers map[string]compileExprHandler

func init() {
	compileExprHandlers = map[string]compileExprHandler{
		"+":    compileBinop("+"),
		"-":    compileBinop("-"),
		"*":    compileBinop("*"),
		"/":    compileBinop("/"),
		"=":    compileBinop("=="),
		"=seq": compileBinop("=="),
		"++":   compileBinop("++"),
		">=":   compileBinop(">="),
		">":    compileBinop(">"),
		"<":    compileBinop("<"),
		"<=":   compileBinop("<="),
		"&&":   compileBinop("&&"),
		"||":   compileBinop("||"),
		"if":   compileTernary,
		"seq":  compileSeq,
		"null?": func(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {

			pfuncs, compiled := compileExpr(n.Children[1], ctx)

			return pfuncs, BinOp{Len(compiled), Atom{"0"}, "=="}
		},
		"slice": func(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {
			pfuncs, compiledSeq := compileExpr(n.Children[1], ctx)
			compiledLow := compileAppendPureFuncs(compileExpr, n.Children[2], ctx, &pfuncs)
			compiledHigh := compileAppendPureFuncs(compileExpr, n.Children[3], ctx, &pfuncs)
			return pfuncs, SeqSlice{compiledSeq, &compiledLow, &compiledHigh}
		},
		"letrec": compileInnerFunction,
		"cond": func(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {
			panic("todo: conditionals")
		},
	}
}

func Compile(items []parser.AstNode) []PureFunc {
	res := make([]PureFunc, 0, len(items))
	// fmt.Printf("compiling items: %v\n", items)

	for _, item := range items {
		res = append(res, compileItem(item)...)
	}

	return res
}

type handlerFunc = func(item parser.Node) []PureFunc
type compileExprHandler func(parser.Node, *ctx) ([]PureFunc, Expr)

type ctx struct {
	args map[string]string
}

func getChildrenByLabel(it parser.Node, label string) []parser.AstNode {
	res := make([]parser.AstNode, 0)
	for _, i := range it.Children {
		if i.Label() == label {
			res = append(res, i)
		}
	}

	return res
}

func getChildrenByFunc(it parser.Node, label func(string) bool) []parser.AstNode {
	res := make([]parser.AstNode, 0)
	for _, i := range it.Children {
		if label(i.Label()) {
			res = append(res, i)
		}
	}

	return res
}

func stripString(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "\""), "\"")
}

func compileAppendPureFuncs[A any, B any](compile func(A, *ctx) ([]PureFunc, B), n A, ctx *ctx, pf *[]PureFunc) B {
	more, compiled := compile(n, ctx)
	*pf = append(*pf, more...)
	return compiled
}

func compileBinop(op string) compileExprHandler {
	return func(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {
		arg1 := n.Children[1]
		arg2 := n.Children[2]

		funcs, compiledArg1 := compileExpr(arg1, ctx)
		more, compiledArg2 := compileExpr(arg2, ctx)

		return append(funcs, more...),
			BinOp{compiledArg1, compiledArg2, op}
	}
}

func compileTernary(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {
	cond := n.Children[1]
	then := n.Children[2]
	otherwise := n.Children[3]

	funcs := make([]PureFunc, 0)

	compiledCond := compileAppendPureFuncs(compileExpr, cond, ctx, &funcs)
	compiledThen := compileAppendPureFuncs(compileExpr, then, ctx, &funcs)
	compiledOtherwise := compileAppendPureFuncs(compileExpr, otherwise, ctx, &funcs)
	return funcs, TernaryCond{compiledCond, compiledThen, compiledOtherwise}
}

func compileSeq(n parser.Node, ctx *ctx) ([]PureFunc, Expr) {
	typ := stripString(n.Children[1].String())
	elems := n.Children[2:]

	pfuncs := make([]PureFunc, 0)

	compiledElems := make([]Expr, len(elems))
	for i, arg := range elems {
		var more []PureFunc
		more, compiledElems[i] = compileExpr(arg, ctx)
		pfuncs = append(pfuncs, more...)
	}

	return pfuncs, SeqLiteral{typ, compiledElems}
}

func compileInnerFunction(item parser.Node, ctx *ctx) ([]PureFunc, Expr) {

	// letrec is only supported like this:
	// (letrec ((name (lambda (...args) (begin ...) ) )) expr )

	panic("todo: inner functions")
}

func compileExpr(item parser.AstNode, ctx *ctx) ([]PureFunc, Expr) {

	switch it := item.(type) {
	case parser.Atom:
		return nil, Atom{it.String()}
	case parser.Node:

		specialHandler, isSpecial := compileExprHandlers[item.Label()]

		if isSpecial {
			return specialHandler(it, ctx)
		}

		fn := it.Children[0]
		args := it.Children[1:]

		pfuncs, compiledFn := compileExpr(fn, ctx)

		compiledArgs := make([]Expr, len(args))
		for i, arg := range args {
			var more []PureFunc
			more, compiledArgs[i] = compileExpr(arg, ctx)
			pfuncs = append(pfuncs, more...)
		}

		return pfuncs, Call{fn: compiledFn, args: compiledArgs}
	}

	panic("unreachable")
}

func compileSpec(conditions []parser.AstNode, ctx *ctx) ([]PureFunc, []Expr) {
	fmt.Printf("conditions: %v\n", conditions)
	res := make([]Expr, 0)
	pfuncs := make([]PureFunc, 0)

	for _, cond := range conditions {
		condVal := cond.(parser.Node).Children[1]
		pf, compiled := compileExpr(condVal, ctx)
		pfuncs = append(pfuncs, pf...)
		res = append(res, compiled)
	}

	return pfuncs, res
}

func compileDefine(item parser.Node) []PureFunc {
	name := item.Children[1].Label()

	body := getChildrenByLabel(item, "begin")[0].(parser.Node)

	preconditions := getChildrenByLabel(body, "requires")
	postconditions := getChildrenByLabel(body, "ensures")
	preservesconditions := getChildrenByLabel(body, "preserves")
	args := getChildrenByLabel(body, "args")[0].(parser.Node)
	retTyp := getChildrenByLabel(body, "returns")[0].(parser.Node).Children[1].String()

	compiledArgs := make([]Arg, len(args.Children)-1)
	argTypes := make(map[string]string)
	for i, arg := range args.Children[1:] {
		a := arg.(parser.Node)
		name := a.Children[0].String()
		typ := stripString(a.Children[1].String())
		compiledArgs[i] = Arg{name: name, typ: typ}
		argTypes[name] = typ
	}

	exprs := getChildrenByFunc(body, func(label string) bool {
		return label != "args" && label != "returns" && label != "requires" && label != "preserves" && label != "ensures"
	})
	expr := exprs[len(exprs)-1]

	ctx := &ctx{}

	pfuncs, compiledExpr := compileExpr(expr, ctx)

	compiledPre := compileAppendPureFuncs(compileSpec, preconditions, ctx, &pfuncs)
	compiledPres := compileAppendPureFuncs(compileSpec, preservesconditions, ctx, &pfuncs)
	compiledPost := compileAppendPureFuncs(compileSpec, postconditions, ctx, &pfuncs)

	return append(pfuncs, PureFunc{
		name:   name,
		args:   compiledArgs,
		retTyp: stripString(retTyp),
		ret:    compiledExpr,
		req:    compiledPre,
		pres:   compiledPres,
		ens:    compiledPost,
	})
}

var handlers map[string]handlerFunc = map[string]handlerFunc{
	"define": compileDefine,
}

func compileItem(item parser.AstNode) []PureFunc {

	switch it := item.(type) {
	case parser.Atom:
		return nil
	case parser.Node:
		name := it.Label()
		if name == "" {
			// fmt.Printf("skipping %s because it has label\n", item)
			return nil
		}
		handler, ok := handlers[name]
		if !ok {
			// fmt.Printf("skipping %s because it is not handled\n", item)
			return nil
		}

		return handler(it)
	}

	return nil
}
