package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) didOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return StoreNotFound(uri)
	}
	store.OpenDocument(uri)
	return nil
}

func (s *Server) didChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return StoreNotFound(uri)
	}
	return store.ChangeDocument(uri, params.ContentChanges)
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
