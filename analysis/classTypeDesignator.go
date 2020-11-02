package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ClassTypeDesignator represents a reference to object creation (e.g. new TestClass())
type ClassTypeDesignator struct {
	Expression

	children []Symbol
}

var _ BlockSymbol = (*ClassTypeDesignator)(nil)

func newClassTypeDesignator(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	classTypeDesignator := &ClassTypeDesignator{}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ClassTypeDesignator:
				document.addSymbol(classTypeDesignator)
				classTypeDesignator.analyseNode(a, document, p)
			case phrase.ArgumentExpressionList:
				newArgumentList(a, document, p)
			case phrase.AnonymousClassDeclaration:
				classTypeDesignator.Location = document.GetNodeLocation(p)
				document.addSymbol(classTypeDesignator)
				document.pushBlock(classTypeDesignator)
				scanNode(a, document, p)
				document.popBlock()
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

func (s *ClassTypeDesignator) referenceClass(class *Class) {
	s.Name = class.Name.GetOriginal()
	s.Type.add(class.Name)
}

func (s *ClassTypeDesignator) GetLocation() protocol.Location {
	return s.Location
}

func (s *ClassTypeDesignator) GetTypes() TypeComposite {
	return s.Type
}

func (s *ClassTypeDesignator) ResolveToHasParams(ctx ResolveContext) []HasParams {
	hasParams := []HasParams{}
	q := ctx.query
	for _, typeString := range s.GetTypes().Resolve() {
		for _, class := range q.GetClasses(typeString.GetFQN()) {
			constructor := q.GetClassConstructor(class)
			if constructor.Method != nil {
				hasParams = append(hasParams, constructor.Method)
			}
		}
	}
	return hasParams
}

func (s *ClassTypeDesignator) GetChildren() []Symbol {
	return s.children
}

func (s *ClassTypeDesignator) addChild(child Symbol) {
	if class, ok := child.(*Class); ok {
		s.referenceClass(class)
	}
	s.children = append(s.children, child)
}
