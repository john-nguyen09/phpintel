package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Const contains information of constants
type Const struct {
	location protocol.Location

	Name  TypeString
	Value string
}

func newConstDeclaration(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ConstElementList {
			scanForChildren(a, document, p)
		}
		child = traverser.Advance()
	}
	return nil
}

func newConst(a analyser, document *Document, node *phrase.Phrase) Symbol {
	constant := &Const{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	hasEquals := false
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				constant.Name = NewTypeString(document.getTokenText(token))
			case lexer.Equals:
				hasEquals = true
				traverser.SkipToken(lexer.Whitespace)
			default:
				if hasEquals {
					constant.Value += document.getTokenText(token)
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				constant.Value += document.getPhraseText(p)
			}
		}

		child = traverser.Advance()
	}
	constant.Name.SetNamespace(document.currImportTable().GetNamespace())
	return constant
}

func (s *Const) GetLocation() protocol.Location {
	return s.location
}

func (s *Const) GetName() string {
	return s.Name.GetFQN()
}

func (s *Const) GetDescription() string {
	return s.GetName() + " = " + s.Value
}

func (s *Const) GetCollection() string {
	return constCollection
}

func (s *Const) GetKey() string {
	return s.GetName() + KeySep + s.location.URI
}

func (s *Const) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Const) GetIndexCollection() string {
	return constCompletionIndex
}

func (s *Const) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteString(s.Value)
}

func ReadConst(d *storage.Decoder) *Const {
	return &Const{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
		Value:    d.ReadString(),
	}
}
