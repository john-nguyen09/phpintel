package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/indexer"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/sourcegraph/go-lsp"
)

// Expression represents a reference
type Expression struct {
	Type     TypeComposite
	Scope    *Expression
	Location lsp.Location
	Name     string
}

type hasTypes interface {
	getTypes() TypeComposite
}

type expressionConstructorForPhrase func(*Document, *phrase.Phrase) hasTypes

var /* const */ skipPhraseTypes = map[phrase.PhraseType]bool{
	phrase.ObjectCreationExpression: true,
}

func scanForExpression(document *Document, node *phrase.Phrase) hasTypes {
	var phraseToExpressionConstructor = map[phrase.PhraseType]expressionConstructorForPhrase{
		phrase.FunctionCallExpression:         newFunctionCall,
		phrase.ConstantAccessExpression:       newConstantAccess,
		phrase.ScopedPropertyAccessExpression: newScopedPropertyAccess,
		phrase.ScopedCallExpression:           newScopedMethodAccess,
		phrase.ClassConstantAccessExpression:  newScopedConstantAccess,
		phrase.ClassTypeDesignator:            newClassTypeDesignator,
		phrase.ObjectCreationExpression:       newClassTypeDesignator,
		phrase.SimpleVariable:                 newVariableExpression,
	}
	var expression hasTypes = nil
	defer func() {
		if symbol, ok := expression.(Symbol); ok {
			document.addSymbol(symbol)
		}
	}()
	if _, ok := skipPhraseTypes[node.Type]; ok {
		for _, child := range node.Children {
			if p, ok := child.(*phrase.Phrase); ok {
				expression = scanForExpression(document, p)
				return expression
			}
		}
	}
	if constructor, ok := phraseToExpressionConstructor[node.Type]; ok {
		expression = constructor(document, node)
	}
	return expression
}

func (s *Expression) Write(serialiser *indexer.Serialiser) {
	s.Type.Write(serialiser)
	if s.Scope == nil {
		serialiser.WriteBool(false)
	} else {
		serialiser.WriteBool(true)
		s.Scope.Write(serialiser)
	}
	util.WriteLocation(serialiser, s.Location)
	serialiser.WriteString(s.Name)
}

func ReadExpression(serialiser *indexer.Serialiser) Expression {
	expr := Expression{
		Type: ReadTypeComposite(serialiser),
	}
	if serialiser.ReadBool() {
		scope := ReadExpression(serialiser)
		expr.Scope = &scope
	}
	expr.Location = util.ReadLocation(serialiser)
	expr.Name = serialiser.ReadString()
	return expr
}
