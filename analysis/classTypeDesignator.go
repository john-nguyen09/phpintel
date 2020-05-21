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

func newClassTypeDesignator(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	classTypeDesignator := &ClassTypeDesignator{}
	document.addSymbol(classTypeDesignator)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassTypeDesignator:
				classTypeDesignator.analyseNode(a, document, p)
			case phrase.ArgumentExpressionList:
				newArgumentList(a, document, p)
			}
		}
		child = traverser.Advance()
	}
	return classTypeDesignator, false
}

func (s *ClassTypeDesignator) analyseNode(a analyser, document *Document, node *phrase.Phrase) {
	s.Location = document.GetNodeLocation(node)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				typeString := transformQualifiedName(p, document)
				typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
				s.Name = typeString.GetOriginal()
				s.Type.add(typeString)
			case phrase.RelativeScope:
				relativeScope := newRelativeScope(document, s.Location)
				s.Type.merge(relativeScope.Types)
			case phrase.SimpleVariable:
				if variable, ok := newVariable(a, document, p, false); ok {
					document.addSymbol(variable)
				}
			}
		}
		child = traverser.Advance()
	}
}

func (s *ClassTypeDesignator) GetLocation() protocol.Location {
	return s.Location
}

func (s *ClassTypeDesignator) GetTypes() TypeComposite {
	return s.Type
}

func (s *ClassTypeDesignator) ResolveToHasParams(ctx ResolveContext) []HasParams {
	hasParams := []HasParams{}
	store := ctx.store
	for _, typeString := range s.GetTypes().Resolve() {
		methods := store.GetMethods(typeString.GetFQN(), "__construct")
		for _, method := range methods {
			hasParams = append(hasParams, method)
		}
	}
	return hasParams
}
