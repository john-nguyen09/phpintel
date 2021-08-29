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
		return nil, nil
	}
	document := store.GetOrCreateDocument(ctx, uri)
	if document == nil {
		return nil, nil
	}
	pos := params.TextDocumentPositionParams.Position
	nodes := document.NodeSpineAt(document.OffsetAtPosition(pos))
	// log.Printf("Reference: %v %s", pos, nodes)
	parent := nodes.Parent()
	fallbackHandler := func() {
		sym := document.HasTypesAtPos(pos)
		refs := analysis.SymToRefs(document, sym)
		for _, ref := range refs {
			results = append(results, store.GetReferences(ref)...)
		}
	}
	switch parent.Type {
	case phrase.Identifier:
		node := nodes.Parent()
		switch node.Type {
		case phrase.MethodDeclarationHeader, phrase.ClassConstElement:
			ref := "."
			ref += document.GetNodeText(&parent)
			if node.Type == phrase.MethodDeclarationHeader {
				ref += "()"
			}
			results = append(results, store.GetReferences(ref)...)
		default:
			fallbackHandler()
		}
	case phrase.PropertyElement:
		ref := "."
		ref += document.GetNodeText(&parent)
		results = append(results, store.GetReferences(ref)...)
	case phrase.ClassDeclarationHeader,
		phrase.InterfaceDeclarationHeader,
		phrase.TraitDeclarationHeader:
		nameToken := nodes.Token()
		name := analysis.NewTypeString(document.GetNodeText(&nameToken))
		name.SetNamespace(document.ImportTableAtPos(document.NodeRange(nameToken).Start).GetNamespace())
		ref := name.GetFQN()
		results = append(results, store.GetReferences(ref)...)
	case phrase.FunctionDeclarationHeader:
		nameToken := nodes.Token()
		name := analysis.NewTypeString(document.GetNodeText(&nameToken))
		refs := document.ImportTableAtPos(document.NodeRange(nameToken).Start).FunctionPossibleFQNs(name)
		for _, ref := range refs {
			ref += "()"
			results = append(results, store.GetReferences(ref)...)
		}
	default:
		fallbackHandler()
	}
	return results, nil
}

func (s *Server) rename(ctx context.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	uri := params.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, nil
	}
	doc := store.GetOrCreateDocument(ctx, uri)
	if doc == nil {
		return nil, nil
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
