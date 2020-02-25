package analysis

// func processInstanceofExpression(document *Document, node *phrase.Phrase) Symbol {
// 	traverser := util.NewTraverser(node)
// 	child := traverser.Advance()
// 	var lhs HasTypes = nil
// 	beforeInstanceOf := true
// 	for child != nil {
// 		if p, ok := child.(*phrase.Phrase); ok {
// 			switch p.Type {
// 			case phrase.InstanceofTypeDesignator:
// 				typeDeclaration := newTypeDeclaration(document, p)
// 				document.addSymbol(typeDeclaration)

// 				if canAddType, ok := lhs.(CanAddType); ok {
// 					canAddType.AddTypes(typeDeclaration.GetTypes())
// 				}
// 			default:
// 				if beforeInstanceOf {
// 					lhs = scanForExpression(document, p)
// 				}
// 			}
// 		}
// 		child = traverser.Advance()
// 	}
// 	return nil
// }
