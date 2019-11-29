package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
)

type symbolConstructorForPhrase func(*Document, *phrase.Phrase) Symbol
type symbolConstructorForToken func(*Document, *lexer.Token) Symbol

var /* const */ scanPhraseTypes = map[phrase.PhraseType]bool{
	phrase.ExpressionStatement:        true,
	phrase.WhileStatement:             true,
	phrase.ClassMemberDeclarationList: true,
	phrase.ClassConstElementList:      true,
	phrase.ClassConstDeclaration:      true,
	phrase.EncapsulatedExpression:     true,
	phrase.CompoundStatement:          true,
	phrase.StatementList:              true,
	phrase.AdditiveExpression:         true,
	phrase.IfStatement:                true,
	phrase.IncludeExpression:          true,
	phrase.EchoIntrinsic:              true,
	phrase.ExpressionList:             true,
	phrase.ClassDeclarationBody:       true,
	phrase.TryStatement:               true,
	phrase.CatchClauseList:            true,
	phrase.CatchClause:                true,
	phrase.ReturnStatement:            true,
	phrase.ObjectCreationExpression:   true,
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
	}
	var tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
		// Expressions
		lexer.DirectoryConstant: newDirectoryConstantAccess,
		lexer.DocumentComment:   newPhpDocFromNode,
	}
	for _, child := range node.Children {
		var childSymbol Symbol = nil
		if p, ok := child.(*phrase.Phrase); ok {
			scanForExpression(document, p)
			if _, ok := scanPhraseTypes[p.Type]; ok {
				scanForChildren(document, p)
				continue
			}
			if constructor, ok := phraseToSymbolConstructor[p.Type]; ok {
				childSymbol = constructor(document, p)
			}
		} else if t, ok := child.(*lexer.Token); ok {
			if constructor, ok := tokenToSymbolConstructor[t.Type]; ok {
				childSymbol = constructor(document, t)
			}
		}

		if childSymbol != nil {
			document.addSymbol(childSymbol)
		}
	}
}
