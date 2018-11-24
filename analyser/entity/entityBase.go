package entity

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type EntityBase interface {
	Consume(node *phrase.Phrase, doc *PhpDoc)
}
