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

func newDirectoryConstantAccess(a analyser, document *Document, token *lexer.Token) Symbol {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(token),
		},
	}
	constantAccess.readName(document, token)
	return constantAccess
}

func newConstantAccess(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	constantAccess := &ConstantAccess{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	constantAccess.readName(document, node)
	return constantAccess, true
}

func (s *ConstantAccess) GetLocation() protocol.Location {
	return s.Location
}

func (s *ConstantAccess) readName(document *Document, node phrase.AstNode) {
	s.Name = document.GetNodeText(node)
}

func (s *ConstantAccess) GetTypes() TypeComposite {
	// TODO: look up constant type
	return s.Type
}
