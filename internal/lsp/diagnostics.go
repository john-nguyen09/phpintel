package lsp

import (
	"context"
	"log"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/internal/xcontext"
)

func (s *Server) provideDiagnostics(ctx context.Context, store *analysis.Store, document *analysis.Document) {
	diagnostics := analysis.GetParserDiagnostics(document)
	diagnostics = append(diagnostics, analysis.UnusedDiagnostics(document)...)
	store.DebouncedDeprecation(func() {
		ctx = xcontext.Detach(ctx)
		params := &protocol.PublishDiagnosticsParams{
			URI:         document.GetURI(),
			Diagnostics: analysis.DeprecatedDiagnostics(analysis.NewResolveContext(store, document)),
		}
		err := s.client.PublishDiagnostics(ctx, params)
		if err != nil {
			log.Println(err)
		}
	})
	params := &protocol.PublishDiagnosticsParams{
		URI:         document.GetURI(),
		Diagnostics: diagnostics,
	}
	err := s.client.PublishDiagnostics(ctx, params)
	if err != nil {
		log.Println(err)
	}
}
