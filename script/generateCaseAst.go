package script

import (
	"encoding/json"
	"io/ioutil"
	"path"
	"strings"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func GenerateCaseAst(args []string) {
	generateAstForDir("case")
}

func generateAstForDir(dir string) {
	fileInfos, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	for _, fileInfo := range fileInfos {
		filePath := path.Join(dir, fileInfo.Name())

		if fileInfo.IsDir() {
			generateAstForDir(filePath)

			return
		}

		if strings.HasSuffix(filePath, ".php") {
			buffer, err := ioutil.ReadFile(filePath)

			if err != nil {
				panic(err)
			}

			ast := parser.Parse(string(buffer))
			jsonBuffer, err := json.MarshalIndent(ast, "", "  ")

			if err != nil {
				panic(err)
			}

			ioutil.WriteFile(filePath+".json", jsonBuffer, 0644)
		}
	}
}
