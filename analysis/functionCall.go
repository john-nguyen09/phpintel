package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	Expression
	hasResolved bool
}

func tryToNewDefine(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				nameLowerCase := strings.ToLower(document.getPhraseText(p))
				if nameLowerCase == "\\define" || nameLowerCase == "define" {
					return newDefine(a, document, node)
				}
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newFunctionCall(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	functionCall := &FunctionCall{
		Expression: Expression{},
	}
	document.addSymbol(functionCall)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	firstChild := child
	if firstChild != nil {
		functionCall.Location = document.GetNodeLocation(firstChild)
		functionCall.Name = document.GetNodeText(firstChild)
		document.currImportTable().useFunctionOrClass(NewTypeString(functionCall.Name))
	}
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ArgumentExpressionList {
				scanNode(a, document, p)
				break
			}
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
	functions := store.GetFunctions(document.currImportTable().GetFunctionReferenceFQN(store, typeString))
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
	typeString.SetFQN(document.currImportTable().GetFunctionReferenceFQN(store, typeString))
	for _, function := range store.GetFunctions(typeString.GetFQN()) {
		functions = append(functions, function)
	}
	return functions
}
