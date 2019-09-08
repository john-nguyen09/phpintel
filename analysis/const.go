package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Const struct {
	location lsp.Location

	Name  string
	Value string
}

func NewConstDeclaration(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ConstElementList {
			ScanForChildren(parent, p)
		}
		child = traverser.Advance()
	}

	return nil
}

func NewConst(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
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
				{
					constant.Name = util.GetNodeText(token, document.GetText())
				}
			case lexer.Equals:
				{
					hasEquals = true
					next := traverser.Peek()
					for nextToken, ok := next.(*lexer.Token); ok && nextToken.Type == lexer.Whitespace; {
						traverser.Advance()
						next = traverser.Peek()
						nextToken, ok = next.(*lexer.Token)
					}
				}
			default:
				{
					if hasEquals {
						constant.Value += util.GetNodeText(token, document.GetText())
					}
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				constant.Value += util.GetNodeText(p, document.GetText())
			}
		}

		child = traverser.Advance()
	}

	return constant
}

func (s *Const) GetLocation() lsp.Location {
	return s.location
}
