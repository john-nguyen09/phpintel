package analysis

import (
	sitter "github.com/smacker/go-tree-sitter"
)

func processBinaryExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
	op := node.ChildByFieldName("operator")
	switch op.Type() {
	case "instanceof":
		lhs := node.ChildByFieldName("left")
		rhs := node.ChildByFieldName("right")
		if lhs != nil && rhs != nil {
			lhsExpr := scanForExpression(document, lhs)
			rhsExpr := scanForExpression(document, rhs)
			if c, ok := lhsExpr.(CanAddType); ok && rhsExpr != nil {
				c.AddTypes(rhsExpr.GetTypes())
			}
		}
	}
	return nil, false
}
