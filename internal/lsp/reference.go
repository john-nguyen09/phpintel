package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) references(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	results := []protocol.Location{}
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	pos := params.TextDocumentPositionParams.Position
	resolveCtx := analysis.NewResolveContext(store, document)
	sym := document.HasTypesAtPos(pos)
	switch v := sym.(type) {
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		fqn := document.ImportTableAtPos(v.GetLocation().Range.Start).GetFunctionReferenceFQN(store, name)
		results = store.GetReferences(fqn)
	case *analysis.ClassTypeDesignator, *analysis.TypeDeclaration, *analysis.ClassAccess, *analysis.TraitAccess:
		for _, t := range v.GetTypes().Resolve() {
			results = append(results, store.GetReferences(t.GetFQN())...)
		}
	case *analysis.MethodAccess, *analysis.PropertyAccess, *analysis.ScopedMethodAccess, *analysis.ScopedPropertyAccess:
		sym.Resolve(resolveCtx)
		h := sym.(analysis.HasTypesHasScope)
		for _, t := range h.GetScopeTypes().Resolve() {
			fqn := t.GetFQN() + "->" + h.MemberName()
			results = append(results, store.GetReferences(fqn)...)
		}
	}
	return results, nil
}
