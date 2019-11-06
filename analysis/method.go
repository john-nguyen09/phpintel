package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Method contains information of methods
type Method struct {
	Function

	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
	ClassModifier      ClassModifierValue
}

func newMethod(document *Document, node *phrase.Phrase) Symbol {
	symbol := newFunction(document, node)
	method := &Method{
		IsStatic: false,
	}

	if function, ok := symbol.(*Function); ok {
		method.Function = *function
	}
	method.location = document.GetNodeLocation(node)
	lastClass := document.getLastClass()
	if theClass, ok := lastClass.(*Class); ok {
		method.Scope = theClass.Name
	} else if theInterface, ok := lastClass.(*Interface); ok {
		method.Scope = theInterface.Name
	} else if trait, ok := lastClass.(*Trait); ok {
		method.Scope = trait.Name
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseType(child, phrase.MethodDeclarationHeader); ok {
			method.analyseHeader(p)
		}
		child = traverser.Advance()
	}

	return method
}

func (s Method) GetLocation() protocol.Location {
	return s.location
}

func (s *Method) analyseHeader(methodHeader *phrase.Phrase) {
	traverser := util.NewTraverser(methodHeader)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.MemberModifierList:
				s.VisibilityModifier, s.IsStatic, s.ClassModifier = getMemberModifier(p)
			}
		}
		child = traverser.Advance()
	}
}

func (s *Method) GetCollection() string {
	return methodCollection
}

func (s *Method) GetKey() string {
	return s.Scope.fqn + KeySep + s.Name + KeySep + s.location.URI
}

func (s *Method) Serialise(serialiser *Serialiser) {
	s.Function.Serialise(serialiser)
	s.Scope.Write(serialiser)
	serialiser.WriteInt(int(s.VisibilityModifier))
	serialiser.WriteBool(s.IsStatic)
	serialiser.WriteInt(int(s.ClassModifier))
}

func ReadMethod(serialiser *Serialiser) *Method {
	return &Method{
		Function:           *ReadFunction(serialiser),
		Scope:              ReadTypeString(serialiser),
		VisibilityModifier: VisibilityModifierValue(serialiser.ReadInt()),
		IsStatic:           serialiser.ReadBool(),
		ClassModifier:      ClassModifierValue(serialiser.ReadInt()),
	}
}
