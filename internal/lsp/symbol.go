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
	scopedSymbols := make(map[string][]analysis.Symbol)
	for _, child := range document.Children {
		switch v := child.(type) {
		case analysis.HasScope:
			key := v.GetScope().GetFQN()
			scopedSymbols[key] = append(scopedSymbols[key], child)
		}
	}

	for _, child := range document.Children {
		switch child.(type) {
		case analysis.HasScope:
			continue
		}
		symbol, fqn := symbolToProtocolDocumentSymbol(child)
		if symbol.Kind == 0 {
			continue
		}
		if childSymbols, ok := scopedSymbols[fqn]; ok {
			for _, childSymbol := range childSymbols {
				childDocumentSymbol, _ := symbolToProtocolDocumentSymbol(childSymbol)
				if childDocumentSymbol.Kind == 0 {
					continue
				}
				symbol.Children = append(symbol.Children, childDocumentSymbol)
			}
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func symbolToProtocolDocumentSymbol(symbol analysis.Symbol) (protocol.DocumentSymbol, string) {
	switch v := symbol.(type) {
	case *analysis.Class:
		return protocol.DocumentSymbol{
			Kind:           protocol.Class,
			Name:           v.GetName(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, v.Name.GetFQN()
	case *analysis.Const:
		return protocol.DocumentSymbol{
			Kind:           protocol.Constant,
			Name:           v.GetName(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.ClassConst:
		return protocol.DocumentSymbol{
			Kind:           protocol.Constant,
			Name:           v.GetName(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.Define:
		return protocol.DocumentSymbol{
			Kind:           protocol.Constant,
			Name:           v.GetName(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.Function:
		return protocol.DocumentSymbol{
			Kind:           protocol.Function,
			Name:           v.GetName().GetOriginal(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.Interface:
		return protocol.DocumentSymbol{
			Kind:           protocol.Interface,
			Name:           v.Name.GetOriginal(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, v.Name.GetFQN()
	case *analysis.Method:
		return protocol.DocumentSymbol{
			Kind:           protocol.Method,
			Name:           v.Name,
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.Property:
		return protocol.DocumentSymbol{
			Kind:           protocol.Property,
			Name:           v.Name,
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, ""
	case *analysis.Trait:
		return protocol.DocumentSymbol{
			Kind:           protocol.Class,
			Name:           v.GetName(),
			Detail:         v.GetDescription(),
			Range:          v.GetLocation().Range,
			SelectionRange: v.GetLocation().Range,
		}, v.Name.GetFQN()
	}
	return protocol.DocumentSymbol{}, ""
}

func (s *Server) workspaceSymbol(ctx context.Context, params *protocol.WorkspaceSymbolParams) ([]protocol.SymbolInformation, error) {
	symbols := []protocol.SymbolInformation{}
	if params.Query != "" {
		for _, store := range s.store.stores {
			for _, class := range store.SearchClasses(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Class,
					Name:     class.Name.GetFQN(),
					Location: class.GetLocation(),
				})
			}
			for _, constant := range store.SearchConsts(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Constant,
					Name:     constant.Name.GetFQN(),
					Location: constant.GetLocation(),
				})
			}
			for _, classConst := range store.SearchClassConsts("", params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Constant,
					Name:          classConst.Name,
					Location:      classConst.GetLocation(),
					ContainerName: classConst.Scope.GetFQN(),
				})
			}
			for _, define := range store.SearchDefines(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Constant,
					Name:     define.Name.GetFQN(),
					Location: define.GetLocation(),
				})
			}
			for _, function := range store.SearchFunctions(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Function,
					Name:     function.Name.GetFQN(),
					Location: function.GetLocation(),
				})
			}
			for _, theInterface := range store.SearchInterfaces(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Interface,
					Name:     theInterface.Name.GetFQN(),
					Location: theInterface.GetLocation(),
				})
			}
			for _, method := range store.SearchMethods("", params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Method,
					Name:          method.Name,
					Location:      method.GetLocation(),
					ContainerName: method.Scope.GetFQN(),
				})
			}
			for _, property := range store.SearchProperties("", params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Property,
					Name:          property.Name,
					Location:      property.GetLocation(),
					ContainerName: property.Scope.GetFQN(),
				})
			}
			for _, trait := range store.SearchTraits(params.Query) {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Class,
					Name:     trait.Name.GetFQN(),
					Location: trait.GetLocation(),
				})
			}
		}
	}
	return symbols, nil
}
