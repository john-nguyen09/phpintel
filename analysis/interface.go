package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Interface contains information of interfaces
type Interface struct {
	location    protocol.Location
	description string

	Name    TypeString
	Extends []TypeString
}

func newInterface(document *Document, node *phrase.Phrase) Symbol {
	theInterface := &Interface{
		location: document.GetNodeLocation(node),
	}
	document.addClass(theInterface)
	document.addSymbol(theInterface)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceDeclarationHeader:
				theInterface.analyseHeader(document, p)
			case phrase.InterfaceDeclarationBody:
				scanForChildren(document, p)
			}
		}
		child = traverser.Advance()
	}
	theInterface.Name.SetNamespace(document.importTable.namespace)
	return nil
}

func (s *Interface) analyseHeader(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = NewTypeString(document.GetTokenText(token))
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.InterfaceBaseClause:
				{
					s.extends(document, p)
				}
			}
		}

		child = traverser.Advance()
	}
}

func (s *Interface) extends(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedNameList {
			traverser, _ = traverser.Descend()
			child = traverser.Advance()
			for child != nil {
				if p, ok = child.(*phrase.Phrase); ok && p.Type == phrase.QualifiedName {
					s.Extends = append(s.Extends, transformQualifiedName(p, document))
				}

				child = traverser.Advance()
			}

			break
		}

		traverser.Advance()
		child = traverser.Peek()
	}
}

func (s *Interface) GetLocation() protocol.Location {
	return s.location
}

func (s *Interface) GetDescription() string {
	return s.description
}

func (s *Interface) GetCollection() string {
	return interfaceCollection
}

func (s *Interface) GetKey() string {
	return s.Name.GetFQN() + KeySep + s.location.URI
}

func (s *Interface) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Interface) GetIndexCollection() string {
	return interfaceCompletionIndex
}

func (s *Interface) GetPrefixes() []string {
	scope, _ := GetScopeAndNameFromString(s.Name.GetFQN())
	prefixes := []string{""}
	if scope != "" {
		prefixes = append(prefixes, scope)
	}
	return prefixes
}

func (s *Interface) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	s.Name.Write(serialiser)
	serialiser.WriteInt(len(s.Extends))
	for _, extend := range s.Extends {
		extend.Write(serialiser)
	}
}

func ReadInterface(serialiser *Serialiser) *Interface {
	theInterface := &Interface{
		location: serialiser.ReadLocation(),
		Name:     ReadTypeString(serialiser),
	}
	countExtends := serialiser.ReadInt()
	for i := 0; i < countExtends; i++ {
		theInterface.Extends = append(theInterface.Extends, ReadTypeString(serialiser))
	}
	return theInterface
}
