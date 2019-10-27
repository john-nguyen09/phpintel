package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

// ClassTypeDesignator represents a reference to object creation (e.g. new TestClass())
type ClassTypeDesignator struct {
	Expression
}

func newClassTypeDesignator(document *Document, node *phrase.Phrase) hasTypes {
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
				classTypeDesignator.Name = typeString.original
				classTypeDesignator.Type.add(typeString)
			}
		}
		child = traverser.Advance()
	}
	return classTypeDesignator
}

func (s *ClassTypeDesignator) getTypes() TypeComposite {
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
