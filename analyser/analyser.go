package analyser

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analyser/entity"
)

type Analyser struct {
	doc      *entity.PhpDoc
	entities []entity.EntityBase
	nodes    []*phrase.Phrase
}

func NewAnalyser(doc *entity.PhpDoc) *Analyser {
	return &Analyser{
		doc:      doc,
		entities: make([]entity.EntityBase, 0),
		nodes:    make([]*phrase.Phrase, 0)}
}

func (a *Analyser) Preorder(node *phrase.Phrase) {
	var entityBase entity.EntityBase

	switch node.Type {
	case phrase.ClassDeclaration:
		classEntity := entity.NewClass(node, a.doc)
		a.doc.AddClass(classEntity)

		entityBase = classEntity
	case phrase.ParameterDeclaration:
		paramEntity := entity.NewParam(node, a.doc)

		entityBase = paramEntity
	}

	a.pushEntity(entityBase)
	if entityBase == nil {
		a.pushNode(node)
	}
}

func (a *Analyser) Postorder(node *phrase.Phrase) {
	if len(a.entities) == 0 {
		return
	}

	var entityBase entity.EntityBase
	entityBase, a.entities = a.entities[len(a.entities)-1], a.entities[:len(a.entities)-1]

	if entityBase == nil {
		return
	}

	for _, node := range a.nodes {
		entityBase.Consume(node, a.doc)
	}

	a.nodes = nil
}

func (a *Analyser) pushNode(node *phrase.Phrase) {
	a.nodes = append(a.nodes, node)
}

func (a *Analyser) pushEntity(entity entity.EntityBase) {
	a.entities = append(a.entities, entity)
}
