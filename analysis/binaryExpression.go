package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newInstanceOfTypeDesignator(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				return newClassAccess(a, document, p), true
			}
		}
		child = traverser.Advance()
	}
	return nil, false
}

func processInstanceOfExpression(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	lhs := traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	traverser.SkipToken(lexer.InstanceOf)
	traverser.SkipToken(lexer.Whitespace)
	rhs := traverser.Advance()
	if lhs, ok1 := lhs.(*phrase.Phrase); ok1 {
		if rhs, ok2 := rhs.(*phrase.Phrase); ok2 {
			lhsExpr := scanForExpression(a, document, lhs)
			rhsExpr := scanForExpression(a, document, rhs)
			if v, ok := lhsExpr.(*Variable); ok && rhsExpr != nil {
				v.Type.merge(rhsExpr.GetTypes())
			}
		}
	}
	return nil, false
}
