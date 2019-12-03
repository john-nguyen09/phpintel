package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type GlobalVariable struct {
	location    protocol.Location
	types       TypeComposite
	description string

	Name string
}

func newGlobalDeclaration(document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.VariableNameList:
				analyseVariableNameList(document, p)
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func analyseVariableNameList(document *Document, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.SimpleVariable {
			globalVariable := newGlobalVariable(document, p)
			if globalVariable != nil {
				document.addSymbol(globalVariable)
			}
		}
		child = traverser.Advance()
	}
}

func newGlobalVariable(document *Document, node *phrase.Phrase) Symbol {
	globalVariable := &GlobalVariable{
		location: document.GetNodeLocation(node),
		Name:     document.GetPhraseText(node),
		types:    newTypeComposite(),
	}
	phpDoc := document.getValidPhpDoc(globalVariable.location)
	if phpDoc != nil {
		globalVariable.applyPhpDoc(phpDoc)
	}
	variableTable := document.getCurrentVariableTable()
	variableTable.setReferenceGlobal(globalVariable.GetName())
	return globalVariable
}

func (s *GlobalVariable) applyPhpDoc(phpDoc *phpDocComment) {
	tags := phpDoc.Globals
	for _, tag := range tags {
		if tag.Name == s.Name {
			s.types.add(NewTypeString(tag.TypeString))
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

func (s *GlobalVariable) Serialise(serialiser *Serialiser) {
	serialiser.WriteLocation(s.location)
	s.types.Write(serialiser)
	serialiser.WriteString(s.description)
	serialiser.WriteString(s.Name)
}

func ReadGlobalVariable(serialiser *Serialiser) *GlobalVariable {
	return &GlobalVariable{
		location:    serialiser.ReadLocation(),
		types:       ReadTypeComposite(serialiser),
		description: serialiser.ReadString(),
		Name:        serialiser.ReadString(),
	}
}
