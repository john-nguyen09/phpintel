package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
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
func newConstantAccess(document *Document, node *phrase.Phrase) hasTypes {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	constantAccess.readName(document, node)
	return constantAccess
}

func (s *ConstantAccess) getLocation() lsp.Location {
	return s.Location
}

func (s *ConstantAccess) readName(document *Document, node phrase.AstNode) {
	s.Name = util.GetNodeText(node, document.text)
}

func (s *ConstantAccess) getTypes() TypeComposite {
	// TODO: look up constant type
	return s.Type
}

func (s *ConstantAccess) Serialise(serialiser *indexer.Serialiser) {
	s.Expression.Serialise(serialiser)
}

func ReadConstantAccess(serialiser *indexer.Serialiser) *ConstantAccess {
	return &ConstantAccess{
		Expression: ReadExpression(serialiser),
	}
}
