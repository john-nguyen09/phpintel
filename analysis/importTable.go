package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type importItem struct {
	name          string
	locationRange protocol.Range
	isUsed        bool
}

type ImportTable struct {
	start     protocol.Position
	namespace *Namespace
	classes   map[string]*importItem
	functions map[string]*importItem
	constants map[string]*importItem
}

func newImportTable(document *Document, node *phrase.Phrase) *ImportTable {
	start := 0
	firstToken := util.FirstToken(node)
	if firstToken != nil {
		start = firstToken.Offset
	}
	return &ImportTable{
		start:     document.positionAt(start),
		namespace: nil,
		classes:   map[string]*importItem{},
		functions: map[string]*importItem{},
		constants: map[string]*importItem{},
	}
}

func makeSureAliasIsNotEmpty(alias string, name string) string {
	if alias == "" {
		parts := strings.Split(name, "\\")
		alias = parts[len(parts)-1]
	}
	return alias
}

func (i *ImportTable) addClassName(alias string, name string, r protocol.Range) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.classes[alias] = &importItem{
		name:          name,
		locationRange: r,
	}
}

func (i *ImportTable) addFunctionName(alias string, name string, r protocol.Range) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.functions[alias] = &importItem{
		name:          name,
		locationRange: r,
	}
}

func (i *ImportTable) addConstName(alias string, name string, r protocol.Range) {
	alias = makeSureAliasIsNotEmpty(alias, name)
	i.constants[alias] = &importItem{
		name:          name,
		locationRange: r,
	}
}

func (i *ImportTable) setNamespace(namespace *Namespace) {
	i.namespace = namespace
}

func (i ImportTable) GetClassReferenceFQN(name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	if item, ok := i.classes[firstPart]; ok {
		fqn := "\\" + item.name
		if len(parts) > 0 {
			fqn += "\\" + strings.Join(parts, "\\")
		}
		name.SetFQN(fqn)
		i.useClass(firstPart)
	} else {
		name.SetNamespace(i.GetNamespace())
	}
	return name.GetFQN()
}

func (i ImportTable) GetFunctionReferenceFQN(q *Query, name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	aliasTable := i.functions
	if len(parts) > 0 {
		aliasTable = i.classes
	}

	if item, ok := aliasTable[firstPart]; ok {
		return "\\" + item.name
	}
	fqn := name.GetFQN()
	if !IsFQN(fqn) {
		fqn = "\\" + fqn
	}
	functions := q.GetFunctions(fqn)
	if len(functions) > 0 {
		return fqn
	}
	name.SetNamespace(i.GetNamespace())
	return name.GetFQN()
}

func (i ImportTable) functionPossibleFQNs(name TypeString) []string {
	firstPart, parts := name.GetFirstAndRestParts()
	aliasTable := i.functions
	if len(parts) > 0 {
		aliasTable = i.classes
	}
	results := []string{}
	if item, ok := aliasTable[firstPart]; ok {
		results = append(results, "\\"+item.name)
		return results
	}
	fqn := name.GetFQN()
	if !IsFQN(fqn) {
		fqn = "\\" + fqn
	}
	results = append(results, fqn)
	name.SetNamespace(i.GetNamespace())
	if name.GetFQN() != fqn {
		results = append(results, name.GetFQN())
	}
	return results
}

func (i ImportTable) GetConstReferenceFQN(q *Query, name TypeString) string {
	firstPart, parts := name.GetFirstAndRestParts()
	aliasTable := i.constants
	if len(parts) > 0 {
		aliasTable = i.classes
	}

	if item, ok := aliasTable[firstPart]; ok {
		return "\\" + item.name
	}
	fqn := name.GetFQN()
	if !IsFQN(fqn) {
		fqn = "\\" + fqn
	}
	constants := q.GetConsts(fqn)
	if len(constants) > 0 {
		return fqn
	}
	// TODO: Defines do not have implicit namespace except
	// explicitly stated, e.g. define(__NAMESPACE__ . '\const1', true)
	defines := q.GetDefines(fqn)
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
	if item, ok := i.classes[firstPart]; ok && strings.Index(word, item.name) == 0 {
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
	for alias, item := range i.classes {
		if "\\"+item.name == name.GetFQN() {
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

func (i *ImportTable) useClass(name string) {
	if item, ok := i.classes[name]; ok {
		item.isUsed = true
	}
}

func (i *ImportTable) useFunction(name string) {
	if item, ok := i.functions[name]; ok {
		item.isUsed = true
	}
}

func (i *ImportTable) useConstant(name string) {
	if item, ok := i.constants[name]; ok {
		item.isUsed = true
	}
}

func (i *ImportTable) useFunctionOrClass(name TypeString) {
	firstPart, parts := name.GetFirstAndRestParts()
	if len(parts) > 0 {
		i.useClass(firstPart)
	} else {
		i.useFunction(firstPart)
	}
}

func (i *ImportTable) useConstOrClass(name TypeString) {
	firstPart, parts := name.GetFirstAndRestParts()
	if len(parts) > 0 {
		i.useClass(firstPart)
	} else {
		i.useConstant(firstPart)
	}
}

func (i ImportTable) unusedImportItems() []*importItem {
	var results []*importItem
	for _, item := range i.classes {
		if !item.isUsed {
			results = append(results, item)
		}
	}
	for _, item := range i.functions {
		if !item.isUsed {
			results = append(results, item)
		}
	}
	for _, item := range i.constants {
		if !item.isUsed {
			results = append(results, item)
		}
	}
	return results
}
