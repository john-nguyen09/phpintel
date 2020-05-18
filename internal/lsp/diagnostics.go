package lsp

import (
	"context"
	"log"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) provideDiagnostics(ctx context.Context, document *analysis.Document) {
	diagnostics := analysis.GetParserDiagnostics(document)
	diagnostics = append(diagnostics, analysis.UnusedDiagnostics(document)...)
	params := &protocol.PublishDiagnosticsParams{
		URI:         document.GetURI(),
		Diagnostics: diagnostics,
	}
	err := s.client.PublishDiagnostics(ctx, params)
	if err != nil {
		log.Println(err)
	}
}
