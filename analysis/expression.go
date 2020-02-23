package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
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
		"qualified_name":                    newConstantAccess,
		"scoped_property_access_expression": newScopedPropertyAccess,
		"scoped_call_expression":            newScopedMethodAccess,
		"class_constant_access_expression":  newScopedConstantAccess,
		"object_creation_expression":        newClassTypeDesignator,
		"variable_name":                     newVariableExpression,
		"member_access_expression":          newPropertyAccess,
		"member_call_expression":            newMethodAccess,
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

func (s *Expression) Serialise(e *storage.Encoder) {
	s.Type.Write(e)
	switch expression := s.Scope.(type) {
	case *Variable:
		e.WriteBool(true)
		e.WriteInt(int(variableKind))
		expression.Serialise(e)
	case *ClassAccess:
		e.WriteBool(true)
		e.WriteInt(int(classAccessKind))
		expression.Serialise(e)
	case *ClassTypeDesignator:
		e.WriteBool(true)
		e.WriteInt(int(classTypeDesignatorKind))
		expression.Serialise(e)
	case *ConstantAccess:
		e.WriteBool(true)
		e.WriteInt(int(constantAccessKind))
		expression.Serialise(e)
	case *FunctionCall:
		e.WriteBool(true)
		e.WriteInt(int(functionCallKind))
		expression.Serialise(e)
	case *PropertyAccess:
		e.WriteBool(true)
		e.WriteInt(int(propertyAccessKind))
		expression.Serialise(e)
	case *ScopedConstantAccess:
		e.WriteBool(true)
		e.WriteInt(int(scopedConstantAccessKind))
		expression.Serialise(e)
	case *ScopedMethodAccess:
		e.WriteBool(true)
		e.WriteInt(int(scopedMethodAccessKind))
		expression.Serialise(e)
	case *ScopedPropertyAccess:
		e.WriteBool(true)
		e.WriteInt(int(scopedPropertyAccessKind))
		expression.Serialise(e)
	default:
		e.WriteBool(false)
	}
	e.WriteLocation(s.Location)
	e.WriteString(s.Name)
}

func ReadExpression(d *storage.Decoder) Expression {
	expr := Expression{
		Type: ReadTypeComposite(d),
	}
	if d.ReadBool() {
		switch expressionKind(d.ReadInt()) {
		case variableKind:
			expr.Scope = ReadVariable(d)
		case classAccessKind:
			expr.Scope = ReadClassAccess(d)
		case classTypeDesignatorKind:
			expr.Scope = ReadClassTypeDesignator(d)
		case constantAccessKind:
			expr.Scope = ReadConstantAccess(d)
		case functionCallKind:
			expr.Scope = ReadFunctionCall(d)
		case propertyAccessKind:
			expr.Scope = ReadPropertyAccess(d)
		case scopedConstantAccessKind:
			expr.Scope = ReadScopedConstantAccess(d)
		case scopedMethodAccessKind:
			expr.Scope = ReadScopedMethodAccess(d)
		case scopedPropertyAccessKind:
			expr.Scope = ReadScopedPropertyAccess(d)
		}
	}
	expr.Location = d.ReadLocation()
	expr.Name = d.ReadString()
	return expr
}
