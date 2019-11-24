package lsp

import (
	"context"
	"path"
	"path/filepath"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func getDataDir() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		// No storage what can we do here?
		panic(err)
	}
	return filepath.Join(homeDir, ".phpintel")
}

func (s *Server) initialize(ctx context.Context, params *protocol.InitializeParams) (*protocol.InitializeResult, error) {
	s.stateMu.Lock()
	state := s.state
	if state >= serverInitializing {
		return nil, jsonrpc2.NewErrorf(jsonrpc2.CodeInvalidRequest, "already initialised")
	}
	s.state = serverInitializing
	s.stateMu.Unlock()

	s.pendingFolders = params.WorkspaceFolders
	if len(s.pendingFolders) == 0 {
		if params.RootURI != "" {
			s.pendingFolders = []protocol.WorkspaceFolder{{
				URI:  params.RootURI,
				Name: path.Base(params.RootURI),
			}}
		} else {
			return nil, errors.Errorf("single file is not supported")
		}
	}
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			CompletionProvider: &protocol.CompletionOptions{
				TriggerCharacters: []string{
					"$", ">", ":",
					".", "<", "/",
				},
			},
			DefinitionProvider:     true,
			DocumentSymbolProvider: false,
			HoverProvider:          true,
			SignatureHelpProvider:  nil,
			// SignatureHelpProvider: &protocol.SignatureHelpOptions{
			// 	TriggerCharacters: []string{"(", ","},
			// },
			TextDocumentSync: &protocol.TextDocumentSyncOptions{
				Change:    protocol.Incremental,
				OpenClose: true,
				Save: &protocol.SaveOptions{
					IncludeText: false,
				},
			},
			Workspace: &struct {
				WorkspaceFolders *struct {
					Supported           bool   "json:\"supported,omitempty\""
					ChangeNotifications string "json:\"changeNotifications,omitempty\""
				} "json:\"workspaceFolders,omitempty\""
			}{
				WorkspaceFolders: &struct {
					Supported           bool   "json:\"supported,omitempty\""
					ChangeNotifications string "json:\"changeNotifications,omitempty\""
				}{
					Supported:           true,
					ChangeNotifications: "workspace/didChangeWorkspaceFolders",
				},
			},
		},
	}, nil
}

func (s *Server) initialized(ctx context.Context, params *protocol.InitializedParams) error {
	s.stateMu.Lock()
	s.state = serverInitialized
	s.stateMu.Unlock()
	for _, folder := range s.pendingFolders {
		s.store.addView(folder.URI)
	}
	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	if s.state < serverInitialized {
		return jsonrpc2.NewErrorf(jsonrpc2.CodeInvalidRequest, "not intialised")
	}
	s.store.close()
	return nil
}
