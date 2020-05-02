package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newInstanceOfTypeDesignator(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.QualifiedName, phrase.FullyQualifiedName:
				return newClassAccess(document, p), true
			}
		}
		child = traverser.Advance()
	}
	return nil, false
}

func processInstanceOfExpression(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	lhs := traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	traverser.SkipToken(lexer.InstanceOf)
	traverser.SkipToken(lexer.Whitespace)
	rhs := traverser.Advance()
	if lhs, ok1 := lhs.(*phrase.Phrase); ok1 {
		if rhs, ok2 := rhs.(*phrase.Phrase); ok2 {
			lhsExpr := scanForExpression(document, lhs)
			rhsExpr := scanForExpression(document, rhs)
			if c, ok := lhsExpr.(CanAddType); ok && rhsExpr != nil {
				c.AddTypes(rhsExpr.GetTypes())
			}
		}
	}
	return nil, false
}

// func processBinaryExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
// 	op := node.ChildByFieldName("operator")
// 	if op == nil {
// 		return nil, false
// 	}
// 	switch op.Type() {
// 	case "instanceof":
// 		lhs := node.ChildByFieldName("left")
// 		rhs := node.ChildByFieldName("right")
// 		if lhs != nil && rhs != nil {
// 			lhsExpr := scanForExpression(document, lhs)
// 			rhsExpr := scanForExpression(document, rhs)
// 			if c, ok := lhsExpr.(CanAddType); ok && rhsExpr != nil {
// 				c.AddTypes(rhsExpr.GetTypes())
// 			}
// 		}
// 	case ".":
// 		lhs := node.ChildByFieldName("left")
// 		rhs := node.ChildByFieldName("right")
// 		if lhs != nil && rhs != nil {
// 			scanForExpression(document, lhs)
// 			scanForExpression(document, rhs)
// 		}
// 	default:
// 		scanForChildren(document, node)
// 	}
// 	return nil, false
// }
