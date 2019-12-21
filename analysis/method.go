package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
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

func newMethod(document *Document, node *phrase.Phrase) Symbol {
	symbol := newFunction(document, node)
	method := &Method{
		IsStatic: false,
	}

	if function, ok := symbol.(*Function); ok {
		method.Name = function.Name.GetOriginal()
		method.Params = function.Params
		method.returnTypes = function.returnTypes
		method.description = function.description
	}
	method.location = document.GetNodeLocation(node)
	lastClass := document.getLastClass()
	if theClass, ok := lastClass.(*Class); ok {
		method.Scope = theClass.Name
		method.Scope.SetNamespace(document.GetImportTable().GetNamespace())
	} else if theInterface, ok := lastClass.(*Interface); ok {
		method.Scope = theInterface.Name
		method.Scope.SetNamespace(document.GetImportTable().GetNamespace())
	} else if trait, ok := lastClass.(*Trait); ok {
		method.Scope = trait.Name
		method.Scope.SetNamespace(document.GetImportTable().GetNamespace())
	}

	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := util.IsOfPhraseType(child, phrase.MethodDeclarationHeader); ok {
			method.analyseHeader(p)
		}
		child = traverser.Advance()
	}

	return method
}

func (s Method) GetLocation() protocol.Location {
	return s.location
}

func (s *Method) analyseHeader(methodHeader *phrase.Phrase) {
	traverser := util.NewTraverser(methodHeader)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.MemberModifierList:
				s.VisibilityModifier, s.IsStatic, s.ClassModifier = getMemberModifier(p)
			}
		}
		child = traverser.Advance()
	}
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

func (s *Method) GetPrefixes() []string {
	return []string{s.GetScope().GetFQN()}
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

func (s *Method) GetScope() TypeString {
	return s.Scope
}

func (s *Method) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	serialiser.WriteString(s.Name)
	serialiser.WriteInt(len(s.Params))
	for _, param := range s.Params {
		param.Write(serialiser)
	}
	s.returnTypes.Write(serialiser)
	serialiser.WriteString(s.description)

	s.Scope.Write(serialiser)
	serialiser.WriteInt(int(s.VisibilityModifier))
	serialiser.WriteBool(s.IsStatic)
	serialiser.WriteInt(int(s.ClassModifier))
}

func ReadMethod(serialiser *Serialiser) *Method {
	method := Method{
		location: serialiser.ReadLocation(),
		Name:     serialiser.ReadString(),
		Params:   make([]*Parameter, 0),
	}
	countParams := serialiser.ReadInt()
	for i := 0; i < countParams; i++ {
		method.Params = append(method.Params, ReadParameter(serialiser))
	}
	method.returnTypes = ReadTypeComposite(serialiser)
	method.description = serialiser.ReadString()

	method.Scope = ReadTypeString(serialiser)
	method.VisibilityModifier = VisibilityModifierValue(serialiser.ReadInt())
	method.IsStatic = serialiser.ReadBool()
	method.ClassModifier = ClassModifierValue(serialiser.ReadInt())

	return &method
}
