package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type ImportTable struct {
	start     protocol.Position
	namespace *Namespace
	classes   map[string]string
	functions map[string]string
	constants map[string]string
}

func newImportTable(document *Document, node *sitter.Node) *ImportTable {
	return &ImportTable{
		start:     util.PointToPosition(node.StartPoint()),
		namespace: nil,
		classes:   map[string]string{},
		functions: map[string]string{},
		constants: map[string]string{},
	}
}

func makeSureAliasIsNotEmpty(alias string, name string) string {
	if alias == "" {
		parts := strings.Split(name, "\\")
		alias = parts[len(parts)-1]
	}
	return alias
}

func (i *ImportTable) addClassName(alias string, name string) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.classes[alias] = name
}

func (i *ImportTable) addFunctionName(alias string, name string) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.functions[alias] = name
}

func (i *ImportTable) addConstName(alias string, name string) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.constants[alias] = name
}

func (i *ImportTable) setNamespace(namespace *Namespace) {
	i.namespace = namespace
}

func (i ImportTable) GetClassReferenceFQN(name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	if fqn, ok := i.classes[firstPart]; ok {
		fqn = "\\" + fqn
		if len(parts) > 0 {
			fqn += "\\" + strings.Join(parts, "\\")
		}
		name.SetFQN(fqn)
	} else {
		name.SetNamespace(i.GetNamespace())
	}
	return name.GetFQN()
}

func (i ImportTable) GetFunctionReferenceFQN(store *Store, name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	aliasTable := i.functions
	if len(parts) > 0 {
		aliasTable = i.classes
	}

	if fqn, ok := aliasTable[firstPart]; ok {
		return "\\" + fqn
	}
	fqn := name.GetFQN()
	if !IsFQN(fqn) {
		fqn = "\\" + fqn
	}
	functions := store.GetFunctions(fqn)
	if len(functions) > 0 {
		return fqn
	}
	name.SetNamespace(i.GetNamespace())
	return name.GetFQN()
}

func (i ImportTable) GetConstReferenceFQN(store *Store, name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	aliasTable := i.constants
	if len(parts) > 0 {
		aliasTable = i.classes
	}

	if fqn, ok := aliasTable[firstPart]; ok {
		return "\\" + fqn
	}
	fqn := name.GetFQN()
	if !IsFQN(fqn) {
		fqn = "\\" + fqn
	}
	constants := store.GetConsts(fqn)
	if len(constants) > 0 {
		return fqn
	}
	// TODO: Defines do not have implicit namespace except
	// explicitly stated, e.g. define(__NAMESPACE__ . '\const1', true)
	defines := store.GetDefines(fqn)
	if len(defines) > 0 {
		return fqn
	}
	name.SetNamespace(i.GetNamespace())
	return name.GetFQN()
}

func (i ImportTable) ResolveToQualified(document *Document, symbol Symbol, name TypeString, word string) (string, *protocol.TextEdit) {
	if IsFQN(word) {
		return name.GetOriginal(), nil
	}
	wordNamespace := i.ResolveScopeNamespace(word)
	nameNamespace := name.GetNamespace()
	if wordNamespace == nameNamespace {
		return name.GetOriginal(), nil
	}
	nameParts := name.GetParts()
	firstPart, nameParts := nameParts[0], nameParts[1:]
	if fqn, ok := i.classes[firstPart]; ok && strings.Index(word, fqn) == 0 {
		if len(nameParts) > 0 {
			return strings.Join(nameParts, "\\"), nil
		}
		return firstPart, nil
	}
	// TODO: Defines do not have implicit namespace except
	// explicitly stated, e.g. define(__NAMESPACE__ . '\const1', true)
	switch symbol.(type) {
	case *Function, *Const, *Define:
		// Functions, constants on \ can be used within any namespaces without the need of
		// use statement
		if nameNamespace == "\\" {
			return name.GetOriginal(), nil
		}
	}
	// Aliases
	for alias, fqn := range i.classes {
		if "\\"+fqn == name.GetFQN() {
			return alias, nil
		}
	}
	// Anything below will require insert use
	insertUse := GetInsertUseContext(document)
	// Calculate what to inlude in insert use, because it is possible
	// that the current word has a scope, e.g. TestNamespace1\Class.
	// Therefore, it only needs `use TestNamespace1;`
	currentWordScope, _ := GetScopeAndNameFromString(word)
	if currentWordScope != "" && currentWordScope == nameNamespace {
		// If this goes here we don't need to include `function` or `const` in use
		// because we only insert the namespace
		return name.GetOriginal(), insertUse.GetUseEdit(NewTypeString(nameNamespace), nil, "")
	}
	return name.GetOriginal(), insertUse.GetUseEdit(name, symbol, "")
}

func (i ImportTable) GetNamespace() string {
	if i.namespace == nil {
		return ""
	}
	return i.namespace.Name
}

func (i ImportTable) ResolveScopeNamespace(word string) string {
	name := NewTypeString(word)
	name.SetNamespace(i.GetNamespace())
	return name.GetNamespace()
}
