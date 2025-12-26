package main

import (
	"os"

	"monyet/internal/monyet"
)

func main() {
	src, _ := os.ReadFile(os.Args[1])

	lexer := monyet.NewLexer(string(src))
	parser := monyet.NewParser(lexer)
	prog := parser.Parse()

	env := monyet.NewEnv()
	monyet.Eval(prog, env)
}
