package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Define contains information of define constants
type Define struct {
	location protocol.Location

	Name  TypeString
	Value string
}

func newDefine(document *Document, node *sitter.Node) Symbol {
	define := &Define{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "arguments" {
			symbol := newArgumentList(document, child)
			if args, ok := symbol.(*ArgumentList); ok {
				define.analyseArgs(document, args)
			}
		}
		child = traverser.Advance()
	}
	define.Name.SetNamespace(document.currImportTable().GetNamespace())
	return define
}

func (s *Define) GetLocation() protocol.Location {
	return s.location
}

func (s *Define) GetName() string {
	return s.Name.GetFQN()
}

func (s *Define) GetDescription() string {
	return s.GetName() + " = " + s.Value
}

func (s *Define) analyseArgs(document *Document, args *ArgumentList) {
	firstArg := args.GetArguments()[0]
	if firstArg.Type() == "string" {
		stringText := document.GetNodeText(firstArg)
		s.Name = NewTypeString(stringText[1 : len(stringText)-1])
	}
	if len(args.GetArguments()) >= 2 {
		secondArg := args.GetArguments()[1]
		s.Value = document.GetNodeText(secondArg)
	}
}

func (s *Define) GetCollection() string {
	return defineCollection
}

func (s *Define) GetKey() string {
	return s.GetName() + KeySep + s.location.URI
}

func (s *Define) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Define) GetIndexCollection() string {
	return defineCompletionIndex
}

func (s *Define) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteString(s.Value)
}

func ReadDefine(d *storage.Decoder) *Define {
	return &Define{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
		Value:    d.ReadString(),
	}
}
