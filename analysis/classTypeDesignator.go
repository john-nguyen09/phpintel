package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ClassTypeDesignator represents a reference to object creation (e.g. new TestClass())
type ClassTypeDesignator struct {
	Expression
}

func newClassTypeDesignator(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	classTypeDesignator := &ClassTypeDesignator{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName:
				typeString := transformQualifiedName(p, document)
				typeString.SetFQN(document.GetImportTable().GetClassReferenceFQN(typeString))
				classTypeDesignator.Name = typeString.GetOriginal()
				classTypeDesignator.Type.add(typeString)
			case phrase.RelativeScope:
				relativeScope := newRelativeScope(document, classTypeDesignator.Location)
				classTypeDesignator.Type.merge(relativeScope.Types)
			}
		}
		child = traverser.Advance()
	}
	return classTypeDesignator, true
}

func (s *ClassTypeDesignator) GetLocation() protocol.Location {
	return s.Location
}

func (s *ClassTypeDesignator) Resolve(store *Store) {

}

func (s *ClassTypeDesignator) GetTypes() TypeComposite {
	return s.Type
}

func (s *ClassTypeDesignator) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadClassTypeDesignator(serialiser *Serialiser) *ClassTypeDesignator {
	return &ClassTypeDesignator{
		Expression: ReadExpression(serialiser),
	}
}
