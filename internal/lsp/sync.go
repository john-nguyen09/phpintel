package lsp

import (
	"context"
	"os"
	"strings"

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
	for _, change := range params.Changes {
		if change.Type == protocol.Deleted {
			s.store.deleteJobs <- change.URI
			continue
		}

		filePath := util.UriToPath(change.URI)
		matched := strings.HasSuffix(filePath, ".php")

		if matched {
			s.store.createJobs <- CreatorJob{
				filePath: filePath,
			}
			continue
		}

		stats, err := os.Stat(filePath)
		if err != nil {
			continue
		}
		if !stats.IsDir() {
			continue
		}

		go func(change protocol.FileEvent) {
			godirwalk.Walk(filePath, &godirwalk.Options{
				Callback: func(path string, de *godirwalk.Dirent) error {
					if !de.IsDir() && strings.HasSuffix(path, ".php") {
						s.store.createJobs <- CreatorJob{
							filePath: path,
						}
					}
					return nil
				},
				Unsorted: true,
			})
		}(change)
	}
	return nil
}
