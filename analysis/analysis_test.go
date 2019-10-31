package analysis

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/karrick/godirwalk"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

type ParsingContext struct {
	db        *badger.DB
	documents []*Document
}

func newParsingContext() *ParsingContext {
	fmt.Println("Hello???")
	db, _ := badger.Open(badger.DefaultOptions("./testData"))
	return &ParsingContext{
		db:        db,
		documents: []*Document{},
	}
}

func (s *ParsingContext) addDocument(document *Document) {
	s.documents = append(s.documents, document)
	writeDocument(s.db, document)
}

func (s *ParsingContext) close() {
	s.db.Close()
}

func BenchmarkAnalysis(t *testing.B) {
	dir, _ := filepath.Abs("../../go-phpparser/cases/moodle")
	jobs := make(chan string)
	numOfWorkers := 4
	context := newParsingContext()
	defer context.close()

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

// func TestReadData(t *testing.T) {
// 	context := newParsingContext()
// 	defer context.close()
// 	context.db.View(func(txn *badger.Txn) error {
// 		it := txn.NewIterator(badger.DefaultIteratorOptions)
// 		defer it.Close()
// 		for it.Rewind(); it.Valid(); it.Next() {
// 			item := it.Item()
// 			fmt.Println(string(item.Key()))
// 		}
// 		return nil
// 	})
// }

func analyse(context *ParsingContext, id int, filePaths <-chan string) {
	for filePath := range filePaths {
		data, _ := ioutil.ReadFile(filePath)
		text := string(data)
		rootNode := parser.Parse(text)
		context.addDocument(newDocument(util.PathToUri(filePath), text, rootNode))
	}
}
