package lsp

import (
	"context"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"runtime/pprof"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func newNoDuplicateMethodsOptions() analysis.SearchOptions {
	excludeNames := map[string]bool{}
	return analysis.NewSearchOptions().
		WithPredicate(func(symbol analysis.Symbol) bool {
			method := symbol.(*analysis.Method)
			methodKey := method.GetName()
			if _, ok := excludeNames[methodKey]; ok {
				return false
			}
			excludeNames[methodKey] = true
			return true
		})
}

func staticMethodsScopeAware(opt analysis.SearchOptions, classScope string,
	name string) analysis.SearchOptions {
	return opt.WithPredicate(func(symbol analysis.Symbol) bool {
		method := symbol.(*analysis.Method)
		if analysis.IsNameParent(name) {
			// parent:: excludes methods from current class
			if method.GetScope().GetFQN() == classScope {
				return false
			}
			// or from parents but private
			if method.VisibilityModifier == analysis.Private {
				return false
			}
			return true
		}
		// static:: and self:: exclude private methods that are not from current class
		if analysis.IsNameRelative(name) {
			if method.GetScope().GetFQN() != classScope &&
				method.VisibilityModifier == analysis.Private {
				return false
			}
			// And also accept non-static
			return true
		}
		// Not parent:: or static:: or self:: so accept only public static
		return method.IsStatic && method.VisibilityModifier == analysis.Public
	})
}

func newNoDuplicatePropsOptions() analysis.SearchOptions {
	excludeNames := map[string]bool{}
	return analysis.NewSearchOptions().
		WithPredicate(func(symbol analysis.Symbol) bool {
			prop := symbol.(*analysis.Property)
			propKey := prop.GetName()
			if _, ok := excludeNames[propKey]; ok {
				return false
			}
			excludeNames[propKey] = true
			return true
		})
}

func staticPropsScopeAware(opt analysis.SearchOptions, classScope string, name string) analysis.SearchOptions {
	return opt.WithPredicate(func(symbol analysis.Symbol) bool {
		prop := symbol.(*analysis.Property)
		// Properties are different from methods,
		// and static can only be accessed using :: (static::, self::, parent::, TestClass1::)
		if !prop.IsStatic {
			return false
		}
		if analysis.IsNameParent(name) {
			if prop.GetScope().GetFQN() == classScope || prop.VisibilityModifier == analysis.Private {
				return false
			}
			return true
		}
		if analysis.IsNameRelative(name) {
			if prop.GetScope().GetFQN() != classScope && prop.VisibilityModifier == analysis.Private {
				return false
			}
		}
		return prop.VisibilityModifier == analysis.Public
	})
}

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
			DocumentSymbolProvider: true,
			HoverProvider:          true,
			SignatureHelpProvider: &protocol.SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
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
			WorkspaceSymbolProvider: true,
		},
	}, nil
}

func (s *Server) initialized(ctx context.Context, params *protocol.InitializedParams) error {
	s.stateMu.Lock()
	s.state = serverInitialized
	s.stateMu.Unlock()
	version := protocol.GetVersion(ctx)
	log.Println("phpintel server initialised. Version: " + version)
	for _, folder := range s.pendingFolders {
		s.store.addView(s, ctx, folder.URI)
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
	memprofile := protocol.GetMemprofile(ctx)
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal("Could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not start memory profile: ", err)
		}
	}
	return nil
}
