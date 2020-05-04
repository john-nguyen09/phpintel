package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
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

var _ HasScope = (*ClassConst)(nil)
var _ Symbol = (*ClassConst)(nil)

func newClassConst(document *Document, node *phrase.Phrase) Symbol {
	classConst := &ClassConst{
		location: document.GetNodeLocation(node),
	}
	lastClass := document.getLastClass()
	if theClass, ok := lastClass.(*Class); ok {
		classConst.Scope = theClass.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if theInterface, ok := lastClass.(*Interface); ok {
		classConst.Scope = theInterface.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if trait, ok := lastClass.(*Trait); ok {
		classConst.Scope = trait.Name
		classConst.Scope.SetNamespace(document.currImportTable().GetNamespace())
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
					classConst.Value += document.getTokenText(token)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				classConst.Value += document.getPhraseText(p)
			} else {
				switch p.Type {
				case phrase.Identifier:
					{
						classConst.Name = document.getPhraseText(p)
					}
				}
			}
		}
		child = traverser.Advance()
	}
	return classConst
}

func (s *ClassConst) GetLocation() protocol.Location {
	return s.location
}

func (s *ClassConst) GetName() string {
	return s.Name
}

func (s *ClassConst) GetDescription() string {
	// TODO: Implement docblock description
	return ""
}

func (s *ClassConst) GetCollection() string {
	return classConstCollection
}

func (s *ClassConst) GetKey() string {
	return s.Scope.fqn + KeySep + s.Name
}

func (s *ClassConst) GetIndexableName() string {
	return s.GetName()
}

func (s *ClassConst) GetIndexCollection() string {
	return classConstCompletionIndex
}

func (s *ClassConst) GetScope() string {
	return s.Scope.GetFQN()
}

func (s *ClassConst) IsScopeSymbol() bool {
	return true
}

func (s *ClassConst) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.Name)
	e.WriteString(s.Value)
	s.Scope.Write(e)
}

func ReadClassConst(d *storage.Decoder) *ClassConst {
	return &ClassConst{
		location: d.ReadLocation(),
		Name:     d.ReadString(),
		Value:    d.ReadString(),
		Scope:    ReadTypeString(d),
	}
}
