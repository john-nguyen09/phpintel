package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Const contains information of constants
type Const struct {
	location lsp.Location

	Name  string
	Value string
}

func newConstDeclaration(document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ConstElementList {
			scanForChildren(document, p)
		}
		child = traverser.Advance()
	}

	return nil
}

func newConst(document *Document, node *phrase.Phrase) Symbol {
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
					traverser.SkipToken(lexer.Whitespace)
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

func (s *Const) getLocation() lsp.Location {
	return s.location
}
