package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type symbolConstructorForPhrase func(*Document, *sitter.Node) Symbol

var /* const */ scanPhraseTypes = map[string]bool{
	"expression_statement": true,
	"while_statement":      true,
}

var /* const */ skipAddingSymbol map[string]bool = map[string]bool{
	"arguments": true,
}

// var /*const */ tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
// 	// Expressions
// 	lexer.DirectoryConstant: newDirectoryConstantAccess,
// 	lexer.DocumentComment:   newPhpDocFromNode,
// }
var phraseToSymbolConstructor map[string]symbolConstructorForPhrase

func init() {
	phraseToSymbolConstructor = map[string]symbolConstructorForPhrase{
		"interface_declaration":                  newInterface,
		"class_declaration":                      newClass,
		"function_definition":                    newFunction,
		"const_declaration":                      newConstDeclaration,
		"const_element":                          newConst,
		"arguments":                              newArgumentList,
		"trait_declaration":                      newTrait,
		"function_call_expression":               tryToNewDefine,
		"assignment_expression":                  newAssignment,
		"global_declaration":                     newGlobalDeclaration,
		"namespace_use_declaration":              processNamespaceUseDeclaration,
		"anonymous_function_creation_expression": newAnonymousFunction,
	}
}

func scanForChildren(document *Document, node *sitter.Node) {
	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		var childSymbol Symbol = nil
		shouldSkipAdding := false
		if child.Type() == "namespace_definition" {
			namespace := newNamespace(document, child)
			document.setNamespace(namespace)
			continue
		}

		scanForExpression(document, child)
		if _, ok := scanPhraseTypes[child.Type()]; ok {
			scanForChildren(document, child)
			continue
		}
		if constructor, ok := phraseToSymbolConstructor[child.Type()]; ok {
			childSymbol = constructor(document, child)
		}
		if _, ok := skipAddingSymbol[child.Type()]; ok {
			shouldSkipAdding = true
		}

		if !shouldSkipAdding && childSymbol != nil {
			document.addSymbol(childSymbol)
		}
	}
}
