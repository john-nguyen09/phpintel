package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type TraitAccess struct {
	Expression
}

func processTraitUseClause(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	currentClass := document.getLastClass()
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			traitAccess := newTraitAccess(document, child)
			document.addSymbol(traitAccess)
			if class, ok := currentClass.(*Class); ok {
				for _, typeString := range traitAccess.Type.Resolve() {
					class.AddUse(typeString)
				}
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newTraitAccess(document *Document, node *sitter.Node) *TraitAccess {
	traitAccess := &TraitAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
			Name:     document.GetPhraseText(node),
		},
	}
	types := newTypeComposite()
	if node.Type() == "qualified_name" {
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
