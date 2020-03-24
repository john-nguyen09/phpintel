package analysis

import (
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
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
	GetLocation() protocol.Location
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

type exprConstructor func(*Document, *sitter.Node) (HasTypes, bool)

var nodeTypeToExprConstructor map[string]exprConstructor

var /* const */ skipPhraseTypes = map[string]bool{}

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
		"parenthesized_expression":          newParenthesised,
		"clone_expression":                  newCloneExpression,
	}
}

func scanForExpression(document *Document, node *sitter.Node) HasTypes {
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
	if _, ok := skipPhraseTypes[node.Type()]; ok {
		childCount := int(node.ChildCount())
		for i := 0; i < childCount; i++ {
			child := node.Child(i)
			return scanForExpression(document, child)
		}
	}
	if constructor, ok := nodeTypeToExprConstructor[node.Type()]; ok {
		expression, shouldAdd = constructor(document, node)
	}
	return expression
}

type cloneExpression struct {
	Expression
	hasResolved bool
}

var _ HasTypes = (*cloneExpression)(nil)

func newCloneExpression(document *Document, node *sitter.Node) (HasTypes, bool) {
	cloneExpr := &cloneExpression{
		Expression: Expression{
			Location: document.GetNodeLocation(node),
		},
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		expr := scanForExpression(document, child)
		if expr != nil {
			cloneExpr.Scope = expr
			break
		}
		child = traverser.Advance()
	}
	return cloneExpr, true
}

func (s *cloneExpression) GetLocation() protocol.Location {
	return s.Location
}

func (s *cloneExpression) GetTypes() TypeComposite {
	if s.Scope != nil {
		return s.Scope.GetTypes()
	}
	return s.Type
}

func (s *cloneExpression) Resolve(ctx ResolveContext) {
	if s.hasResolved {
		return
	}
	s.hasResolved = true
	if s.Scope != nil {
		s.Scope.Resolve(ctx)
	}
}
