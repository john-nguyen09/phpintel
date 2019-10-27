package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Interface contains information of interfaces
type Interface struct {
	document *Document
	location lsp.Location

	Name    TypeString
	Extends []TypeString
}

func newInterface(document *Document, node *phrase.Phrase) Symbol {
	theInterface := &Interface{
		document: document,
		location: document.GetNodeLocation(node),
	}
	document.addClass(theInterface)
	if interfaceHeader, ok := node.Children[0].(*phrase.Phrase); ok && interfaceHeader.Type == phrase.InterfaceDeclarationHeader {
		theInterface.analyseHeader(interfaceHeader)
	}
	if len(node.Children) >= 2 {
		if interfaceBody, ok := node.Children[1].(*phrase.Phrase); ok {
			scanForChildren(document, interfaceBody)
		}
	}

	return theInterface
}

func (s *Interface) analyseHeader(node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = newTypeString(util.GetNodeText(token, s.document.text))
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceBaseClause:
				{
					s.extends(p)
				}
			}
		}

		child = traverser.Advance()
	}
}

func (s *Interface) extends(node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedNameList {
			traverser.Descend()
			child = traverser.Advance()
			for child != nil {
				if p, ok = child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
					s.Extends = append(s.Extends, transformQualifiedName(p, s.document))
				}

				child = traverser.Advance()
			}

			break
		}

		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Interface) getLocation() lsp.Location {
	return s.location
}

func (s *Interface) getDocument() *Document {
	return s.document
}

func (s *Interface) Serialise() []byte {
	serialiser := NewSerialiser()
	serialiser.WriteLocation(s.location)
	s.Name.Write(serialiser)
	serialiser.WriteInt(len(s.Extends))
	for _, extend := range s.Extends {
		extend.Write(serialiser)
	}
	return serialiser.GetBytes()
}

func DeserialiseInterface(document *Document, bytes []byte) *Interface {
	serialiser := SerialiserFromByteSlice(bytes)
	theInterface := &Interface{
		document: document,
		location: serialiser.ReadLocation(),
		Name:     ReadTypeString(serialiser),
	}
	countExtends := serialiser.ReadInt()
	for i := 0; i < countExtends; i++ {
		theInterface.Extends = append(theInterface.Extends, ReadTypeString(serialiser))
	}
	return theInterface
}
