package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// ClassConst contains information of class constants
type ClassConst struct {
	location protocol.Location

	Name  string
	Value string
	Scope TypeString
}

func newClassConst(document *Document, node *phrase.Phrase) Symbol {
	classConst := &ClassConst{
		location: document.GetNodeLocation(node),
	}

	parent := document.getLastClass()
	if theClass, ok := parent.(*Class); ok {
		classConst.Scope = theClass.Name
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEquals := false
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Equals:
				{
					hasEquals = true
					traverser.SkipToken(lexer.Whitespace)
				}
			default:
				if hasEquals {
					classConst.Value += document.GetTokenText(token)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				classConst.Value += document.GetPhraseText(p)
			} else {
				switch p.Type {
				case phrase.Identifier:
					{
						classConst.Name = document.GetPhraseText(p)
					}
				}
			}
		}

		child = traverser.Advance()
	}

	return classConst
}

func (s *ClassConst) getLocation() protocol.Location {
	return s.location
}

func (s *ClassConst) GetCollection() string {
	return classConstCollection
}

func (s *ClassConst) GetKey() string {
	return s.Scope.fqn + KeySep + s.Name
}

func (s *ClassConst) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteString(s.Name)
	serialiser.WriteString(s.Value)
	s.Scope.Write(serialiser)
}

func ReadClassConst(serialiser *Serialiser) *ClassConst {
	return &ClassConst{
		location: serialiser.ReadLocation(),
		Name:     serialiser.ReadString(),
		Value:    serialiser.ReadString(),
		Scope:    ReadTypeString(serialiser),
	}
}
