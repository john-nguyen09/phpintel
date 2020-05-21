package analysis

import (
	"log"

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
	prevVariableTable := document.getCurrentVariableTable()
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
				anonFunc.analyseHeader(a, document, p, variableTable, prevVariableTable)
				for _, param := range anonFunc.Params {
					lastToken := util.LastToken(p)
					variableTable.add(a, param.ToVariable(), document.positionAt(lastToken.Offset+lastToken.Length), true)
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

func (s *AnonymousFunction) analyseHeader(a analyser, document *Document, node *phrase.Phrase,
	variableTable *VariableTable, prevVariableTable *VariableTable) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				s.analyseParameterDeclarationList(a, document, p)
			case phrase.AnonymousFunctionUseClause:
				s.analyseUseClause(a, document, p, variableTable, prevVariableTable)
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

func (s *AnonymousFunction) analyseUseClause(a analyser, document *Document, node *phrase.Phrase,
	variableTable *VariableTable, prevVariableTable *VariableTable) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ClosureUseList {
				var err error
				traverser, err = traverser.Descend()
				if err != nil {
					log.Println(err)
				} else {
					child = traverser.Advance()
					for child != nil {
						if p, ok := child.(*phrase.Phrase); ok {
							if p.Type == phrase.AnonymousFunctionUseVariable {
								variable, shouldAdd := newVariable(a, document, p, true)
								prevVariableTable.add(a, variable, variable.GetLocation().Range.End, false)
								if shouldAdd {
									document.addSymbol(variable)
								}
								prevVariable := prevVariableTable.get(variable.Name, variable.Location.Range.Start)
								if prevVariable != nil {
									variable.mergeTypesWithVariable(prevVariable)
								}
							}
						}
						child = traverser.Advance()
					}
				}
				traverser, err = traverser.Ascend()
				if err != nil {
					log.Println(err)
				}
			}
		}
		traverser.Advance()
		child = traverser.Peek()
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
