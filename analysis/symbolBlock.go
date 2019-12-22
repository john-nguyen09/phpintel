package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
)

type symbolConstructorForPhrase func(*Document, *phrase.Phrase) Symbol
type symbolConstructorForToken func(*Document, *lexer.Token) Symbol

var /* const */ scanPhraseTypes = map[phrase.PhraseType]bool{
	phrase.ExpressionStatement:            true,
	phrase.WhileStatement:                 true,
	phrase.ClassMemberDeclarationList:     true,
	phrase.InterfaceMemberDeclarationList: true,
	phrase.ClassConstElementList:          true,
	phrase.ClassConstDeclaration:          true,
	phrase.EncapsulatedExpression:         true,
	phrase.CompoundStatement:              true,
	phrase.StatementList:                  true,
	phrase.AdditiveExpression:             true,
	phrase.IfStatement:                    true,
	phrase.ElseClause:                     true,
	phrase.IncludeExpression:              true,
	phrase.EchoIntrinsic:                  true,
	phrase.ExpressionList:                 true,
	phrase.ClassDeclarationBody:           true,
	phrase.TryStatement:                   true,
	phrase.CatchClauseList:                true,
	phrase.CatchClause:                    true,
	phrase.ReturnStatement:                true,
	phrase.ObjectCreationExpression:       true,
	phrase.ScopedCallExpression:           true,
	phrase.ArrayCreationExpression:        true,
	phrase.ArrayInitialiserList:           true,
	phrase.ArrayElement:                   true,
	phrase.ArrayValue:                     true,
	phrase.ArrayKey:                       true,
	phrase.LogicalExpression:              true,
	phrase.EqualityExpression:             true,
	phrase.ForeachStatement:               true,
	phrase.ForeachCollection:              true,
	phrase.ForStatement:                   true,
	phrase.UnaryOpExpression:              true,
	phrase.ThrowStatement:                 true,
	phrase.ElseIfClauseList:               true,
	phrase.ElseIfClause:                   true,
	phrase.TernaryExpression:              true,
	phrase.SubscriptExpression:            true,
	phrase.EmptyIntrinsic:                 true,
	phrase.UnsetIntrinsic:                 true,
	phrase.IssetIntrinsic:                 true,
	phrase.EvalIntrinsic:                  true,
	phrase.VariableList:                   true,
	phrase.TraitMemberDeclarationList:     true,
}

var /* const */ skipAddingSymbol map[phrase.PhraseType]bool = map[phrase.PhraseType]bool{
	phrase.ArgumentExpressionList: true,
}
var /*const */ tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
	// Expressions
	lexer.DirectoryConstant: newDirectoryConstantAccess,
	lexer.DocumentComment:   newPhpDocFromNode,
}

func scanForChildren(document *Document, node *phrase.Phrase) {
	var phraseToSymbolConstructor = map[phrase.PhraseType]symbolConstructorForPhrase{
		// Symbols
		phrase.InterfaceDeclaration:       newInterface,
		phrase.ClassDeclaration:           newClass,
		phrase.FunctionDeclaration:        newFunction,
		phrase.ClassConstElement:          newClassConst,
		phrase.ConstDeclaration:           newConstDeclaration,
		phrase.ConstElement:               newConst,
		phrase.ArgumentExpressionList:     newArgumentList,
		phrase.TraitDeclaration:           newTrait,
		phrase.MethodDeclaration:          newMethod,
		phrase.FunctionCallExpression:     tryToNewDefine,
		phrase.SimpleAssignmentExpression: newAssignment,
		phrase.PropertyDeclaration:        newPropertyDeclaration,
		phrase.GlobalDeclaration:          newGlobalDeclaration,
		phrase.NamespaceUseDeclaration:    processNamespaceUseDeclaration,
		phrase.InstanceOfExpression:       processInstanceofExpression,
		phrase.TraitUseClause:             processTraitUseClause,

		phrase.AnonymousFunctionCreationExpression: newAnonymousFunction,
	}
	for _, child := range node.Children {
		var childSymbol Symbol = nil
		shouldSkipAdding := false
		if p, ok := child.(*phrase.Phrase); ok {
			if p.Type == phrase.NamespaceDefinition {
				namespace := newNamespace(document, p)
				document.setNamespace(namespace)
				continue
			}

			scanForExpression(document, p)
			if _, ok := scanPhraseTypes[p.Type]; ok {
				scanForChildren(document, p)
				continue
			}
			if constructor, ok := phraseToSymbolConstructor[p.Type]; ok {
				childSymbol = constructor(document, p)
			}
			if _, ok := skipAddingSymbol[p.Type]; ok {
				shouldSkipAdding = true
			}
		} else if t, ok := child.(*lexer.Token); ok {
			if constructor, ok := tokenToSymbolConstructor[t.Type]; ok {
				childSymbol = constructor(document, t)
			}
		}

		if !shouldSkipAdding && childSymbol != nil {
			document.addSymbol(childSymbol)
		}
	}
}
