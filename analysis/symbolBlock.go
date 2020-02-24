package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type symbolConstructorForPhrase func(*Document, *sitter.Node) Symbol

var /* const */ scanPhraseTypes = map[string]bool{
	"ERROR":                     true,
	"expression_statement":      true,
	"compound_statement":        true,
	"while_statement":           true,
	"case_statement":            true,
	"default_statement":         true,
	"array_creation_expression": true,
	"array_element_initializer": true,
	"if_statement":              true,
	"else_if_clause":            true,
	"else_if_clause_2":          true,
	"else_clause":               true,
	"else_clause_2":             true,
	"include_expression":        true,
	"include_once_expression":   true,
	"require_expression":        true,
	"require_once_expression":   true,
	"conditional_expression":    true,
	"subscript_expression":      true,
	"cast_expression":           true,
	"unary_op_expression":       true,
	"binary_expression":         true,
	"parenthesized_expression":  true,
	"echo_statement":            true,
	"unset_statement":           true,
	"print_intrinsic":           true,
	"try_statement":             true,
	"catch_clause":              true,
	"finally_clause":            true,
	"return_statement":          true,
	"throw_statement":           true,
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
		"property_declaration":                   newPropertyDeclaration,
		"method_declaration":                     newMethod,
		"constructor_declaration":                newMethod,
		"destructor_declaration":                 newMethod,
		"trait_use_clause":                       processTraitUseClause,
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
		"comment":                                newPhpDocFromNode,
	}
}

func scanNode(document *Document, node *sitter.Node) {
	var symbol Symbol = nil
	shouldSkipAdding := false
	if node.Type() == "namespace_definition" {
		namespace := newNamespace(document, node)
		document.setNamespace(namespace)
		return
	}

	scanForExpression(document, node)
	if _, ok := scanPhraseTypes[node.Type()]; ok {
		scanForChildren(document, node)
		return
	}
	if constructor, ok := phraseToSymbolConstructor[node.Type()]; ok {
		symbol = constructor(document, node)
	}
	if _, ok := skipAddingSymbol[node.Type()]; ok {
		shouldSkipAdding = true
	}

	if !shouldSkipAdding && symbol != nil {
		document.addSymbol(symbol)
	}
}

func scanForChildren(document *Document, node *sitter.Node) {
	childCount := int(node.ChildCount())
	for i := 0; i < childCount; i++ {
		child := node.Child(i)
		scanNode(document, child)
	}
}
