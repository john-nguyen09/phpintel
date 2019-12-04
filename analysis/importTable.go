package analysis

import (
	"strings"
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
	if fqn, ok := i.classes[name.FirstPart()]; ok {
		fqn = "\\" + fqn
		name.SetFQN(fqn)
	} else {
		name.SetNamespace(i.namespace)
	}
	return name.GetFQN()
}

func (i ImportTable) GetFunctionReferenceFQN(name TypeString) string {
	if fqn, ok := i.functions[name.FirstPart()]; ok {
		name.SetNamespace(fqn)
	} else {
		name.SetNamespace(i.namespace)
	}
	return name.GetFQN()
}

func (i ImportTable) GetConstReferenceFQN(name TypeString) string {
	if fqn, ok := i.constants[name.FirstPart()]; ok {
		name.SetNamespace(fqn)
	} else {
		name.SetNamespace(i.namespace)
	}
	return name.GetFQN()
}
