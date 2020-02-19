package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Function contains information of functions
type Function struct {
	location protocol.Location

	Name        TypeString `json:"Name"`
	Params      []*Parameter
	returnTypes TypeComposite
	description string
}

func newFunction(document *Document, node *phrase.Phrase) Symbol {
	function := &Function{
		location:    document.GetNodeLocation(node),
		Params:      make([]*Parameter, 0),
		returnTypes: newTypeComposite(),
	}
	phpDoc := document.getValidPhpDoc(function.location)
	document.pushVariableTable(node)

	variableTable := document.getCurrentVariableTable()
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationHeader,
			phrase.MethodDeclarationHeader,
		}); ok {
			function.analyseHeader(document, p)
			if phpDoc != nil {
				function.applyPhpDoc(document, *phpDoc)
			}
			for _, param := range function.Params {
				variableTable.add(param.ToVariable())
			}
		}
		if p, ok := util.IsOfPhraseTypes(child, []phrase.PhraseType{
			phrase.FunctionDeclarationBody,
			phrase.MethodDeclarationBody,
		}); ok {
			scanForChildren(document, p)
		}
		child = traverser.Advance()
	}
	function.Name.SetNamespace(document.importTable.namespace)
	document.popVariableTable()
	return function
}

func (s *Function) analyseHeader(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if token, ok := child.(*lexer.Token); ok {
			switch token.Type {
			case lexer.Name:
				{
					s.Name = NewTypeString(document.GetTokenText(token))
				}
			}
		} else if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclarationList:
				{
					s.analyseParameterDeclarationList(document, p)
				}
			case phrase.Identifier:
				s.Name = NewTypeString(document.GetPhraseText(p))
			}
		}
		child = traverser.Advance()
	}
}

func (s *Function) analyseParameterDeclarationList(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.ParameterDeclaration {
			param := newParameter(document, p)
			s.Params = append(s.Params, param)
		}

		child = traverser.Advance()
	}
}

func (s *Function) applyPhpDoc(document *Document, phpDoc phpDocComment) {
	tags := phpDoc.Returns
	for _, tag := range tags {
		s.returnTypes.merge(typesFromPhpDoc(document, tag.TypeString))
	}
	for index, param := range s.Params {
		tag := phpDoc.findParamTag(param.Name)
		if tag != nil {
			s.Params[index].Type.merge(typesFromPhpDoc(document, tag.TypeString))
			s.Params[index].description = tag.Description
		}
	}
	s.description = phpDoc.Description
}

func (s *Function) GetLocation() protocol.Location {
	return s.location
}

func (s *Function) GetName() TypeString {
	return s.Name
}

func (s *Function) GetDescription() string {
	return s.description
}

func (s *Function) GetDetail() string {
	return s.returnTypes.ToString()
}

func (s *Function) GetReturnTypes() TypeComposite {
	return s.returnTypes
}

func (s *Function) GetCollection() string {
	return functionCollection
}

func (s *Function) GetKey() string {
	return s.Name.GetFQN() + KeySep + s.location.URI
}

func (s *Function) GetIndexableName() string {
	return s.Name.GetOriginal()
}

func (s *Function) GetIndexCollection() string {
	return functionCompletionIndex
}

func (s *Function) GetScope() string {
	return s.Name.GetNamespace()
}

func (s *Function) GetNameLabel() string {
	return s.Name.GetOriginal()
}

func (s *Function) GetParams() []*Parameter {
	return s.Params
}

func (s *Function) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.Name.Write(e)
	e.WriteInt(len(s.Params))
	for _, param := range s.Params {
		param.Write(e)
	}
	s.returnTypes.Write(e)
	e.WriteString(s.description)
}

func ReadFunction(d *storage.Decoder) *Function {
	function := Function{
		location: d.ReadLocation(),
		Name:     ReadTypeString(d),
		Params:   make([]*Parameter, 0),
	}
	countParams := d.ReadInt()
	for i := 0; i < countParams; i++ {
		function.Params = append(function.Params, ReadParameter(d))
	}
	function.returnTypes = ReadTypeComposite(d)
	function.description = d.ReadString()
	return &function
}
