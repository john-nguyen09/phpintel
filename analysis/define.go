package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Define contains information of define constants
type Define struct {
	location      protocol.Location
	description   string
	children      []Symbol
	deprecatedTag *tag
	Name          TypeString
	Value         string
}

var _ HasTypes = (*Define)(nil)
var _ HasParamsResolvable = (*Define)(nil)

func newDefine(a analyser, document *Document, node *phrase.Phrase) Symbol {
	define := &Define{
		location: document.GetNodeLocation(node),
	}
	phpDoc := document.getValidPhpDoc(define.location)
	document.addSymbol(define)
	document.pushBlock(define)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.ArgumentExpressionList {
				symbol := newArgumentList(a, document, p)
				if args, ok := symbol.(*ArgumentList); ok {
					define.analyseArgs(document, args)
					if phpDoc != nil {
						define.description = phpDoc.Description
						define.deprecatedTag = phpDoc.deprecated()
					}
				}
			}
		}
		child = traverser.Advance()
	}
	define.Name.SetNamespace(document.currImportTable().GetNamespace())
	document.popBlock()
	return nil
}

func (s *Define) GetLocation() protocol.Location {
	return s.location
}

func (s *Define) GetName() string {
	return s.Name.GetFQN()
}

func (s *Define) GetDescription() string {
	return s.GetName() + " = " + s.Value + "; " + s.description
}

func (s *Define) analyseArgs(document *Document, args *ArgumentList) {
	if len(args.GetArguments()) == 0 {
		return
	}
	firstArg := args.GetArguments()[0]
	if token, ok := firstArg.(*lexer.Token); ok {
		if token.Type == lexer.StringLiteral {
			stringText := document.getTokenText(token)
			s.Name = NewTypeString(stringText[1 : len(stringText)-1])
		}
	}
	if len(args.GetArguments()) >= 2 {
		secondArg := args.GetArguments()[1]
		s.Value = document.GetNodeText(secondArg)
	}
}

func (s *Define) addChild(child Symbol) {
	s.children = append(s.children, child)
}

func (s *Define) GetChildren() []Symbol {
	return s.children
}

func (s *Define) GetTypes() TypeComposite {
	return newTypeComposite()
}

func (s *Define) Resolve(ctx ResolveContext) {

}

func (s *Define) ResolveToHasParams(ctx ResolveContext) []HasParams {
	functions := []HasParams{}
	typeString := NewTypeString("\\define")
	q := ctx.query
	document := ctx.document
	typeString.SetFQN(document.currImportTable().GetFunctionReferenceFQN(q, typeString))
	for _, function := range q.GetFunctions(typeString.GetFQN()) {
		functions = append(functions, function)
	}
	return functions
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

func (s *Define) Serialise(e storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.description)
	serialiseDeprecatedTag(e, s.deprecatedTag)
	s.Name.Write(e)
	e.WriteString(s.Value)
}

func ReadDefine(d storage.Decoder) *Define {
	return &Define{
		location:      d.ReadLocation(),
		description:   d.ReadString(),
		deprecatedTag: deserialiseDeprecatedTag(d),
		Name:          ReadTypeString(d),
		Value:         d.ReadString(),
	}
}
