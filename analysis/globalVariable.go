package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type GlobalVariable struct {
	location    protocol.Location
	types       TypeComposite
	description string

	Name string
}

func newGlobalDeclaration(document *Document, node *ast.Node) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "variable_name":
			globalVariable := newGlobalVariable(document, child)
			if globalVariable != nil {
				document.addSymbol(globalVariable)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func newGlobalVariable(document *Document, node *ast.Node) Symbol {
	globalVariable := &GlobalVariable{
		location: document.GetNodeLocation(node),
		Name:     document.GetNodeText(node),
		types:    newTypeComposite(),
	}
	phpDoc := document.getValidPhpDoc(globalVariable.location)
	if phpDoc != nil {
		globalVariable.applyPhpDoc(document, phpDoc)
	}
	variableTable := document.getCurrentVariableTable()
	variableTable.setReferenceGlobal(globalVariable.GetName())
	return globalVariable
}

func (s *GlobalVariable) applyPhpDoc(document *Document, phpDoc *phpDocComment) {
	tags := phpDoc.Globals
	for _, tag := range tags {
		if tag.Name == s.Name {
			s.types.merge(typesFromPhpDoc(document, tag.TypeString))
			s.description = tag.Description
			break
		}
	}
}

func (s *GlobalVariable) GetLocation() protocol.Location {
	return s.location
}

func (s *GlobalVariable) GetName() string {
	return s.Name
}

func (s *GlobalVariable) GetDescription() string {
	return s.description
}

func (s *GlobalVariable) GetDetail() string {
	return s.types.ToString()
}

func (s *GlobalVariable) GetCollection() string {
	return globalVariableCollection
}

func (s *GlobalVariable) GetKey() string {
	if s.types.IsEmpty() {
		return ""
	}
	return s.Name + KeySep + s.location.URI
}

func (s *GlobalVariable) Serialise(e *storage.Encoder) {
	e.WriteLocation(s.location)
	s.types.Write(e)
	e.WriteString(s.description)
	e.WriteString(s.Name)
}

func ReadGlobalVariable(d *storage.Decoder) *GlobalVariable {
	return &GlobalVariable{
		location:    d.ReadLocation(),
		types:       ReadTypeComposite(d),
		description: d.ReadString(),
		Name:        d.ReadString(),
	}
}
