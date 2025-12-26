package main

import (
	"os"
	"path/filepath"

	"monyet/internal/monyet"
)

func main() {
	src, _ := os.ReadFile(os.Args[1])
	absPath, _ := filepath.Abs(os.Args[1])
	baseDir := filepath.Dir(absPath)

	lexer := monyet.NewLexer(string(src))
	parser := monyet.NewParser(lexer)
	prog := parser.Parse()

	env := monyet.NewEnv()
	env.SetVar("__BASE_DIR__", baseDir)
	// fmt.Printf("DEBUG: Berhasil parse %d statement\n", len(prog.Statements))
	// fmt.Printf("Parsed statements: %d\n", len(prog.Statements))
	monyet.Eval(prog, env)
}
