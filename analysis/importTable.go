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
	if !isFQN(fqn) {
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
	if !isFQN(fqn) {
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
	insertUse := GetInsertUseContext(document)
	parts := name.GetParts()
	currentScope := i.ResolveScopeNamespace(word)
	firstPart, parts := parts[0], parts[1:]
	if fqn, ok := i.classes[firstPart]; ok && strings.Index(word, fqn) == 0 {
		if len(parts) > 0 {
			return strings.Join(parts, "\\"), nil
		}
		return firstPart, nil
	}
	if currentScope != "" && strings.Index(name.GetFQN(), currentScope) == 0 {
		return name.GetFQN()[len(currentScope)+1:], nil
	}
	if currentScope != "" && currentScope == name.GetNamespace() {
		return name.GetOriginal(), nil
	}
	wordScope, _ := GetScopeAndNameFromString(word)
	scope := name.GetNamespace()
	if wordScope != "" && wordScope == scope {
		return name.GetOriginal(), insertUse.GetUseEdit(NewTypeString(scope), nil, "")
	}
	for alias, fqn := range i.classes {
		if "\\"+fqn == name.GetFQN() {
			return alias, nil
		}
	}
	if isFQN(word) {
		return name.GetOriginal(), nil
	}
	// TODO: Defines do not have implicit namespace except
	// explicitly stated, e.g. define(__NAMESPACE__ . '\const1', true)
	switch symbol.(type) {
	case *Function, *Const, *Define:
		if scope == "" {
			return name.GetOriginal(), nil
		}
	}
	return name.GetOriginal(), insertUse.GetUseEdit(name, symbol, "")
}

func (i ImportTable) GetNamespace() string {
	if i.namespace == nil {
		return "\\"
	}
	return i.namespace.Name
}

func (i ImportTable) ResolveScopeNamespace(word string) string {
	name := NewTypeString(word)
	name.SetNamespace(i.GetNamespace())
	return name.GetNamespace()
}
