package analysis

import (
	"encoding/json"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/storage"
)

// Aliases is a constant to look up aliases (e.g. boolean is bool)
var /* const */ Aliases = map[string]string{
	"boolean": "bool",
	"integer": "int",
}

// Natives is a constant to look up native types
var /* const */ Natives = map[string]bool{
	"mixed":    true,
	"null":     true,
	"bool":     true,
	"true":     true,
	"false":    true,
	"int":      true,
	"float":    true,
	"real":     true,
	"double":   true,
	"string":   true,
	"binary":   true,
	"array":    true,
	"object":   true,
	"callable": true,
	"void":     true,

	"__DIR__":  true,
	"__FILE__": true,
}

// TypeString contains fqn and original name of type
type TypeString struct {
	fqn        string
	original   string
	arrayLevel int
}

func NewTypeString(typeString string) TypeString {
	symbolTypeString := TypeString{
		original: typeString,
	}
	symbolTypeString = symbolTypeString.resolveRawWithArrayLevel()

	if alias, ok := Aliases[symbolTypeString.original]; ok {
		symbolTypeString.original = alias
	}
	symbolTypeString.fqn = symbolTypeString.original

	return symbolTypeString
}

// MarshalJSON is used for json.Marshal
func (t *TypeString) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.fqn)
}

func (t TypeString) resolveRawWithArrayLevel() TypeString {
	o := t.original
	for i := len(o) - 1; i >= 0; i -= 2 {
		if o[i] != ']' || o[i-1] != '[' {
			t.original = o[:i+1]
			break
		}
		t.arrayLevel++
	}
	return t
}

func (t TypeString) Dearray() (TypeString, bool) {
	if t.arrayLevel == 0 {
		return t, false
	}
	t.arrayLevel--
	return t, true
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

// ToString returns the string representation of the type
func (t TypeString) ToString() string {
	arraySuffices := []string{}
	for i := 0; i < t.arrayLevel; i++ {
		arraySuffices = append(arraySuffices, "[]")
	}
	return t.GetFQN() + strings.Join(arraySuffices, "")
}

func (t TypeString) GetNamespace() string {
	lastBackslashIndex := strings.LastIndex(t.GetFQN(), "\\")
	return t.GetFQN()[:lastBackslashIndex+1]
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

func (t TypeString) GetFirstAndRestParts() (string, []string) {
	parts := t.GetParts()
	return parts[0], parts[1:]
}

func (t TypeString) GetParts() []string {
	parts := strings.Split(t.GetFQN(), "\\")
	if len(parts) > 0 && parts[0] == "" {
		return parts[1:]
	}
	return parts
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

func (t TypeString) Write(e *storage.Encoder) {
	e.WriteString(t.original)
	e.WriteString(t.fqn)
	e.WriteInt(t.arrayLevel)
}

func ReadTypeString(d *storage.Decoder) TypeString {
	return TypeString{
		original:   d.ReadString(),
		fqn:        d.ReadString(),
		arrayLevel: d.ReadInt(),
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

func typesFromPhpDoc(document *Document, text string) TypeComposite {
	parts := strings.Split(text, "|")
	types := newTypeComposite()
	for _, part := range parts {
		if IsNameRelative(part) {
			currentClass := document.getLastClass()
			switch v := currentClass.(type) {
			case *Class:
				types.add(v.Name)
			case *Interface:
				types.add(v.Name)
			case *Trait:
				types.add(v.Name)
			}
			continue
		}
		typeString := NewTypeString(strings.TrimSpace(part))
		typeString.SetFQN(document.GetImportTable().GetClassReferenceFQN(typeString))
		types.add(typeString)
	}
	return types
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

func (t *TypeComposite) Write(e *storage.Encoder) {
	e.WriteInt(len(t.typeStrings))
	for _, typeString := range t.typeStrings {
		typeString.Write(e)
	}
}

func ReadTypeComposite(d *storage.Decoder) TypeComposite {
	count := d.ReadInt()
	types := TypeComposite{}
	for i := 0; i < count; i++ {
		types.typeStrings = append(types.typeStrings, ReadTypeString(d))
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
			contents = append(contents, typeString.ToString())
		}
	}
	return strings.Join(contents, "|")
}

func GetNameParts(name string) []string {
	parts := strings.Split(name, "\\")
	if len(parts) == 1 {
		return parts
	}
	if parts[0] == "" {
		return parts[1:]
	}
	return parts
}

func GetScopeAndNameFromString(name string) (string, string) {
	parts := GetNameParts(name)
	if len(parts) == 1 {
		return "", parts[0]
	}
	return "\\" + strings.Join(parts[:len(parts)-1], "\\"), parts[len(parts)-1]
}
