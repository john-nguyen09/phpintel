package analysis

import (
	"encoding/json"

	"github.com/john-nguyen09/phpintel/indexer"
)

// SymbolType is an interface to symbol types
type SymbolType interface {
	Resolve() []TypeString
}

// Aliases is a constant to look up aliases (e.g. boolean is bool)
var /* const */ Aliases = map[string]string{
	"boolean": "bool",
	"integer": "int",
}

// Natives is a constant to look up native types
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

// TypeString contains fqn and original name of type
type TypeString struct {
	fqn      string
	original string
}

func newTypeString(typeString string) TypeString {
	symbolTypeString := TypeString{
		original: typeString,
	}

	if alias, ok := Aliases[typeString]; ok {
		typeString = alias
	}
	symbolTypeString.fqn = typeString

	return symbolTypeString
}

// MarshalJSON is used for json.Marshal
func (t *TypeString) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.fqn)
}

// IsFqn checks whether the type is fqn
func (t TypeString) IsFqn() bool {
	if _, ok := Natives[t.fqn]; ok {
		return true
	}

	return []rune(t.fqn)[0] == '\\'
}

// SetFqn is a setter to FQN
func (t *TypeString) SetFqn(fqn string) {
	t.fqn = fqn
}

// GetType gets the FQN of type
func (t TypeString) GetType() string {
	return t.fqn
}

func (t *TypeString) Write(serialiser *indexer.Serialiser) {
	serialiser.WriteString(t.original)
	serialiser.WriteString(t.fqn)
}

func ReadTypeString(serialiser *indexer.Serialiser) TypeString {
	return TypeString{
		original: serialiser.ReadString(),
		fqn:      serialiser.ReadString(),
	}
}

// TypeComposite contains multiple type strings
type TypeComposite struct {
	typeStrings []TypeString
}

func newTypeComposite() TypeComposite {
	return TypeComposite{
		typeStrings: []TypeString{},
	}
}

// MarshalJSON marshals TypeComposite to JSON
func (t TypeComposite) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.typeStrings)
}

func (t *TypeComposite) add(typeString TypeString) {
	t.typeStrings = append(t.typeStrings, typeString)
}

func (t *TypeComposite) merge(types TypeComposite) {
	for _, typeString := range types.typeStrings {
		t.add(typeString)
	}
}

func (t *TypeComposite) Write(serialiser *indexer.Serialiser) {
	for _, typeString := range t.typeStrings {
		typeString.Write(serialiser)
	}
}

// Resolve resolves the type to slice of TypeString
func (t TypeComposite) Resolve() []TypeString {
	return t.typeStrings
}
