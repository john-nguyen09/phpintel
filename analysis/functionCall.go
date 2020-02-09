package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// FunctionCall represents a reference to function call
type FunctionCall struct {
	Expression
	hasResolved bool
}

func tryToNewDefine(document *Document, node *phrase.Phrase) Symbol {
	if len(node.Children) >= 1 {
		nameLowerCase := strings.ToLower(document.GetNodeText(node.Children[0]))
		if nameLowerCase == "\\define" || nameLowerCase == "define" {
			return newDefine(document, node)
		}
		scanForChildren(document, node)
	}
	return nil
}

func newFunctionCall(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	functionCall := &FunctionCall{
		Expression: Expression{},
	}
	document.addSymbol(functionCall)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	firstChild := child
	if firstChild != nil {
		functionCall.Location = document.GetNodeLocation(node.Children[0])
		functionCall.Name = document.GetNodeText(node.Children[0])
	}
	var open *lexer.Token = nil
	var close *lexer.Token = nil
	hasArgs := false
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ArgumentExpressionList {
				hasArgs = true
				break
			}
		} else if t, ok := child.(*lexer.Token); ok {
			switch t.Type {
			case lexer.OpenParenthesis:
				open = t
			case lexer.CloseParenthesis:
				close = t
			}
		}
		child = traverser.Advance()
	}
	if !hasArgs {
		args := newEmptyArgumentList(document, open, close)
		document.addSymbol(args)
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

func (s *FunctionCall) Serialise(e *storage.Encoder) {
	s.Expression.Serialise(e)
}

func ReadFunctionCall(d *storage.Decoder) *FunctionCall {
	return &FunctionCall{
		Expression: ReadExpression(d),
	}
}
