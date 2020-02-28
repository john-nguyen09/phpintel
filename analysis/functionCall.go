package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	Expression
	hasResolved bool
}

func tryToNewDefine(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			name := strings.ToLower(document.GetNodeText(child))
			if name == "\\define" || name == "define" {
				return newDefine(document, node)
			}
		case "arguments":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	return nil
}

func newFunctionCall(document *Document, node *sitter.Node) (HasTypes, bool) {
	functionCall := &FunctionCall{
		Expression: Expression{},
	}
	document.addSymbol(functionCall)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	firstChild := child
	if firstChild != nil {
		functionCall.Location = document.GetNodeLocation(node.Child(0))
		functionCall.Name = document.GetNodeText(node.Child(0))
	}
	for child != nil {
		switch child.Type() {
		case "arguments":
			scanNode(document, child)
			break
		}
		child = traverser.Advance()
	}
	return functionCall, false
}

func (s *FunctionCall) GetLocation() protocol.Location {
	return s.Location
}

func (s *FunctionCall) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	document := ctx.document
	store := ctx.store
	s.hasResolved = true
	typeString := NewTypeString(s.Name)
	functions := store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(store, typeString))
	for _, function := range functions {
		s.Type.merge(function.returnTypes)
	}
}

func (s *FunctionCall) GetTypes() TypeComposite {
	return s.Type
}

func (s *FunctionCall) ResolveToHasParams(ctx ResolveContext) []HasParams {
	functions := []HasParams{}
	typeString := NewTypeString(s.Name)
	store := ctx.store
	document := ctx.document
	typeString.SetFQN(document.GetImportTable().GetFunctionReferenceFQN(store, typeString))
	for _, function := range store.GetFunctions(typeString.GetFQN()) {
		functions = append(functions, function)
	}
	return functions
}
