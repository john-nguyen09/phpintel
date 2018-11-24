package entity

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

type Param struct {
	Location lsp.Range
	Name     string
	Type     string
}

func NewParam(phrase *phrase.Phrase, doc *PhpDoc) *Param {
	param := &Param{
		Location: util.NodeRange(phrase, doc.Text),
		Name:     "",
		Type:     ""}

	param.Consume(phrase, doc)

	return param
}

func (param *Param) Consume(p *phrase.Phrase, doc *PhpDoc) {
	switch p.Type {
	case phrase.ParameterDeclaration:

	}
}
