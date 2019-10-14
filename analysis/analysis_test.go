package analysis

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func BenchmarkAnalysis(t *testing.B) {
	dir := "../../go-phpparser/cases/moodle"
	jobs := make(chan string)
	numOfWorkers := 4

	for i := 0; i < numOfWorkers; i++ {
		go analyse(i, jobs)
	}

	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if !f.IsDir() && strings.HasSuffix(path, ".php") {
			jobs <- path
		}
		return nil
	})
}

func analyse(id int, filePaths <-chan string) {
	for filePath := range filePaths {
		data, _ := ioutil.ReadFile(filePath)
		text := string(data)
		rootNode := parser.Parse(text)
		newDocument(util.PathToUri(filePath), text, rootNode)
	}
}
