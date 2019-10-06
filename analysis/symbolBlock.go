package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"

	"github.com/john-nguyen09/go-phpparser/phrase"
)

type symbolBlock interface {
	getDocument() *Document
}

type symbolConstructorForPhrase func(*Document, symbolBlock, *phrase.Phrase) Symbol
type symbolConstructorForToken func(*Document, symbolBlock, *lexer.Token) Symbol

var /* const */ scanPhraseTypes = map[phrase.PhraseType]bool{
	phrase.ExpressionStatement:        true,
	phrase.SimpleAssignmentExpression: true,
	phrase.WhileStatement:             true,
	phrase.ClassMemberDeclarationList: true,
	phrase.ClassConstElementList:      true,
	phrase.ClassConstDeclaration:      true,
	phrase.EncapsulatedExpression:     true,
	phrase.CompoundStatement:          true,
	phrase.StatementList:              true,
	phrase.FunctionCallExpression:     true,
	phrase.AdditiveExpression:         true,
	phrase.IfStatement:                true,
	phrase.IncludeExpression:          true,
	phrase.EchoIntrinsic:              true,
	phrase.ExpressionList:             true,
}

func scanForChildren(s symbolBlock, node *phrase.Phrase) {
	var phraseToSymbolConstructor = map[phrase.PhraseType]symbolConstructorForPhrase{
		// Symbols
		phrase.InterfaceDeclaration:   newInterface,
		phrase.ClassDeclaration:       newClass,
		phrase.FunctionDeclaration:    newFunction,
		phrase.ClassConstElement:      newClassConst,
		phrase.ConstDeclaration:       newConstDeclaration,
		phrase.ConstElement:           newConst,
		phrase.ArgumentExpressionList: newArgumentList,
		phrase.TraitDeclaration:       newTrait,
		phrase.MethodDeclaration:      newMethod,
		// Expressions
		phrase.FunctionCallExpression:         newFunctionCall,
		phrase.ConstantAccessExpression:       newConstantAccess,
		phrase.ScopedPropertyAccessExpression: newScopedPropertyAccess,
		phrase.ScopedCallExpression:           newScopedMethodAccess,
		phrase.ClassConstantAccessExpression:  newScopedConstantAccess,
	}
	var tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
		// Expressions
		lexer.DirectoryConstant: newDirectoryConstantAccess,
	}
	for _, child := range node.Children {
		var childSymbol Symbol = nil
		if p, ok := child.(*phrase.Phrase); ok {
			// // Following lines are for debugging
			// fmt.Println(p.Type.String() + ": " + util.GetNodeText(p, s.getDocument().text))
			// jsonData, err := json.MarshalIndent(p.Children, "", "  ")
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Println(string(jsonData))
			if _, ok := scanPhraseTypes[p.Type]; ok {
				scanForChildren(s, p)
				continue
			}
			if constructor, ok := phraseToSymbolConstructor[p.Type]; ok {
				childSymbol = constructor(s.getDocument(), s, p)
			}
		} else if t, ok := child.(*lexer.Token); ok {
			if constructor, ok := tokenToSymbolConstructor[t.Type]; ok {
				childSymbol = constructor(s.getDocument(), s, t)
			}
		}
		if childSymbol != nil {
			consumeIfIsConsumer(s, childSymbol)
		}
	}
}

func consumeIfIsConsumer(parent symbolBlock, symbol Symbol) {
	if consumer, ok := parent.(hasConsume); ok {
		consumer.consume(symbol)
	}
}
