package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
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

func newGlobalDeclaration(a analyser, document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.VariableNameList:
				analyseVariableNameList(a, document, p)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func analyseVariableNameList(a analyser, document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.SimpleVariable {
			globalVariable := newGlobalVariable(a, document, p)
			document.addSymbol(globalVariable)
		}
		child = traverser.Advance()
	}
}

func newGlobalVariable(a analyser, document *Document, node *phrase.Phrase) Symbol {
	globalVariable := &GlobalVariable{
		location: document.GetNodeLocation(node),
		Name:     document.getPhraseText(node),
		types:    newTypeComposite(),
	}
	phpDoc := document.getValidPhpDoc(globalVariable.location)
	if phpDoc != nil {
		globalVariable.applyPhpDoc(document, phpDoc)
	}
	variableTable := document.getCurrentVariableTable()
	variableTable.setReferenceGlobal(globalVariable.GetName())
	document.pushVariable(a, globalVariable.toVariable(), globalVariable.location.Range.End, true)
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

func (s *GlobalVariable) Serialise(e storage.Encoder) {
	e.WriteLocation(s.location)
	s.types.Write(e)
	e.WriteString(s.description)
	e.WriteString(s.Name)
}

func (s GlobalVariable) toVariable() *Variable {
	return &Variable{
		Expression: Expression{
			Location: s.location,
			Type:     s.types,
			Name:     s.Name,
		},
		description: s.description,
	}
}

func ReadGlobalVariable(d storage.Decoder) *GlobalVariable {
	return &GlobalVariable{
		location:    d.ReadLocation(),
		types:       ReadTypeComposite(d),
		description: d.ReadString(),
		Name:        d.ReadString(),
	}
}
