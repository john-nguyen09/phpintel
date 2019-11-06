package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/cmd"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/pkg/errors"
)

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, errors.Errorf("store not found for %s", uri)
	}
	document := store.GetOrCreateDocument(uri)
	document.Load()
	symbol := document.SymbolAtPos(params.TextDocumentPositionParams.Position)
	hover := protocol.Hover{}

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
		}
	case *analysis.ClassAccess:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
		}
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		for _, typeString := range v.Type.Resolve() {
			consts = append(consts, store.GetConsts(typeString.GetFQN())...)
			if len(consts) > 0 {
				hover = cmd.ConstToHover(symbol, *consts[0])
				break
			}
			defines = append(defines, store.GetDefines(typeString.GetFQN())...)
			if len(defines) > 0 {
				hover = cmd.DefineToHover(symbol, *defines[0])
				break
			}
		}
	case *analysis.FunctionCall:
		functions := []*analysis.Function{}
		for _, typeString := range v.Type.Resolve() {
			functions = append(functions, store.GetFunctions(typeString.GetFQN())...)
			if len(functions) > 0 {
				hover = cmd.FunctionToHover(symbol, *functions[0])
				break
			}
		}
	case *analysis.ScopedConstantAccess:
		classConsts := []*analysis.ClassConst{}
		for _, typeString := range v.Type.Resolve() {
			for _, scopeType := range v.Scope.GetTypes() {
				classConsts = append(classConsts, store.GetClassConsts(scopeType, typeString.GetFQN())...)
			}
		}
	}
	return &hover, nil
}
