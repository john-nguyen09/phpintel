package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

// Method contains information of methods
type Method struct {
	location protocol.Location

	Name               string
	Params             []*Parameter
	returnTypes        TypeComposite
	description        string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
	ClassModifier      ClassModifierValue
}

var _ HasScope = (*Method)(nil)
var _ Symbol = (*Method)(nil)

func newMethodFromPhpDocTag(document *Document, class *Class, methodTag tag, location protocol.Location) *Method {
	method := &Method{
		IsStatic:    methodTag.IsStatic,
		Name:        methodTag.Name,
		location:    location,
		returnTypes: typesFromPhpDoc(document, methodTag.TypeString),
		Params:      []*Parameter{},
		description: methodTag.Description,
		Scope:       class.Name,
	}
	for _, paramTag := range methodTag.Parameters {
		param := &Parameter{
			location: location,
			Name:     paramTag.Name,
			Value:    paramTag.Value,
			Type:     typesFromPhpDoc(document, paramTag.TypeString),
		}
		method.Params = append(method.Params, param)
	}
	return method
}

func (s *Method) analyseMethodNode(document *Document, node *sitter.Node) {
	s.Params = []*Parameter{}
	s.returnTypes = newTypeComposite()
	phpDoc := document.getValidPhpDoc(s.location)
	document.pushVariableTable(node)

	variableTable := document.getCurrentVariableTable()
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "visibility_modifier":
			s.VisibilityModifier = getMemberModifier(child)
		case "static_modifier":
			s.IsStatic = true
		case "class_modifier":
			s.ClassModifier = getClassModifier(child)
		case "name":
			s.Name = document.GetNodeText(child)
		case "formal_parameters":
			s.analyseParameterDeclarationList(document, child)
			if phpDoc != nil {
				s.applyPhpDoc(document, *phpDoc)
			}
			document.addSymbol(s)
			for _, param := range s.Params {
				variableTable.add(param.ToVariable())
			}
		case "compound_statement":
			scanForChildren(document, child)
		}
		child = traverser.Advance()
	}
	document.popVariableTable()
}

func newMethod(document *Document, node *sitter.Node) Symbol {
	method := &Method{
		IsStatic:    false,
		location:    document.GetNodeLocation(node),
		Params:      []*Parameter{},
		returnTypes: newTypeComposite(),
	}
	method.analyseMethodNode(document, node)

	lastClass := document.getLastClass()
	if theClass, ok := lastClass.(*Class); ok {
		method.Scope = theClass.Name
		method.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if theInterface, ok := lastClass.(*Interface); ok {
		method.Scope = theInterface.Name
		method.Scope.SetNamespace(document.currImportTable().GetNamespace())
	} else if trait, ok := lastClass.(*Trait); ok {
		method.Scope = trait.Name
		method.Scope.SetNamespace(document.currImportTable().GetNamespace())
	}

	return nil
}

func (s Method) GetLocation() protocol.Location {
	return s.location
}

func (s *Method) analyseParameterDeclarationList(document *Document, node *sitter.Node) {
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

func (s *Method) applyPhpDoc(document *Document, phpDoc phpDocComment) {
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

func (s Method) GetName() string {
	return s.Name
}

func (s Method) GetDescription() string {
	return s.description
}

func (s Method) GetReturnTypes() TypeComposite {
	return s.returnTypes
}

func (s *Method) GetCollection() string {
	return methodCollection
}

func (s *Method) GetKey() string {
	return s.Scope.GetFQN() + KeySep + s.Name + KeySep + s.location.URI
}

func (s *Method) GetIndexableName() string {
	return s.Name
}

func (s *Method) GetIndexCollection() string {
	return methodCompletionIndex
}

func (s *Method) GetNameLabel() string {
	label := s.VisibilityModifier.ToString()
	if s.IsStatic {
		label += " static"
	}
	label += " " + s.Name
	return label
}

func (s *Method) GetParams() []*Parameter {
	return s.Params
}

func (s *Method) GetScope() string {
	return s.Scope.GetFQN()
}

func (s *Method) IsScopeSymbol() bool {
	return true
}

func (s *Method) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	e.WriteString(s.Name)
	e.WriteInt(len(s.Params))
	for _, param := range s.Params {
		param.Write(e)
	}
	s.returnTypes.Write(e)
	e.WriteString(s.description)

	s.Scope.Write(e)
	e.WriteInt(int(s.VisibilityModifier))
	e.WriteBool(s.IsStatic)
	e.WriteInt(int(s.ClassModifier))
}

func ReadMethod(d *storage.Decoder) *Method {
	method := Method{
		location: d.ReadLocation(),
		Name:     d.ReadString(),
		Params:   make([]*Parameter, 0),
	}
	countParams := d.ReadInt()
	for i := 0; i < countParams; i++ {
		method.Params = append(method.Params, ReadParameter(d))
	}
	method.returnTypes = ReadTypeComposite(d)
	method.description = d.ReadString()

	method.Scope = ReadTypeString(d)
	method.VisibilityModifier = VisibilityModifierValue(d.ReadInt())
	method.IsStatic = d.ReadBool()
	method.ClassModifier = ClassModifierValue(d.ReadInt())

	return &method
}
