package lsp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/karrick/godirwalk"
)

const numOfWorkers int = 2

type workspaceStore struct {
	ctx    context.Context
	stores map[string]*analysis.Store
	jobs   chan string
}

func newWorkspaceStore(ctx context.Context) *workspaceStore {
	workspaceStore := &workspaceStore{
		ctx:    ctx,
		stores: map[string]*analysis.Store{},
		jobs:   make(chan string),
	}
	for i := 0; i < numOfWorkers; i++ {
		go workspaceStore.analyse(i)
	}
	return workspaceStore
}

func (s *workspaceStore) analyse(id int) {
	for filePath := range s.jobs {
		s.addDocument(filePath)
	}
}

func (s *workspaceStore) close() {
	for _, store := range s.stores {
		store.Close()
	}
}

func (s *workspaceStore) addView(uri protocol.DocumentURI) {
	h := md5.New()
	io.WriteString(h, uri)
	hash := hex.EncodeToString(h.Sum(nil))
	storagePath := filepath.Join(getDataDir(), hash)
	store, err := analysis.NewStore(storagePath)
	if err != nil {
		// TODO: don't crash the whole server just because 1
		// folder fails to grasp the storagePath
		panic(err)
	}
	s.stores[uri] = store
	folderPath := util.UriToPath(uri)
	s.indexFolder(folderPath)
}

func (s *workspaceStore) indexFolder(folderPath string) {
	go func() {
		log.Println("Start indexing")
		start := time.Now()
		count := 0
		godirwalk.Walk(folderPath, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
				if !de.IsDir() && strings.HasSuffix(path, ".php") {
					count++
					s.jobs <- path
				}
				return nil
			},
			Unsorted: true,
		})
		elapsed := time.Since(start)
		log.Printf("Finished indexing %d files in %s", count, elapsed)
	}()
}

func (s *workspaceStore) getStore(uri protocol.DocumentURI) *analysis.Store {
	for workspaceURI, store := range s.stores {
		if strings.HasPrefix(uri, workspaceURI) {
			return store
		}
	}
	return nil
}

func (s *workspaceStore) addDocument(filePath string) {
	store := s.getStore(util.PathToUri(filePath))
	if store != nil {
		store.IndexDocument(filePath)
	}
}
