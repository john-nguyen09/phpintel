package main

import (
	"errors"
	"os"

	"github.com/john-nguyen09/phpintel/script"
)

func main() {
	if len(os.Args) < 2 {
		panic(errors.New("Usage: go run main.go COMMAND [options]"))
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "generateCaseAst":
		script.GenerateCaseAst(args)
	case "printAst":
		script.PrintAst(args)
	}
}
