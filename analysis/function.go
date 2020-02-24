package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Function contains information of functions
type Function struct {
	location protocol.Location

	Name        TypeString `json:"Name"`
	Params      []*Parameter
	returnTypes TypeComposite
	description string
}

var _ HasScope = (*Function)(nil)
var _ Symbol = (*Function)(nil)

func newFunction(document *Document, node *sitter.Node) Symbol {
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
		switch child.Type() {
		case "name":
			function.Name = NewTypeString(document.GetNodeText(child))
		case "formal_parameters":
			function.analyseParameterDeclarationList(document, child)
			if phpDoc != nil {
				function.applyPhpDoc(document, *phpDoc)
			}
			for _, param := range function.Params {
				variableTable.add(param.ToVariable())
			}
		case "compound_statement":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	function.Name.SetNamespace(document.importTable.namespace)
	document.popVariableTable()
	return function
}

func (s *Function) analyseParameterDeclarationList(document *Document, node *sitter.Node) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "simple_parameter" {
			param := newParameter(document, child)
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

func (s *Function) IsScopeSymbol() bool {
	return false
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
