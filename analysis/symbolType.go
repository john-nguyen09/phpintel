package analysis

import (
	"encoding/json"
	"strings"
)

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

	"__DIR__":  true,
	"__FILE__": true,
}

// TypeString contains fqn and original name of type
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

// MarshalJSON is used for json.Marshal
func (t *TypeString) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.fqn)
}

// IsEmpty checks whether TypeString is empty
func (t TypeString) IsEmpty() bool {
	return t.fqn == "" || t.fqn == "\\"
}

// GetOriginal gets original name
func (t TypeString) GetOriginal() string {
	return t.original
}

// GetFQN gets the FQN converted name
func (t TypeString) GetFQN() string {
	return t.fqn
}

func (t *TypeString) SetFQN(fqn string) {
	t.fqn = fqn
}

func (t *TypeString) SetNamespace(namespace string) {
	if !isFQN(t.fqn) {
		if namespace == "" || namespace == "\\" {
			t.fqn = "\\" + t.fqn
		} else {
			t.fqn = "\\" + namespace + "\\" + t.fqn
		}
	}
}

func (t TypeString) FirstPart() string {
	if strings.Index(t.original, "\\") != -1 {
		return strings.Split(t.original, "\\")[0]
	}
	return t.original
}

func (t TypeString) GetParts() []string {
	return strings.Split(t.original, "\\")
}

func isFQN(name string) bool {
	if name == "" {
		return false
	}
	if _, ok := Natives[name]; ok {
		return true
	}
	return name[0] == '\\'
}

// GetType gets the FQN of type
func (t TypeString) GetType() string {
	return t.fqn
}

func (t *TypeString) Write(serialiser *Serialiser) {
	serialiser.WriteString(t.original)
	serialiser.WriteString(t.fqn)
}

func ReadTypeString(serialiser *Serialiser) TypeString {
	return TypeString{
		original: serialiser.ReadString(),
		fqn:      serialiser.ReadString(),
	}
}

// TypeComposite contains multiple type strings
type TypeComposite struct {
	typeStrings []TypeString
	uniqueFQNs  map[string]bool
}

func newTypeComposite() TypeComposite {
	return TypeComposite{
		typeStrings: []TypeString{},
		uniqueFQNs:  map[string]bool{},
	}
}

// MarshalJSON marshals TypeComposite to JSON
func (t TypeComposite) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.typeStrings)
}

func (t *TypeComposite) add(typeString TypeString) {
	if _, ok := t.uniqueFQNs[typeString.GetFQN()]; ok {
		return
	}
	t.typeStrings = append(t.typeStrings, typeString)
	if t.uniqueFQNs == nil {
		t.uniqueFQNs = map[string]bool{}
	}
	t.uniqueFQNs[typeString.GetFQN()] = true
}

func (t *TypeComposite) merge(types TypeComposite) {
	for _, typeString := range types.typeStrings {
		t.add(typeString)
	}
}

func (t *TypeComposite) Write(serialiser *Serialiser) {
	serialiser.WriteInt(len(t.typeStrings))
	for _, typeString := range t.typeStrings {
		typeString.Write(serialiser)
	}
}

func ReadTypeComposite(serialiser *Serialiser) TypeComposite {
	count := serialiser.ReadInt()
	types := TypeComposite{}
	for i := 0; i < count; i++ {
		types.typeStrings = append(types.typeStrings, ReadTypeString(serialiser))
	}
	return types
}

// Resolve resolves the type to slice of TypeString
func (t TypeComposite) Resolve() []TypeString {
	return t.typeStrings
}

func (t TypeComposite) IsEmpty() bool {
	types := t.Resolve()
	isAllTypesEmpty := true
	if len(types) > 0 {
		for _, typeString := range types {
			if !typeString.IsEmpty() {
				isAllTypesEmpty = false
				break
			}
		}
	}
	return len(types) == 0 || isAllTypesEmpty
}

func (t TypeComposite) ToString() string {
	types := t.Resolve()
	contents := []string{}
	if len(types) > 0 {
		for _, typeString := range types {
			if typeString.IsEmpty() {
				continue
			}
			contents = append(contents, typeString.GetFQN())
		}
	}
	return strings.Join(contents, "|")
}
