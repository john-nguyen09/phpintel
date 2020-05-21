package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type TraitAccess struct {
	Expression
}

var _ HasTypes = (*TraitAccess)(nil)

func processTraitUseClause(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedNameList:
				traitAnalyseQualifiedNameList(document, p)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func traitAnalyseQualifiedNameList(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	currentClass := document.getLastClass()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				traitAccess := newTraitAccess(document, p)
				document.addSymbol(traitAccess)
				if class, ok := currentClass.(*Class); ok {
					for _, typeString := range traitAccess.Type.Resolve() {
						class.AddUse(typeString)
					}
				}
			}
		}
		child = traverser.Advance()
	}
}

func newTraitAccess(document *Document, node *phrase.Phrase) *TraitAccess {
	traitAccess := &TraitAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.getPhraseText(node),
		},
	}
	types := newTypeComposite()
	if node.Type == phrase.QualifiedName || node.Type == phrase.FullyQualifiedName {
		typeString := transformQualifiedName(node, document)
		typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
		types.add(typeString)
	}
	traitAccess.Type = types
	return traitAccess
}

func (s *TraitAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *TraitAccess) GetTypes() TypeComposite {
	return s.Type
}
