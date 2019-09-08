package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
)

type SymbolBlock interface {
	GetDocument() *Document
	GetChildren() []Symbol
}

func ScanForChildren(s SymbolBlock, node *phrase.Phrase) {
	var nodeToSymbolConstructor = map[phrase.PhraseType]interface{}{
		phrase.InterfaceDeclaration:   NewInterface,
		phrase.ClassDeclaration:       NewClass,
		phrase.FunctionDeclaration:    NewFunction,
		phrase.ClassConstDeclaration:  NewClassConstDeclaration,
		phrase.ConstDeclaration:       NewConstDeclaration,
		phrase.ConstElement:           NewConst,
		phrase.FunctionCallExpression: NewFunctionCall,
		phrase.ExpressionStatement:    ProcessExpressionStatement,
		phrase.ArgumentExpressionList: NewArgumentList,
	}

	for _, child := range node.Children {
		if p, ok := child.(*phrase.Phrase); ok {
			if constructor, ok := nodeToSymbolConstructor[p.Type]; ok {
				if block, ok := s.(SymbolBlock); ok {
					childSymbol := constructor.(func(*Document, SymbolBlock, *phrase.Phrase) Symbol)(
						block.GetDocument(), s, p)

					if childSymbol == nil {
						continue
					}

					if consumer, ok := s.(HasConsume); ok {
						consumer.Consume(childSymbol)
					}
				}
			}
		}
	}
}
