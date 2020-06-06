package lsp

import (
	"context"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func classScopeFQNAt(document *analysis.Document, pos protocol.Position) string {
	classScope := document.ClassAt(pos)
	var scopeFQN string
	if classScope != nil {
		switch c := classScope.(type) {
		case *analysis.Class:
			scopeFQN = c.Name.GetFQN()
		case *analysis.Interface:
			scopeFQN = c.Name.GetFQN()
		case *analysis.Trait:
			scopeFQN = c.Name.GetFQN()
		}
	}
	return scopeFQN
}

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
	q := analysis.NewQuery(store)
	resolveCtx := analysis.NewResolveContext(q, document)
	nodes := document.NodeSpineAt(document.OffsetAtPosition(pos))
	// log.Printf("Reference: %v %s", pos, nodes)
	parent := nodes.Parent()
	switch parent.Type {
	case phrase.Identifier:
		node := nodes.Parent()
		switch node.Type {
		case phrase.MethodDeclarationHeader, phrase.ClassConstElement:
			scopeFQN := classScopeFQNAt(document, pos)
			if scopeFQN != "" {
				name := document.GetNodeText(&parent)
				if node.Type == phrase.MethodDeclarationHeader {
					name += "()"
				}
				fqn := scopeFQN + "::" + name
				results = append(results, store.GetReferences(fqn)...)
			}
		}
	case phrase.PropertyElement:
		scopeFQN := classScopeFQNAt(document, pos)
		if scopeFQN != "" {
			name := document.GetNodeText(&parent)
			fqn := scopeFQN + "::" + name
			results = append(results, store.GetReferences(fqn)...)
		}
	case phrase.FunctionDeclarationHeader,
		phrase.ClassDeclarationHeader,
		phrase.InterfaceDeclarationHeader,
		phrase.TraitDeclarationHeader:
		nameToken := nodes.Token()
		name := analysis.NewTypeString(document.GetNodeText(&nameToken))
		name.SetNamespace(document.ImportTableAtPos(pos).GetNamespace())
		results = append(results, store.GetReferences(name.GetFQN())...)
	}
	sym := document.HasTypesAtPos(pos)
	switch v := sym.(type) {
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		fqn := document.ImportTableAtPos(v.GetLocation().Range.Start).GetFunctionReferenceFQN(q, name)
		results = store.GetReferences(fqn)
	case *analysis.ClassTypeDesignator, *analysis.TypeDeclaration, *analysis.ClassAccess, *analysis.TraitAccess:
		for _, t := range v.GetTypes().Resolve() {
			results = append(results, store.GetReferences(t.GetFQN())...)
		}
	case *analysis.MethodAccess, *analysis.PropertyAccess, *analysis.ScopedMethodAccess, *analysis.ScopedPropertyAccess, *analysis.ScopedConstantAccess:
		sym.Resolve(resolveCtx)
		h := sym.(analysis.HasTypesHasScope)
		for _, t := range h.GetScopeTypes().Resolve() {
			fqn := t.GetFQN() + "::" + h.MemberName()
			results = append(results, store.GetReferences(fqn)...)
		}
	}
	return results, nil
}

func (s *Server) rename(params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	doc := store.GetOrCreateDocument(uri)
	if doc == nil {
		return nil, DocumentNotFound(uri)
	}
	pos := params.Position
	symbol := doc.HasTypesAtPos(pos)
	result := &protocol.WorkspaceEdit{
		Changes: make(map[string][]protocol.TextEdit),
	}
	switch v := symbol.(type) {
	case *analysis.Variable:
		varTable := doc.GetVariableTableAt(pos)
		if varTable != nil {
			var changes []protocol.TextEdit
			for _, ctxVar := range varTable.GetContextualVariables(v.Name) {
				changes = append(changes, protocol.TextEdit{
					Range:   ctxVar.Variable().Location.Range,
					NewText: params.NewName,
				})
			}
			result.Changes[uri] = changes
		}
	}
	return result, nil
}
