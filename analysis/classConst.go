package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type ClassConst struct {
	location lsp.Location

	Name  TypeString
	Value string
	Scope TypeString
}

func ClassConstElementList(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	ScanForChildren(parent, node)

	return nil
}

func ClassConstElement(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	ScanForChildren(parent, node)

	return nil
}

// NewClassConstDeclaration is a proxy to NewClassConst due to the Parse Tree structure
func NewClassConstDeclaration(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	ScanForChildren(parent, node)

	return nil
}

func NewClassConst(document *Document, parent SymbolBlock, node *phrase.Phrase) Symbol {
	classConst := &ClassConst{
		location: document.GetNodeLocation(node),
	}

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
					classConst.Value += util.GetNodeText(token, document.GetText())
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			if hasEquals {
				classConst.Value += util.GetNodeText(p, document.GetText())
			} else {
				switch p.Type {
				case phrase.Identifier:
					{
						classConst.Name = NewTypeString(util.GetNodeText(p, document.GetText()))
					}
				}
			}
		}

		child = traverser.Advance()
	}

	return classConst
}

func (s *ClassConst) GetLocation() lsp.Location {
	return s.location
}
