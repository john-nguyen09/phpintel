package lsp

import (
	"context"
	"log"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

const numCreators int = 2
const numDeletors int = 1

type creatorJob struct {
	uri       string
	ctx       context.Context
	waitGroup *sync.WaitGroup
}

type workspaceStore struct {
	server     *Server
	ctx        context.Context
	stores     []*analysis.Store
	createJobs chan creatorJob
	deleteJobs chan string
}

func newWorkspaceStore(ctx context.Context, server *Server) *workspaceStore {
	workspaceStore := &workspaceStore{
		server:     server,
		ctx:        ctx,
		stores:     []*analysis.Store{},
		createJobs: make(chan creatorJob),
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
		uri, err := util.DecodeURIFromQuery(job.uri)
		if err != nil {
			log.Printf("workspaceStore.getStore cannot DecodeURIFromQuery %s, err: %v", job.uri, err)
			continue
		}
		store := s.getStore(uri)
		if store == nil {
			if job.waitGroup != nil {
				job.waitGroup.Done()
			}
			log.Printf("workspaceStore.newCreator store not found: %s", uri)
			log.Printf("Stores:")
			for _, store := range s.stores {
				log.Println(store.GetURI())
			}
			continue
		}
		store.CompareAndIndexDocument(job.ctx, uri)
		if job.waitGroup != nil {
			job.waitGroup.Done()
		}
	}
}

func (s *workspaceStore) newDeletor(id int) {
	for uri := range s.deleteJobs {
		var err error
		uri, err = util.DecodeURIFromQuery(uri)
		if err != nil {
			log.Printf("workspaceStore.getStore cannot DecodeURIFromQuery %s, err: %v", uri, err)
			continue
		}
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

func (s *workspaceStore) addView(ctx context.Context, server *Server, uri protocol.DocumentURI) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Printf("%s: %v", uri, err)
		return
	}
	var (
		fs       protocol.FS
		rootPath string = uri
	)
	switch {
	case s.server.fileExtensionsSupported:
		fs = protocol.NewLSPFS(s.server.Conn)
	case u.Scheme == "file":
		var err error
		fs = protocol.NewFileFS()
		rootPath, err = util.URIToPath(uri)
		if err != nil {
			log.Printf("addView: %s - %v", uri, err)
			return
		}
	}
	if fs == nil {
		log.Printf("No FS found for: %s", uri)
		return
	}
	storagePath := filepath.Join(getDataDir(), "data", util.GetURIID(uri))
	store, err := analysis.NewStore(fs, uri, storagePath)
	if err != nil {
		log.Printf("%s: %v", uri, err)
		return
	}
	store.Migrate(protocol.GetVersion(ctx))
	store.LoadStubs()
	s.stores = append(s.stores, store)
	if err != nil {
		log.Printf("addView error: %v", err)
		return
	}
	s.indexFolder(ctx, store, rootPath)
	err = s.registerFileWatcher(ctx, uri, server)
	if err != nil {
		log.Printf("addView error: %v", err)
	}
}

func (s *workspaceStore) removeView(ctx context.Context, server *Server, uri protocol.DocumentURI) {
	store := s.getStore(uri)
	if store == nil {
		return
	}
	defer store.Close()
	s.removeStore(store.GetURI())
	s.unregisterFileWatcher(ctx, uri, server)
}

func getFileWatcherID(base string) string {
	return base + "-fileWatcher"
}

func (s *workspaceStore) registerFileWatcher(ctx context.Context, base string, server *Server) error {
	// fileExtensions := "php"
	regParams := protocol.DidChangeWatchedFilesRegistrationOptions{
		Watchers: []protocol.FileSystemWatcher{{
			GlobPattern: "**/*",
			Kind:        int(protocol.WatchCreate + protocol.WatchChange + protocol.WatchDelete),
		}},
	}
	return server.client.RegisterCapability(ctx, &protocol.RegistrationParams{
		Registrations: []protocol.Registration{
			{
				ID: getFileWatcherID(base), Method: "workspace/didChangeWatchedFiles", RegisterOptions: regParams,
			},
		},
	})
}

func (s *workspaceStore) unregisterFileWatcher(ctx context.Context, path string, server *Server) error {
	return server.client.UnregisterCapability(ctx, &protocol.UnregistrationParams{
		Unregisterations: []protocol.Unregistration{
			{
				ID: getFileWatcherID(path), Method: "workspace/didChangeWatchedFiles",
			},
		},
	})
}

func (s *workspaceStore) indexFolder(ctx context.Context, store *analysis.Store, rootPath string) {
	var waitGroup sync.WaitGroup
	store.PrepareForIndexing()
	go func() {
		log.Println("Start indexing")
		start := time.Now()
		count := 0
		if docs, err := store.FS.ListFiles(ctx, rootPath); err == nil {
			for _, doc := range docs {
				if strings.HasSuffix(doc.URI, ".php") {
					count++
					waitGroup.Add(1)
					s.createJobs <- creatorJob{
						uri:       store.FS.ConvertToURI(doc.URI),
						ctx:       ctx,
						waitGroup: &waitGroup,
					}
				}
			}
		} else {
			log.Printf("indexFolder: %v", err)
			return
		}
		waitGroup.Wait()
		store.FinishIndexing()
		elapsed := time.Since(start)
		log.Printf("Finished indexing %d files in %s", count, elapsed)
		util.PrintMemUsage()
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

func (s *workspaceStore) removeDocument(store *analysis.Store, uri string) {
	store.DeleteDocument(uri)
	store.DeleteFolder(uri)
}

func (s *workspaceStore) changeDocument(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.getStore(uri)
	if store == nil {
		return nil
	}
	document := store.GetOrCreateDocument(ctx, uri)
	if document == nil {
		return nil
	}
	newDoc := document.CloneForMutate()
	newDoc.ApplyChanges(params.ContentChanges)
	store.SyncDocument(newDoc)
	store.SaveDocOnStore(newDoc)
	s.server.provideDiagnostics(ctx, store, newDoc)
	return nil
}
