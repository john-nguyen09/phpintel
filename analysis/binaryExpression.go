package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
)

func processBinaryExpression(document *Document, node *ast.Node) (HasTypes, bool) {
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
	}
	return nil, false
}
