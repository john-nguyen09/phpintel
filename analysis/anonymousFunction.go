package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type AnonymousFunction struct {
	location protocol.Location
	children []Symbol

	Params []*Parameter
}

var _ BlockSymbol = (*AnonymousFunction)(nil)

func newAnonymousFunction(a analyser, document *Document, node *phrase.Phrase) Symbol {
	anonFunc := &AnonymousFunction{
		location: document.GetNodeLocation(node),
	}
	document.pushVariableTable(node)
	document.pushBlock(anonFunc)
	variableTable := document.getCurrentVariableTable()
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.AnonymousFunctionHeader:
				anonFunc.analyseHeader(a, document, p)
				for _, param := range anonFunc.Params {
					lastToken := util.LastToken(p)
					variableTable.add(a, param.ToVariable(), document.positionAt(lastToken.Offset+lastToken.Length))
				}
			case phrase.FunctionDeclarationBody:
				scanForChildren(a, document, p)
			}
		}
		child = traverser.Advance()
	}
	document.popVariableTable()
	document.popBlock()
	return anonFunc
}

func (s *AnonymousFunction) analyseHeader(a analyser, document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				s.analyseParameterDeclarationList(a, document, p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *AnonymousFunction) analyseParameterDeclarationList(a analyser, document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ParameterDeclaration {
			param := newParameter(a, document, p)
			s.Params = append(s.Params, param)
		}
		child = traverser.Advance()
	}
}

func (s *AnonymousFunction) GetLocation() protocol.Location {
	return s.location
}

func (s *AnonymousFunction) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *AnonymousFunction) GetChildren() []Symbol {
	return s.children
}
