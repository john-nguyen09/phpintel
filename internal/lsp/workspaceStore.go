package lsp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/log"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/karrick/godirwalk"
)

const numOfWorkers int = 4

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
	return workspaceStore
}

func (s *workspaceStore) analyse(id int, filePaths <-chan string) {
	for filePath := range filePaths {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error(s.ctx, "", err)
			continue
		}
		text := string(data)
		rootNode := parser.Parse(text)
		s.addDocument(analysis.NewDocument(util.PathToUri(filePath), text, rootNode))
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
	s.stores[hash] = store
	folderPath := util.UriToPath(uri)
	s.indexFolder(folderPath)
}

func (s *workspaceStore) indexFolder(folderPath string) {
	godirwalk.Walk(folderPath, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() && strings.HasSuffix(path, ".php") {
				s.jobs <- path
			}
			return nil
		},
		Unsorted: true,
	})
}

func (s *workspaceStore) getStore(uri protocol.DocumentURI) *analysis.Store {
	for workspaceURI, store := range s.stores {
		if strings.HasPrefix(uri, workspaceURI) {
			return store
		}
	}
	return nil
}

func (s *workspaceStore) addDocument(document *analysis.Document) {
	store := s.getStore(document.GetURI())
	if store != nil {
		store.SyncDocument(document)
	}
}
