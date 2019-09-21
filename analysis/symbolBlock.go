package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type symbolBlock interface {
	getDocument() *Document
}

func scanForChildren(s symbolBlock, node *phrase.Phrase) {
	var nodeToSymbolConstructor = map[phrase.PhraseType]interface{}{
		phrase.InterfaceDeclaration:   newInterface,
		phrase.ClassDeclaration:       newClass,
		phrase.FunctionDeclaration:    newFunction,
		phrase.ClassConstDeclaration:  newClassConstDeclaration,
		phrase.ClassConstElement:      newClassConst,
		phrase.ConstDeclaration:       newConstDeclaration,
		phrase.ConstElement:           newConst,
		phrase.FunctionCallExpression: newFunctionCall,
		phrase.ExpressionStatement:    ProcessExpressionStatement,
		phrase.ArgumentExpressionList: newArgumentList,
		phrase.TraitDeclaration:       newTrait,
		phrase.MethodDeclaration:      newMethod,

		phrase.ClassMemberDeclarationList: classMemberDeclarationList,
		phrase.ClassConstElementList:      classConstElementList,
	}

	for _, child := range node.Children {
		if p, ok := child.(*phrase.Phrase); ok {
			if constructor, ok := nodeToSymbolConstructor[p.Type]; ok {
				if block, ok := s.(symbolBlock); ok {
					childSymbol := constructor.(func(*Document, symbolBlock, *phrase.Phrase) Symbol)(
						block.getDocument(), s, p)

					if childSymbol == nil {
						continue
					}

					if consumer, ok := s.(hasConsume); ok {
						consumer.consume(childSymbol)
					}
				}
			}
		}
	}
}
