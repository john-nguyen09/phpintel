package analysis

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/karrick/godirwalk"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

type ParsingContext struct {
	documents []*Document
}

func (s *ParsingContext) addDocument(document *Document) {
	s.documents = append(s.documents, document)
}

func BenchmarkAnalysis(t *testing.B) {
	dir := "../../go-phpparser/cases/moodle"
	jobs := make(chan string)
	numOfWorkers := 4
	context := &ParsingContext{}

	for i := 0; i < numOfWorkers; i++ {
		go analyse(context, i, jobs)
	}

	godirwalk.Walk(dir, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.ModeType().IsDir() && strings.HasSuffix(path, ".php") {
				jobs <- path
			}
			return nil
		},
		Unsorted: true,
	})
}

func analyse(context *ParsingContext, id int, filePaths <-chan string) {
	for filePath := range filePaths {
		data, _ := ioutil.ReadFile(filePath)
		text := string(data)
		rootNode := parser.Parse(text)
		context.addDocument(newDocument(util.PathToUri(filePath), text, rootNode))
	}
}
