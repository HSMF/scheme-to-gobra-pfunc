package main

import (
	"fmt"
	"io"
	"os"

	"github.com/HSMF/scheme-to-gobra-pfunc/compiler"
	"github.com/HSMF/scheme-to-gobra-pfunc/parser"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	items, err := parser.Parse(string(input))
	if err != nil {
		panic(err)
	}

	pfuncs := compiler.Compile(items)

	// for _, item := range items {
	// 	fmt.Println(item)
	// 	fmt.Println()
	// }

	o, err := os.Create("res.gobra")
	if err != nil {
		panic(err)
	}

	fmt.Fprintln(o, "package main")
	fmt.Fprintln(o)
	fmt.Fprintln(o, "//+gobra")
	fmt.Fprintln(o)

	for _, pfunc := range pfuncs {
		fmt.Println(pfunc)
		fmt.Println()

		fmt.Fprintln(o, pfunc)
	}

	o.Close()

}
