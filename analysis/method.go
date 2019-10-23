package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Method contains information of methods
type Method struct {
	Function

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

func (s Method) getLocation() lsp.Location {
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

func (s *Method) Serialise() []byte {
	serialiser := indexer.NewSerialiser()
	s.Function.Write(serialiser)
	serialiser.WriteInt(int(s.VisibilityModifier))
	serialiser.WriteBool(s.IsStatic)
	serialiser.WriteInt(int(s.ClassModifier))
	return serialiser.GetBytes()
}

func DeserialiseMethod(document *Document, bytes []byte) *Method {
	serialiser := indexer.SerialiserFromByteSlice(bytes)
	return &Method{
		Function:           ReadFunction(serialiser),
		VisibilityModifier: VisibilityModifierValue(serialiser.ReadInt()),
		IsStatic:           serialiser.ReadBool(),
		ClassModifier:      ClassModifierValue(serialiser.ReadInt()),
	}
}
