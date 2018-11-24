package script

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func PrintAst(args []string) {
	if len(args) != 2 {
		panic(errors.New("Usage go run printAst.go [filePath]"))
	}

	filePath := args[1]

	fileStat, err := os.Stat(filePath)
	if err != nil {
		panic(err)
	}

	if fileStat.IsDir() {
		panic(errors.New("The path is a directory"))
	}

	byteBuffer, err := ioutil.ReadFile(filePath)

	if err != nil {
		panic(err)
	}

	ast := parser.Parse(string(byteBuffer))
	jsonByte, err := json.MarshalIndent(ast, "", "  ")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonByte))
}
