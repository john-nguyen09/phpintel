package lsp

import (
	"context"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/karrick/godirwalk"
)

const numCreators int = 2
const numDeletors int = 1

type CreatorJob struct {
	filePath  string
	waitGroup *sync.WaitGroup
}

type workspaceStore struct {
	server     *Server
	ctx        context.Context
	stores     []*analysis.Store
	createJobs chan CreatorJob
	deleteJobs chan string
}

func newWorkspaceStore(server *Server, ctx context.Context) *workspaceStore {
	workspaceStore := &workspaceStore{
		server:     server,
		ctx:        ctx,
		stores:     []*analysis.Store{},
		createJobs: make(chan CreatorJob),
		deleteJobs: make(chan string),
	}
	for i := 0; i < numCreators; i++ {
		go workspaceStore.newCreator(i)
	}
	for i := 0; i < numDeletors; i++ {
		go workspaceStore.newDeletor(i)
	}
	return workspaceStore
}

func (s *workspaceStore) newCreator(id int) {
	for job := range s.createJobs {
		uri := util.PathToUri(job.filePath)
		store := s.getStore(uri)
		if store == nil {
			if job.waitGroup != nil {
				job.waitGroup.Done()
			}
			continue
		}
		s.addDocument(store, job.filePath)
		if job.waitGroup != nil {
			job.waitGroup.Done()
		}
	}
}

func (s *workspaceStore) newDeletor(id int) {
	for uri := range s.deleteJobs {
		store := s.getStore(uri)
		if store == nil {
			continue
		}
		s.removeDocument(store, uri)
	}
}

func (s *workspaceStore) close() {
	for _, store := range s.stores {
		store.Close()
	}
}

func (s *workspaceStore) addView(server *Server, ctx context.Context, uri protocol.DocumentURI) {
	storagePath := filepath.Join(getDataDir(), util.GetURIID(uri))
	store, err := analysis.NewStore(uri, storagePath)
	if err != nil {
		// TODO: don't crash the whole server just because 1
		// folder fails to grasp the storagePath
		panic(err)
	}
	store.Migrate(protocol.GetVersion(ctx))
	store.LoadStubs()
	s.stores = append(s.stores, store)
	folderPath, err := util.UriToPath(uri)
	if err != nil {
		log.Printf("addView error: %v", err)
		return
	}
	s.indexFolder(store, folderPath)
	err = s.registerFileWatcher(folderPath, server, ctx)
	if err != nil {
		log.Printf("addView error: %v", err)
	}
}

func (s *workspaceStore) removeView(server *Server, ctx context.Context, uri protocol.DocumentURI) {
	store := s.getStore(uri)
	if store == nil {
		log.Println(StoreNotFound(uri))
		return
	}
	defer store.Close()
	s.removeStore(store.GetURI())
	folderPath, err := util.UriToPath(uri)
	if err != nil {
		log.Printf("removeView error: %v", err)
	}
	s.unregisterFileWatcher(folderPath, server, ctx)
}

func getFileWatcherID(path string) string {
	return path + "-fileWatcher"
}

func (s *workspaceStore) registerFileWatcher(path string, server *Server, ctx context.Context) error {
	// fileExtensions := "php"
	regParams := protocol.DidChangeWatchedFilesRegistrationOptions{
		Watchers: []protocol.FileSystemWatcher{{
			GlobPattern: path + "/**/*",
			Kind:        int(protocol.WatchCreate + protocol.WatchDelete),
		}},
	}
	return server.client.RegisterCapability(ctx, &protocol.RegistrationParams{
		Registrations: []protocol.Registration{
			protocol.Registration{
				ID: getFileWatcherID(path), Method: "workspace/didChangeWatchedFiles", RegisterOptions: regParams,
			},
		},
	})
}

func (s *workspaceStore) unregisterFileWatcher(path string, server *Server, ctx context.Context) error {
	return server.client.UnregisterCapability(ctx, &protocol.UnregistrationParams{
		Unregisterations: []protocol.Unregistration{
			protocol.Unregistration{
				ID: getFileWatcherID(path), Method: "workspace/didChangeWatchedFiles",
			},
		},
	})
}

func (s *workspaceStore) indexFolder(store *analysis.Store, folderPath string) {
	var waitGroup sync.WaitGroup
	store.PrepareForIndexing()
	go func() {
		log.Println("Start indexing")
		start := time.Now()
		count := 0
		godirwalk.Walk(folderPath, &godirwalk.Options{
			Callback: func(path string, de *godirwalk.Dirent) error {
				if !de.IsDir() && strings.HasSuffix(path, ".php") {
					count++
					waitGroup.Add(1)
					s.createJobs <- CreatorJob{
						filePath:  path,
						waitGroup: &waitGroup,
					}
				}
				return nil
			},
			Unsorted: true,
		})
		waitGroup.Wait()
		store.FinishIndexing()
		elapsed := time.Since(start)
		log.Printf("Finished indexing %d files in %s", count, elapsed)
	}()
}

func (s *workspaceStore) getStore(uri protocol.DocumentURI) *analysis.Store {
	for _, store := range s.stores {
		if strings.HasPrefix(uri, store.GetURI()) {
			return store
		}
	}
	return nil
}

func (s *workspaceStore) removeStore(uri protocol.DocumentURI) {
	for i, store := range s.stores {
		if store.GetURI() == uri {
			s.stores = append(s.stores[:i], s.stores[i+1:]...)
			break
		}
	}
}

func (s *workspaceStore) addDocument(store *analysis.Store, filePath string) {
	store.CompareAndIndexDocument(filePath)
}

func (s *workspaceStore) removeDocument(store *analysis.Store, uri string) {
	store.DeleteDocument(uri)
	store.DeleteFolder(uri)
}

func (s *workspaceStore) changeDocument(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.getStore(uri)
	if store == nil {
		return StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return DocumentNotFound(uri)
	}
	newDoc := document.CloneForMutate()
	newDoc.ApplyChanges(params.ContentChanges)
	store.SyncDocument(newDoc)
	store.SaveDocOnStore(newDoc)
	s.server.provideDiagnostics(ctx, store, newDoc)
	return nil
}
