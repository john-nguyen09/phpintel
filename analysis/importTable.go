package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type ImportTable struct {
	namespace string
	classes   map[string]string
	functions map[string]string
	constants map[string]string
}

func newImportTable() ImportTable {
	return ImportTable{
		namespace: "\\",
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
	if namespace.Name == "" {
		return
	}
	i.namespace = namespace.Name
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
		name.SetNamespace(i.namespace)
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
	name.SetNamespace(i.namespace)
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
	name.SetNamespace(i.namespace)
	return name.GetFQN()
}

func (i ImportTable) ResolveToQualified(document *Document, symbol Symbol,
	name TypeString, word string) (string, *protocol.TextEdit) {
	insertUse := GetInsertUseContext(document)
	parts := name.GetParts()
	firstPart, parts := parts[0], parts[1:]
	if fqn, ok := i.classes[firstPart]; ok && "\\"+fqn == name.GetFQN() {
		if len(parts) > 0 {
			return firstPart + "\\" + strings.Join(parts, "\\"), nil
		}
		return firstPart, nil
	}
	if strings.Index(name.GetFQN(), i.namespace) == 0 {
		return name.GetFQN()[len(i.namespace):], nil
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
		scope, _ := GetScopeAndNameFromString(name.GetFQN())
		if scope == "\\" {
			return name.GetOriginal(), nil
		}
	}
	return name.GetOriginal(), insertUse.GetUseEdit(name, symbol, "")
}

func (i ImportTable) GetNamespace() string {
	return i.namespace
}
