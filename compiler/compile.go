package compiler

import (
	"fmt"
	"strings"

	"github.com/HSMF/scheme-to-gobra-pfunc/parser"
)

func init() {
	compileExprHandlers = map[string]compileExprHandler{
		"+":   compileBinop("+"),
		"-":   compileBinop("-"),
		"*":   compileBinop("*"),
		"/":   compileBinop("/"),
		"=":   compileBinop("=="),
		"++":  compileBinop("++"),
		">=":  compileBinop(">="),
		">":   compileBinop(">"),
		"<":   compileBinop("<"),
		"<=":  compileBinop("<="),
		"if":  compileTernary,
		"seq": compileSeq,
		"slice": func(n parser.Node) ([]PureFunc, Expr) {
			panic("todo")
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

func compileAppendPureFuncs[A any, B any](compile func(A) ([]PureFunc, B), n A, pf *[]PureFunc) B {
	more, compiled := compile(n)
	*pf = append(*pf, more...)
	return compiled
}

type compileExprHandler func(parser.Node) ([]PureFunc, Expr)

func compileBinop(op string) compileExprHandler {
	return func(n parser.Node) ([]PureFunc, Expr) {
		arg1 := n.Children[1]
		arg2 := n.Children[2]

		funcs, compiledArg1 := compileExpr(arg1)
		more, compiledArg2 := compileExpr(arg2)

		return append(funcs, more...),
			BinOp{compiledArg1, compiledArg2, op}
	}
}

func compileTernary(n parser.Node) ([]PureFunc, Expr) {
	cond := n.Children[1]
	then := n.Children[2]
	otherwise := n.Children[3]

	funcs := make([]PureFunc, 0)

	compiledCond := compileAppendPureFuncs(compileExpr, cond, &funcs)
	compiledThen := compileAppendPureFuncs(compileExpr, then, &funcs)
	compiledOtherwise := compileAppendPureFuncs(compileExpr, otherwise, &funcs)
	return funcs, TernaryCond{compiledCond, compiledThen, compiledOtherwise}
}

func compileSeq(n parser.Node) ([]PureFunc, Expr) {
	typ := stripString(n.Children[1].String())
	elems := n.Children[2:]

	pfuncs := make([]PureFunc, 0)

	compiledElems := make([]Expr, len(elems))
	for i, arg := range elems {
		var more []PureFunc
		more, compiledElems[i] = compileExpr(arg)
		pfuncs = append(pfuncs, more...)
	}

	return pfuncs, SeqLiteral{typ, compiledElems}
}

var compileExprHandlers map[string]compileExprHandler

func compileExpr(item parser.AstNode) ([]PureFunc, Expr) {

	switch it := item.(type) {
	case parser.Atom:
		return nil, Atom{it.String()}
	case parser.Node:

		specialHandler, isSpecial := compileExprHandlers[item.Label()]

		if isSpecial {
			return specialHandler(it)
		}

		fn := it.Children[0]
		args := it.Children[1:]

		pfuncs, compiledFn := compileExpr(fn)

		compiledArgs := make([]Expr, len(args))
		for i, arg := range args {
			var more []PureFunc
			more, compiledArgs[i] = compileExpr(arg)
			pfuncs = append(pfuncs, more...)
		}

		return pfuncs, Call{fn: compiledFn, args: compiledArgs}
	}

	panic("unreachable")
}

func compileSpec(conditions []parser.AstNode) ([]PureFunc, []Expr) {
	fmt.Printf("conditions: %v\n", conditions)
	res := make([]Expr, 0)
	pfuncs := make([]PureFunc, 0)

	for _, cond := range conditions {
		condVal := cond.(parser.Node).Children[1]
		pf, compiled := compileExpr(condVal)
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
	for i, arg := range args.Children[1:] {
		a := arg.(parser.Node)
		compiledArgs[i] = Arg{name: a.Children[0].String(), typ: stripString(a.Children[1].String())}
	}

	exprs := getChildrenByFunc(body, func(label string) bool {
		return label != "args" && label != "returns" && label != "requires" && label != "preserves" && label != "ensures"
	})
	expr := exprs[len(exprs)-1]

	pfuncs, compiledExpr := compileExpr(expr)

	compiledPre := compileAppendPureFuncs(compileSpec, preconditions, &pfuncs)
	compiledPres := compileAppendPureFuncs(compileSpec, preservesconditions, &pfuncs)
	compiledPost := compileAppendPureFuncs(compileSpec, postconditions, &pfuncs)

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
