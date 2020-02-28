package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// ClassTypeDesignator represents a reference to object creation (e.g. new TestClass())
type ClassTypeDesignator struct {
	Expression
}

func newClassTypeDesignator(document *Document, node *sitter.Node) (HasTypes, bool) {
	s := &ClassTypeDesignator{}
	document.addSymbol(s)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			typeString := transformQualifiedName(child, document)
			typeString.SetFQN(document.currImportTable().GetClassReferenceFQN(typeString))
			s.Location = document.GetNodeLocation(child)
			s.Name = typeString.GetOriginal()
			s.Type.add(typeString)
		case "relative_scope":
			relativeScope := newRelativeScope(document, s.Location)
			s.Type.merge(relativeScope.Types)
		case "variable_name":
			if variable, ok := newVariable(document, child); ok {
				document.addSymbol(variable)
			}
		case "arguments":
			newArgumentList(document, child)
		}
		child = traverser.Advance()
	}
	return s, false
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
