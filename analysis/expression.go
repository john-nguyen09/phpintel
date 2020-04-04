package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/ast"
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

type HasName interface {
	GetName() string
}

type CanAddType interface {
	AddTypes(types TypeComposite)
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

type exprConstructor func(*Document, *ast.Node) (HasTypes, bool)

var nodeTypeToExprConstructor map[string]exprConstructor

func init() {
	nodeTypeToExprConstructor = map[string]exprConstructor{
		"function_call_expression":          newFunctionCall,
		"qualified_name":                    processQualifiedName,
		"scoped_property_access_expression": newScopedPropertyAccess,
		"scoped_call_expression":            newScopedMethodAccess,
		"class_constant_access_expression":  newScopedConstantAccess,
		"object_creation_expression":        newClassTypeDesignator,
		"variable_name":                     newVariableExpression,
		"member_access_expression":          newPropertyAccess,
		"member_call_expression":            newMethodAccess,
		"parenthesized_expression":          newDerivedExpression,
		"clone_expression":                  newDerivedExpression,
		"binary_expression":                 processBinaryExpression,
		"unary_op_expression":               processToScanChildren,
		"assignment_expression":             newAssignment,
	}
}

func scanForExpression(document *Document, node *ast.Node) HasTypes {
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
	if constructor, ok := nodeTypeToExprConstructor[node.Type()]; ok {
		expression, shouldAdd = constructor(document, node)
	}
	return expression
}

func processToScanChildren(document *Document, node *ast.Node) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		switch child.Type() {
		case "@", "+", "-", "~", "!":
			child = traverser.Advance()
			continue
		}
		scanForExpression(document, child)
		child = traverser.Advance()
	}
	return nil, false
}

type derivedExpression struct {
	Expression
	hasResolved bool
}

var _ HasTypes = (*derivedExpression)(nil)

func newDerivedExpression(document *Document, node *ast.Node) (HasTypes, bool) {
	derivedExpr := &derivedExpression{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	document.addSymbol(derivedExpr)
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		expr := scanForExpression(document, child)
		if expr != nil {
			derivedExpr.Scope = expr
			break
		}
		child = traverser.Advance()
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
