package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Method struct {
	location lsp.Location

	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
	ClassModifier      ClassModifierValue
	Function           Function
}

func NewMethod(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	symbol := NewFunction(document, parent, node)
	method := &Method{
		location: document.GetNodeLocation(node),
		IsStatic: false,
	}

	if function, ok := symbol.(*Function); ok {
		method.Function = *function
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseType(child, phrase.MethodDeclarationHeader); ok {
			method.analyseHeader(p)
		}
		child = traverser.Advance()
	}

	return method
}

func (s Method) GetLocation() lsp.Location {
	return s.location
}

func (s *Method) analyseHeader(methodHeader *phrase.Phrase) {
	traverser := util.NewTraverser(methodHeader)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.MemberModifierList:
				s.VisibilityModifier, s.IsStatic, s.ClassModifier = GetMemberModifier(p)
			}
		}
		child = traverser.Advance()
	}
}
