package analysis

import (
	"strings"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

func newAssignment(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	haveProcessed := false
	if p, ok := firstChild.(*phrase.Phrase); ok {
		if p.Type == phrase.SimpleVariable {
			analyseVariableAssignment(a, document, p, traverser.Clone(), node)
			haveProcessed = true
		}
		if p.Type == phrase.PropertyAccessExpression {
			block := document.currentBlock()
			if method, ok := block.(*Method); ok && strings.ToLower(method.Name) == "__construct" {
				processPropertyAccessAssignment(a, document, p, node)
				haveProcessed = true
			}
		}
	}
	if !haveProcessed {
		if p, ok := firstChild.(*phrase.Phrase); ok {
			scanNode(a, document, p)
		}
		for child := traverser.Advance(); child != nil; child = traverser.Advance() {
			scanNode(a, document, child)
		}
	}
	return nil, false
}

func analyseVariableAssignment(a analyser, document *Document, lhs *phrase.Phrase, traverser *util.Traverser, parent *phrase.Phrase) {
	traverser.Advance()
	traverser.SkipToken(lexer.Whitespace)
	if parent.Type == phrase.CompoundAssignmentExpression {
		traverser.SkipToken(lexer.DotEquals)
	} else {
		traverser.SkipToken(lexer.Equals)
	}
	traverser.SkipToken(lexer.Whitespace)
	if parent.Type == phrase.ByRefAssignmentExpression {
		traverser.SkipToken(lexer.Ampersand)
		traverser.SkipToken(lexer.Whitespace)
	}
	rhs := traverser.Advance()
	phpDoc := document.getValidPhpDoc(document.GetNodeLocation(lhs))
	variable := newVariableWithoutPushing(document, lhs)
	if phpDoc != nil {
		variable.applyPhpDoc(document, *phpDoc)
	}
	// The variable appears before rhs therefore being added before rhs
	document.addSymbol(variable)

	var expression HasTypes = nil
	isExprType := false
	if p, ok := rhs.(*phrase.Phrase); ok {
		expression = scanForExpression(a, document, p)
		_, isExprType = nodeTypeToExprConstructor[p.Type]
	}
	// But the variable should be pushed after any rhs's variables
	document.pushVariable(a, variable, document.NodeRange(parent).End, true)
	if expression != nil {
		variable.setExpression(expression)
	} else if !isExprType {
		scanNode(a, document, rhs)
	}
	globalVariable := document.getGlobalVariable(variable.Name)
	if globalVariable != nil {
		types := variable.GetTypes()
		if !types.IsEmpty() {
			globalVariable.types.merge(types)
		}
	}
}

func processPropertyAccessAssignment(a analyser, document *Document, node *phrase.Phrase, parent *phrase.Phrase) {
	var (
		lhsExpr HasTypes
		rhsExpr HasTypes
		prop    *Property
	)
	traverser := util.NewTraverser(node)
	firstChild := traverser.Advance()
	parentTraverser := util.NewTraverser(parent)
	lhs := parentTraverser.Advance()
	if p, ok := lhs.(*phrase.Phrase); ok {
		lhsExpr = scanForExpression(a, document, p)
	}
	parentTraverser.SkipToken(lexer.Whitespace)
	parentTraverser.SkipToken(lexer.Equals)
	parentTraverser.SkipToken(lexer.Whitespace)
	rhs := parentTraverser.Advance()
	if rhsP, ok := rhs.(*phrase.Phrase); ok {
		rhsExpr = scanForExpression(a, document, rhsP)
	}
	if rhsExpr == nil {
		scanNode(a, document, rhs)
		return
	}
	if p, ok1 := firstChild.(*phrase.Phrase); ok1 && p.Type == phrase.SimpleVariable &&
		document.getPhraseText(p) == "$this" {
		if propAccess, ok2 := lhsExpr.(*PropertyAccess); ok2 && propAccess != nil {
			class := document.getLastClass()
			switch v := class.(type) {
			case *Class:
				prop = v.findProp(propAccess.MemberName())
			case *Interface:
				prop = v.findProp(propAccess.MemberName())
			case *Trait:
				prop = v.findProp(propAccess.MemberName())
			}
		}
	}
	if prop != nil && rhsExpr != nil {
		prop.Types.merge(rhsExpr.GetTypes())
	}
}
