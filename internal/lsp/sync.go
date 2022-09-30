package lsp

import (
	"context"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil
	}
	document := store.OpenDocument(ctx, uri)
	if document != nil {
		s.provideDiagnostics(ctx, store, document)
	}
	return nil
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	return s.store.changeDocument(ctx, params)
}

func (s *Server) didClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil
	}
	store.CloseDocument(ctx, uri)
	return nil
}

func (s *Server) didChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	go func() {
		var wg sync.WaitGroup
		changes := append(params.Changes[:0:0], params.Changes...)
		for _, change := range changes {
			if change.Type == protocol.Deleted {
				s.store.deleteJobs <- change.URI
				continue
			}

			filePath, err := util.URIToPath(change.URI)
			if err != nil {
				continue
			}
			stats, err := os.Stat(filePath)
			if err != nil {
				if !os.IsNotExist(err) {
					log.Printf("didChangeWatchedFiles error: %v, filePath: %s", err, filePath)
				}
				continue
			}
			matched := strings.HasSuffix(filePath, ".php")

			if matched && !stats.IsDir() {
				wg.Add(1)
				s.store.createJobs <- creatorJob{
					uri:       change.URI,
					ctx:       ctx,
					waitGroup: &wg,
				}
				continue
			}

			if !stats.IsDir() {
				continue
			}

			err = filepath.WalkDir(filePath, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && strings.HasSuffix(path, ".php") {
					wg.Add(1)
					s.store.createJobs <- creatorJob{
						uri:       util.PathToURI(path),
						ctx:       ctx,
						waitGroup: &wg,
					}
				}
				return nil
			})

			if err != nil {
				log.Println(err)
			}
		}
		wg.Wait()
	}()
	return nil
}
