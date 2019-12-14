package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) documentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]protocol.DocumentSymbol, error) {
	symbols := []protocol.DocumentSymbol{}
	uri := params.TextDocument.URI
	store := s.store.getStore(params.TextDocument.URI)
	if store == nil {
		return symbols, nil
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return symbols, nil
	}

	for _, child := range document.Children {
		switch v := child.(type) {
		case *analysis.Class:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Class,
				Name:           v.GetName(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Const:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Constant,
				Name:           v.GetName(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.ClassConst:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Constant,
				Name:           v.GetName(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Define:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Constant,
				Name:           v.GetName(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Function:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Function,
				Name:           v.GetName().GetOriginal(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Interface:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Interface,
				Name:           v.Name.GetOriginal(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Method:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Method,
				Name:           v.Name,
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Property:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Property,
				Name:           v.Name,
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		case *analysis.Trait:
			symbols = append(symbols, protocol.DocumentSymbol{
				Kind:           protocol.Class,
				Name:           v.GetName(),
				Detail:         v.GetDescription(),
				Range:          v.GetLocation().Range,
				SelectionRange: v.GetLocation().Range,
			})
		}
	}

	return symbols, nil
}
