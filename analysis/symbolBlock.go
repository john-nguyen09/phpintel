package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type symbolConstructor func(analyser, *Document, *phrase.Phrase) Symbol
type symbolConstructorForToken func(analyser, *Document, *lexer.Token) Symbol

type analyser struct {
	nodes util.NodeStack
}

func newAnalyser() analyser {
	return analyser{}
}

var /* const */ typesToScanForChildren = util.SetFromArray([]phrase.PhraseType{
	phrase.ExpressionStatement,
	phrase.ClassMemberDeclarationList,
	phrase.InterfaceMemberDeclarationList,
	phrase.TraitMemberDeclarationList,
	phrase.CompoundStatement,
	phrase.WhileStatement,
	phrase.StatementList,
	phrase.AdditiveExpression,
	phrase.MultiplicativeExpression,
	phrase.IfStatement,
	phrase.ElseClause,
	phrase.IncludeExpression,
	phrase.EchoIntrinsic,
	phrase.ExpressionList,
	phrase.TryStatement,
	phrase.CatchClauseList,
	phrase.CatchClause,
	phrase.ReturnStatement,
	phrase.ArrayCreationExpression,
	phrase.ArrayInitialiserList,
	phrase.ArrayElement,
	phrase.ArrayValue,
	phrase.ArrayKey,
	phrase.LogicalExpression,
	phrase.RelationalExpression,
	phrase.EqualityExpression,
	phrase.ForStatement,
	phrase.UnaryOpExpression,
	phrase.ThrowStatement,
	phrase.ElseIfClauseList,
	phrase.ElseIfClause,
	phrase.TernaryExpression,
	phrase.SubscriptExpression,
	phrase.EmptyIntrinsic,
	phrase.UnsetIntrinsic,
	phrase.IssetIntrinsic,
	phrase.EvalIntrinsic,
	phrase.VariableList,
	phrase.CastExpression,
	phrase.SwitchStatement,
	phrase.CaseStatementList,
	phrase.CaseStatement,
	phrase.DefaultStatement,
	phrase.ClassConstElementList,
	phrase.PostfixIncrementExpression,
	phrase.PostfixDecrementExpression,
	phrase.PrefixIncrementExpression,
	phrase.PrefixDecrementExpression,
	phrase.ForInitialiser,
	phrase.ForControl,
	phrase.ForEndOfLoop,
	phrase.DoStatement,
	phrase.DoubleQuotedStringLiteral,
	phrase.EncapsulatedVariableList,
	phrase.EncapsulatedVariable,
	phrase.RequireOnceExpression,
	phrase.ErrorControlExpression,
})

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

		if !isDefine(document, p) {
			scanForExpression(a, document, p)
		}
		if typesToScanForChildren.Has(p.Type) {
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
