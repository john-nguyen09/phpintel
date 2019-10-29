package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Define contains information of define constants
type Define struct {
	document *Document
	location lsp.Location

	Name  TypeString
	Value string
}

func newDefine(document *Document, node *phrase.Phrase) Symbol {
	define := &Define{
		document: document,
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ArgumentExpressionList {
				symbol := newArgumentList(document, p)
				if args, ok := symbol.(*ArgumentList); ok {
					define.analyseArgs(args)
				}
			}
		}
		child = traverser.Advance()
	}

	return define
}

func (s *Define) getLocation() lsp.Location {
	return s.location
}

func (s *Define) getDocument() *Document {
	return s.document
}

func (s *Define) analyseArgs(args *ArgumentList) {
	firstArg := args.GetArguments()[0]
	if token, ok := firstArg.(*lexer.Token); ok {
		if token.Type == lexer.StringLiteral {
			stringText := util.GetTokenText(token, s.getDocument().GetText())
			s.Name = newTypeString(stringText[1 : len(stringText)-1])
		}
	}
	if len(args.GetArguments()) >= 2 {
		secondArg := args.GetArguments()[1]
		s.Value = util.GetNodeText(secondArg, s.getDocument().GetText())
	}
}

func (s *Define) GetCollection() string {
	return "define"
}

func (s *Define) GetKey() string {
	return s.Name.fqn + KeySep + s.document.GetURI()
}

func (s *Define) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	s.Name.Write(serialiser)
	serialiser.WriteString(s.Value)
}

func ReadDefine(serialiser *Serialiser) *Define {
	return &Define{
		location: serialiser.ReadLocation(),
		Name:     ReadTypeString(serialiser),
		Value:    serialiser.ReadString(),
	}
}
