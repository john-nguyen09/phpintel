package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

// Expression represents a reference
type Expression struct {
	Type     TypeComposite
	Scope    HasTypes
	Location protocol.Location
	Name     string
}

type ResolveContext struct {
	store    *Store
	document *Document
}

func NewResolveContext(store *Store, document *Document) ResolveContext {
	return ResolveContext{store, document}
}

func (e *Expression) Resolve(ctx ResolveContext) {

}

func (e *Expression) ResolveAndGetScope(ctx ResolveContext) TypeComposite {
	if e.Scope != nil {
		e.Scope.Resolve(ctx)
		return e.Scope.GetTypes()
	}
	return newTypeComposite()
}

type HasTypes interface {
	Symbol
	GetTypes() TypeComposite
	Resolve(ctx ResolveContext)
}

type HasTypesHasScope interface {
	HasTypes
	MemberName() string
	GetScopeTypes() TypeComposite
}

type HasName interface {
	GetName() string
}

type expressionKind int

const (
	unknownKind              = iota
	variableKind             = iota
	classAccessKind          = iota
	classTypeDesignatorKind  = iota
	constantAccessKind       = iota
	functionCallKind         = iota
	propertyAccessKind       = iota
	scopedConstantAccessKind = iota
	scopedMethodAccessKind   = iota
	scopedPropertyAccessKind = iota
)

type exprConstructor func(analyser, *Document, *phrase.Phrase) (HasTypes, bool)

var nodeTypeToExprConstructor map[phrase.PhraseType]exprConstructor

func init() {
	nodeTypeToExprConstructor = map[phrase.PhraseType]exprConstructor{
		phrase.FunctionCallExpression:         newFunctionCall,
		phrase.ConstantAccessExpression:       newConstantAccess,
		phrase.ScopedPropertyAccessExpression: newScopedPropertyAccess,
		phrase.ScopedCallExpression:           newScopedMethodAccess,
		phrase.ClassConstantAccessExpression:  newScopedConstantAccess,
		phrase.ErrorScopedAccessExpression:    newScopedConstantAccess,
		phrase.ObjectCreationExpression:       newClassTypeDesignator,
		phrase.SimpleVariable:                 newVariableExpression,
		phrase.PropertyAccessExpression:       newPropertyAccess,
		phrase.MethodCallExpression:           newMethodAccess,
		phrase.ForeachStatement:               analyseForeachStatement,
		phrase.EncapsulatedExpression:         newDerivedExpression,
		phrase.CloneExpression:                newDerivedExpression,
		phrase.SimpleAssignmentExpression:     newAssignment,
		phrase.ByRefAssignmentExpression:      newAssignment,
		phrase.CompoundAssignmentExpression:   newAssignment,
		phrase.InstanceOfExpression:           processInstanceOfExpression,
		phrase.InstanceofTypeDesignator:       newInstanceOfTypeDesignator,
	}
}

func scanForExpression(a analyser, document *Document, node *phrase.Phrase) HasTypes {
	var expression HasTypes = nil
	shouldAdd := false
	defer func() {
		if !shouldAdd {
			return
		}
		if symbol, ok := expression.(Symbol); ok {
			document.addSymbol(symbol)
		}
	}()
	if constructor, ok := nodeTypeToExprConstructor[node.Type]; ok {
		expression, shouldAdd = constructor(a, document, node)
	}
	return expression
}

func processToScanChildren(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			scanForExpression(a, document, p)
		}
		child = traverser.Advance()
	}
	return nil, false
}

type derivedExpression struct {
	Expression
	hasResolved bool
}

var _ HasTypes = (*derivedExpression)(nil)

func newDerivedExpression(a analyser, document *Document, node *phrase.Phrase) (HasTypes, bool) {
	derivedExpr := &derivedExpression{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	document.addSymbol(derivedExpr)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	nodesToScan := []phrase.AstNode{}
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			if _, ok := nodeTypeToExprConstructor[p.Type]; !ok {
				nodesToScan = append(nodesToScan, p)
			}
			expr := scanForExpression(a, document, p)
			if expr != nil {
				derivedExpr.Scope = expr
				break
			}
		} else {
			nodesToScan = append(nodesToScan, child)
		}
		child = traverser.Advance()
	}
	for _, node := range nodesToScan {
		scanNode(a, document, node)
	}
	return derivedExpr, false
}

func (s *derivedExpression) GetLocation() protocol.Location {
	return s.Location
}

func (s *derivedExpression) GetTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return s.Type
}

func (s *derivedExpression) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	if s.Scope != nil {
		s.Scope.Resolve(ctx)
	}
}
