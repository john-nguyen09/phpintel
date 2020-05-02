package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type symbolConstructor func(*Document, *phrase.Phrase) Symbol
type symbolConstructorForToken func(*Document, *lexer.Token) Symbol

type void = struct{}

var empty void

var /* const */ typesToScanForChildren = map[phrase.PhraseType]void{
	phrase.ExpressionStatement:            empty,
	phrase.WhileStatement:                 empty,
	phrase.ClassMemberDeclarationList:     empty,
	phrase.InterfaceMemberDeclarationList: empty,
	phrase.ClassConstElementList:          empty,
	phrase.ClassConstDeclaration:          empty,
	phrase.EncapsulatedExpression:         empty,
	phrase.CompoundStatement:              empty,
	phrase.StatementList:                  empty,
	phrase.AdditiveExpression:             empty,
	phrase.IfStatement:                    empty,
	phrase.ElseClause:                     empty,
	phrase.IncludeExpression:              empty,
	phrase.EchoIntrinsic:                  empty,
	phrase.ExpressionList:                 empty,
	phrase.ClassDeclarationBody:           empty,
	phrase.TryStatement:                   empty,
	phrase.CatchClauseList:                empty,
	phrase.CatchClause:                    empty,
	phrase.ReturnStatement:                empty,
	phrase.ObjectCreationExpression:       empty,
	phrase.ScopedCallExpression:           empty,
	phrase.ArrayCreationExpression:        empty,
	phrase.ArrayInitialiserList:           empty,
	phrase.ArrayElement:                   empty,
	phrase.ArrayValue:                     empty,
	phrase.ArrayKey:                       empty,
	phrase.LogicalExpression:              empty,
	phrase.RelationalExpression:           empty,
	phrase.EqualityExpression:             empty,
	phrase.ForStatement:                   empty,
	phrase.UnaryOpExpression:              empty,
	phrase.ThrowStatement:                 empty,
	phrase.ElseIfClauseList:               empty,
	phrase.ElseIfClause:                   empty,
	phrase.TernaryExpression:              empty,
	phrase.SubscriptExpression:            empty,
	phrase.EmptyIntrinsic:                 empty,
	phrase.UnsetIntrinsic:                 empty,
	phrase.IssetIntrinsic:                 empty,
	phrase.EvalIntrinsic:                  empty,
	phrase.VariableList:                   empty,
	phrase.TraitMemberDeclarationList:     empty,
	phrase.CastExpression:                 empty,
	phrase.SwitchStatement:                empty,
	phrase.CaseStatementList:              empty,
	phrase.CaseStatement:                  empty,
}

var /*const */ tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
	lexer.DirectoryConstant: newDirectoryConstantAccess,
}
var symbolConstructorMap map[phrase.PhraseType]symbolConstructor

func init() {
	symbolConstructorMap = map[phrase.PhraseType]symbolConstructor{
		phrase.InterfaceDeclaration:                newInterface,
		phrase.ClassDeclaration:                    newClass,
		phrase.PropertyDeclaration:                 newPropertyDeclaration,
		phrase.MethodDeclaration:                   newMethod,
		phrase.TraitUseClause:                      processTraitUseClause,
		phrase.FunctionDeclaration:                 newFunction,
		phrase.ConstDeclaration:                    newConstDeclaration,
		phrase.ConstElement:                        newConst,
		phrase.ClassConstElement:                   newClassConst,
		phrase.ArgumentExpressionList:              newArgumentList,
		phrase.TraitDeclaration:                    newTrait,
		phrase.FunctionCallExpression:              tryToNewDefine,
		phrase.GlobalDeclaration:                   newGlobalDeclaration,
		phrase.NamespaceUseDeclaration:             processNamespaceUseDeclaration,
		phrase.AnonymousFunctionCreationExpression: newAnonymousFunction,
		phrase.DocumentComment:                     newPhpDocFromNode,
	}
}

func scanNode(document *Document, node phrase.AstNode) {
	var symbol Symbol = nil
	shouldSkipAdding := false

	if p, ok := node.(*phrase.Phrase); ok {
		if p.Type == phrase.NamespaceDefinition {
			newNamespace(document, p)
			return
		}

		scanForExpression(document, p)
		if _, ok := typesToScanForChildren[p.Type]; ok {
			scanForChildren(document, p)
			return
		}
		if constructor, ok := symbolConstructorMap[p.Type]; ok {
			symbol = constructor(document, p)
		}
		switch p.Type {
		case phrase.ArgumentExpressionList:
			shouldSkipAdding = true
		}
	} else if t, ok := node.(*lexer.Token); ok {
		if constructor, ok := tokenToSymbolConstructor[t.Type]; ok {
			symbol = constructor(document, t)
		}
	}

	if !shouldSkipAdding && symbol != nil {
		document.addSymbol(symbol)
	}
}

func scanForChildren(document *Document, node *phrase.Phrase) {
	for _, child := range node.Children {
		scanNode(document, child)
	}
}
