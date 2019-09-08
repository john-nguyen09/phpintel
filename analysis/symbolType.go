package analysis

import (
	"encoding/json"
)

type SymbolType interface {
	Resolve(document *Document) []TypeString
}

var /* const */ Aliases = map[string]string{
	"boolean": "bool",
	"integer": "int",
}
var /* const */ Natives = map[string]bool{
	"mixed":  true,
	"null":   true,
	"bool":   true,
	"int":    true,
	"float":  true,
	"real":   true,
	"double": true,
	"string": true,
	"binary": true,
	"array":  true,
	"object": true,
}

type TypeString struct {
	fqn      string
	original string
}

func NewTypeString(typeString string) TypeString {
	symbolTypeString := TypeString{
		original: typeString,
	}

	if alias, ok := Aliases[typeString]; ok {
		typeString = alias
	}
	symbolTypeString.fqn = typeString

	return symbolTypeString
}

func (t *TypeString) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.fqn)
}

func (t TypeString) IsFqn() bool {
	if _, ok := Natives[t.fqn]; ok {
		return true
	}

	return []rune(t.fqn)[0] == '\\'
}

func (t *TypeString) SetFqn(fqn string) {
	t.fqn = fqn
}

func (t TypeString) GetType() string {
	return t.fqn
}

type TypeComposite struct {
	typeStrings []TypeString
}

func NewTypeComposite() TypeComposite {
	return TypeComposite{
		typeStrings: []TypeString{},
	}
}

func (t *TypeComposite) MarshalJSON() ([]byte, error) {
	typeStrings := []TypeString{}

	for _, typeString := range t.typeStrings {
		typeStrings = append(typeStrings, typeString)
	}

	return json.Marshal(&typeStrings)
}

func (t *TypeComposite) Add(typeString TypeString) {
	t.typeStrings = append(t.typeStrings, typeString)
}

func (t TypeComposite) Resolve(document *Document) []TypeString {
	return t.typeStrings
}
