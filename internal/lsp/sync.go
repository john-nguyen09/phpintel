package lsp

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/karrick/godirwalk"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return StoreNotFound(uri)
	}
	document := store.OpenDocument(uri)
	if document != nil {
		s.provideDiagnostics(ctx, document)
	}
	return nil
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	return s.store.changeDocument(ctx, uri, params.ContentChanges)
}

func (s *Server) didClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return StoreNotFound(uri)
	}
	store.CloseDocument(uri)
	return nil
}

func (s *Server) didChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	go func() {
		start := time.Now()
		var wg sync.WaitGroup
		changes := append(params.Changes[:0:0], params.Changes...)
		log.Printf("Length before: %d", len(changes))
		for _, change := range changes {
			if change.Type == protocol.Deleted {
				s.store.deleteJobs <- change.URI
				continue
			}

			filePath, err := util.UriToPath(change.URI)
			if err != nil {
				log.Printf("didChangeWatchedFiles error: %v", err)
				continue
			}
			stats, err := os.Stat(filePath)
			if err != nil {
				log.Printf("didChangeWatchedFiles error: %v", err)
				continue
			}
			matched := strings.HasSuffix(filePath, ".php")

			if matched && !stats.IsDir() {
				wg.Add(1)
				s.store.createJobs <- CreatorJob{
					filePath:  filePath,
					waitGroup: &wg,
				}
				continue
			}

			if !stats.IsDir() {
				continue
			}

			godirwalk.Walk(filePath, &godirwalk.Options{
				Callback: func(path string, de *godirwalk.Dirent) error {
					if !de.IsDir() && strings.HasSuffix(path, ".php") {
						wg.Add(1)
						s.store.createJobs <- CreatorJob{
							filePath:  path,
							waitGroup: &wg,
						}
					}
					return nil
				},
				Unsorted: true,
			})
		}
		wg.Wait()
		elapsed := time.Since(start)
		log.Printf("%d changes took %s", len(changes), elapsed)
	}()
	return nil
}
