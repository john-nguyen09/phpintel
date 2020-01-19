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
			classes, _ := store.SearchClasses(params.Query, analysis.NewSearchOptions())
			for _, class := range classes {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Class,
					Name:     class.Name.GetFQN(),
					Location: class.GetLocation(),
				})
			}
			consts, _ := store.SearchConsts(params.Query, analysis.NewSearchOptions())
			for _, constant := range consts {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Constant,
					Name:     constant.Name.GetFQN(),
					Location: constant.GetLocation(),
				})
			}
			classConsts, _ := store.SearchClassConsts("", params.Query, analysis.NewSearchOptions())
			for _, classConst := range classConsts {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Constant,
					Name:          classConst.Name,
					Location:      classConst.GetLocation(),
					ContainerName: classConst.Scope.GetFQN(),
				})
			}
			defines, _ := store.SearchDefines(params.Query, analysis.NewSearchOptions())
			for _, define := range defines {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Constant,
					Name:     define.Name.GetFQN(),
					Location: define.GetLocation(),
				})
			}
			functions, _ := store.SearchFunctions(params.Query, analysis.NewSearchOptions())
			for _, function := range functions {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Function,
					Name:     function.Name.GetFQN(),
					Location: function.GetLocation(),
				})
			}
			interfaces, _ := store.SearchInterfaces(params.Query, analysis.NewSearchOptions())
			for _, theInterface := range interfaces {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:     protocol.Interface,
					Name:     theInterface.Name.GetFQN(),
					Location: theInterface.GetLocation(),
				})
			}
			methods, _ := store.SearchMethods("", params.Query, analysis.NewSearchOptions())
			for _, method := range methods {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Method,
					Name:          method.Name,
					Location:      method.GetLocation(),
					ContainerName: method.Scope.GetFQN(),
				})
			}
			properties, _ := store.SearchProperties("", params.Query, analysis.NewSearchOptions())
			for _, property := range properties {
				symbols = append(symbols, protocol.SymbolInformation{
					Kind:          protocol.Property,
					Name:          property.Name,
					Location:      property.GetLocation(),
					ContainerName: property.Scope.GetFQN(),
				})
			}
			traits, _ := store.SearchTraits(params.Query, analysis.NewSearchOptions())
			for _, trait := range traits {
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
