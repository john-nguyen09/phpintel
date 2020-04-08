package analysis

import sitter "github.com/smacker/go-tree-sitter"

func processBinaryExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
	op := node.ChildByFieldName("operator")
	if op == nil {
		return nil, false
	}
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
	case ".":
		lhs := node.ChildByFieldName("left")
		rhs := node.ChildByFieldName("right")
		if lhs != nil && rhs != nil {
			scanForExpression(document, lhs)
			scanForExpression(document, rhs)
		}
	default:
		scanForChildren(document, node)
	}
	return nil, false
}
