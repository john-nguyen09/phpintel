package analysis

import (
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
)

type UseType int

const (
	UseClass    UseType = iota
	UseFunction         = iota
	UseConst            = iota
)

func processNamespaceUseDeclaration(document *Document, node *sitter.Node) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	useType := UseClass
	prefix := ""
	for child != nil {
		switch child.Type() {
		case "namespace_function_or_const":
			switch document.GetNodeText(child) {
			case "function":
				useType = UseFunction
			case "const":
				useType = UseConst
			}
		case "namespace_use_clause":
			processNamespaceUseClause(document, useType, child)
		case "namespace_use_group":
			processNamespaceUseGroupClauseList(document, prefix, useType, child)
		case "namespace_name":
			prefix = document.GetNodeText(child)
		}
		child = traverser.Advance()
	}
	return nil
}

func processNamespaceUseClause(document *Document, useType UseType, node *sitter.Node) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	name := ""
	alias := ""
	for child != nil {
		switch child.Type() {
		case "qualified_name":
			name = document.GetNodeText(child)
		case "namespace_aliasing_clause":
			alias = getAliasFromNode(document, child)
		}
		child = traverser.Advance()
	}
	addUseToImportTable(document, useType, alias, name)
}

func processNamespaceUseGroupClauseList(document *Document, prefix string, useType UseType, node *sitter.Node) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		switch child.Type() {
		case "namespace_use_group_clause":
			var err error = nil
			traverser, err = traverser.Descend()
			if err != nil {
				panic(err) // Should never happen
			}

			name := ""
			alias := ""
			child = traverser.Advance()
			for child != nil {
				switch child.Type() {
				case "namespace_name":
					name = document.GetNodeText(child)
				case "namespace_aliasing_clause":
					alias = getAliasFromNode(document, child)
				}
				child = traverser.Advance()
			}
			name = prefix + "\\" + name
			addUseToImportTable(document, useType, alias, name)

			traverser, err = traverser.Ascend()
			if err != nil {
				panic(err) // Should never happen
			}
		}
		traverser.Advance()
		child = traverser.Peek()
	}
}

func getAliasFromNode(document *Document, node *sitter.Node) string {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if child.Type() == "name" {
			return document.GetNodeText(child)
		}
		child = traverser.Advance()
	}
	return ""
}

func addUseToImportTable(document *Document, useType UseType, alias string, name string) {
	switch useType {
	case UseClass:
		document.importTable.addClassName(alias, name)
	case UseFunction:
		document.importTable.addFunctionName(alias, name)
	case UseConst:
		document.importTable.addConstName(alias, name)
	}
}
