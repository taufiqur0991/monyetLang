package main

import (
	"fmt"
	"os"

	"monyet/internal/monyet"
)

func main() {
	src, _ := os.ReadFile(os.Args[1])

	lexer := monyet.NewLexer(string(src))
	parser := monyet.NewParser(lexer)
	prog := parser.Parse()

	env := monyet.NewEnv()
	fmt.Printf("DEBUG: Berhasil parse %d statement\n", len(prog.Statements))
	monyet.Eval(prog, env)
}
