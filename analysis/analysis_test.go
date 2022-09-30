package analysis

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type ParsingContext struct {
	store     *Store
	waitGroup sync.WaitGroup
}

func withParsingContext(fn func(*ParsingContext)) {
	withTestStore("test", "BenchmarkAnalysis", func(store *Store) {
		fn(&ParsingContext{
			store: store,
		})
	})
}

func (s *ParsingContext) addDocument(document *Document) {
	s.store.SyncDocument(document)
}

func (s *ParsingContext) close() {
	s.store.Close()
}

func BenchmarkAnalysis(t *testing.B) {
	dir, _ := filepath.Abs("../cases/moodle")
	jobs := make(chan string)
	numOfWorkers := 4
	withParsingContext(func(context *ParsingContext) {
		defer context.close()

		for i := 0; i < numOfWorkers; i++ {
			go analyse(context, i, jobs)
		}

		filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
			if err != nil {
				panic(err)
			}

			if !info.IsDir() && strings.HasSuffix(path, ".php") {
				context.waitGroup.Add(1)
				jobs <- path
			}
			return nil
		})
		context.waitGroup.Wait()
	})
}

// func TestReadData(t *testing.T) {
// 	context := newParsingContext()
// 	defer context.close()
// 	it := context.store.db.NewIterator(nil, nil)
// 	defer it.Release()
// 	for it.Next() {
// 		fmt.Println(string(it.Key()))
// 	}
// }

func analyse(context *ParsingContext, id int, filePaths <-chan string) {
	count := 0
	for filePath := range filePaths {
		data, _ := ioutil.ReadFile(filePath)
		document := NewDocument("test"+fmt.Sprint(id)+fmt.Sprint(count), data)
		document.Load()
		context.addDocument(document)
		context.waitGroup.Done()
		count++
	}
}
