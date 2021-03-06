package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type symbolConstructor func(analyser, *Document, *phrase.Phrase) Symbol
type symbolConstructorForToken func(analyser, *Document, *lexer.Token) Symbol

type void = struct{}

var empty void

type analyser struct {
	nodes util.NodeStack
}

func newAnalyser() analyser {
	return analyser{}
}

var /* const */ typesToScanForChildren = map[phrase.PhraseType]void{
	phrase.ExpressionStatement:            empty,
	phrase.ClassMemberDeclarationList:     empty,
	phrase.InterfaceMemberDeclarationList: empty,
	phrase.TraitMemberDeclarationList:     empty,
	phrase.CompoundStatement:              empty,
	phrase.WhileStatement:                 empty,
	phrase.StatementList:                  empty,
	phrase.AdditiveExpression:             empty,
	phrase.MultiplicativeExpression:       empty,
	phrase.IfStatement:                    empty,
	phrase.ElseClause:                     empty,
	phrase.IncludeExpression:              empty,
	phrase.EchoIntrinsic:                  empty,
	phrase.ExpressionList:                 empty,
	phrase.TryStatement:                   empty,
	phrase.CatchClauseList:                empty,
	phrase.CatchClause:                    empty,
	phrase.ReturnStatement:                empty,
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
	phrase.CastExpression:                 empty,
	phrase.SwitchStatement:                empty,
	phrase.CaseStatementList:              empty,
	phrase.CaseStatement:                  empty,
	phrase.DefaultStatement:               empty,
	phrase.ClassConstElementList:          empty,
	phrase.PostfixIncrementExpression:     empty,
	phrase.PostfixDecrementExpression:     empty,
	phrase.PrefixIncrementExpression:      empty,
	phrase.PrefixDecrementExpression:      empty,
	phrase.ForInitialiser:                 empty,
	phrase.ForControl:                     empty,
	phrase.ForEndOfLoop:                   empty,
	phrase.DoStatement:                    empty,
	phrase.DoubleQuotedStringLiteral:      empty,
	phrase.EncapsulatedVariableList:       empty,
	phrase.RequireOnceExpression:          empty,
}

var /*const */ tokenToSymbolConstructor = map[lexer.TokenType]symbolConstructorForToken{
	lexer.DirectoryConstant: newDirectoryConstantAccess,
}
var symbolConstructorMap map[phrase.PhraseType]symbolConstructor

func init() {
	symbolConstructorMap = map[phrase.PhraseType]symbolConstructor{
		phrase.InterfaceDeclaration:                newInterface,
		phrase.ClassDeclaration:                    newClass,
		phrase.ClassConstDeclaration:               newClassConstDeclaration,
		phrase.PropertyDeclaration:                 newPropertyDeclaration,
		phrase.MethodDeclaration:                   newMethod,
		phrase.TraitUseClause:                      processTraitUseClause,
		phrase.FunctionDeclaration:                 newFunction,
		phrase.ConstDeclaration:                    newConstDeclaration,
		phrase.ConstElement:                        newConst,
		phrase.ArgumentExpressionList:              newArgumentList,
		phrase.TraitDeclaration:                    newTrait,
		phrase.FunctionCallExpression:              tryToNewDefine,
		phrase.GlobalDeclaration:                   newGlobalDeclaration,
		phrase.NamespaceUseDeclaration:             processNamespaceUseDeclaration,
		phrase.AnonymousFunctionCreationExpression: newAnonymousFunction,
		phrase.AnonymousClassDeclaration:           newAnonymousClass,
		phrase.DocumentComment:                     newPhpDocFromNode,
		phrase.CatchNameList:                       processCatchNameList,
	}
}

func scanNode(a analyser, document *Document, node phrase.AstNode) {
	var symbol Symbol = nil
	shouldSkipAdding := false

	if p, ok := node.(*phrase.Phrase); ok {
		if p.Type == phrase.NamespaceDefinition {
			newNamespace(a, document, p)
			return
		}

		scanForExpression(a, document, p)
		if _, ok := typesToScanForChildren[p.Type]; ok {
			scanForChildren(a, document, p)
			return
		}
		if constructor, ok := symbolConstructorMap[p.Type]; ok {
			symbol = constructor(a, document, p)
		}
		switch p.Type {
		case phrase.ArgumentExpressionList:
			shouldSkipAdding = true
		}
	} else if t, ok := node.(*lexer.Token); ok {
		if constructor, ok := tokenToSymbolConstructor[t.Type]; ok {
			symbol = constructor(a, document, t)
		}
	}

	if !shouldSkipAdding && symbol != nil {
		document.addSymbol(symbol)
	}
}

func scanForChildren(a analyser, document *Document, node *phrase.Phrase) {
	a.nodes.Push(node)
	for _, child := range node.Children {
		scanNode(a, document, child)
	}
	a.nodes.Pop()
}
