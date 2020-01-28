package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type AnonymousFunction struct {
	location protocol.Location

	Params []*Parameter
}

func newAnonymousFunction(document *Document, node *phrase.Phrase) Symbol {
	anonFunc := &AnonymousFunction{
		location: document.GetNodeLocation(node),
	}
	document.pushVariableTable(node)
	variableTable := document.getCurrentVariableTable()
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.AnonymousFunctionHeader:
				anonFunc.analyseHeader(document, p)
				for _, param := range anonFunc.Params {
					variableTable.add(param.ToVariable())
				}
				document.addSymbol(anonFunc)
			case phrase.FunctionDeclarationBody:
				scanForChildren(document, p)
			}
		}
		child = traverser.Advance()
	}
	document.popVariableTable()
	return nil
}

func (s *AnonymousFunction) GetLocation() protocol.Location {
	return s.location
}

func (s *AnonymousFunction) analyseHeader(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				s.analyseParameterDeclarationList(document, p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *AnonymousFunction) analyseParameterDeclarationList(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ParameterDeclaration {
			param := newParameter(document, p)
			s.Params = append(s.Params, param)
		}
		child = traverser.Advance()
	}
}
