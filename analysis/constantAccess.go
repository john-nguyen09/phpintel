package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// ConstantAccess represents a reference to constant access
type ConstantAccess struct {
	Expression
}

func newDirectoryConstantAccess(document *Document, token *lexer.Token) Symbol {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(token),
		},
	}
	constantAccess.readName(document, token)
	return constantAccess
}
func newConstantAccess(document *Document, node *phrase.Phrase) HasTypes {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	constantAccess.readName(document, node)
	return constantAccess
}

func (s *ConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ConstantAccess) readName(document *Document, node phrase.AstNode) {
	s.Name = document.GetNodeText(node)
	s.Type.add(newTypeString(s.Name))
}

func (s *ConstantAccess) Resolve(store *Store) {

}

func (s *ConstantAccess) GetTypes() TypeComposite {
	// TODO: look up constant type
	return s.Type
}

func (s *ConstantAccess) Serialise(serialiser *Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadConstantAccess(serialiser *Serialiser) *ConstantAccess {
	return &ConstantAccess{
		Expression: ReadExpression(serialiser),
	}
}
