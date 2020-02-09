package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
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

type expressionConstructorForPhrase func(*Document, *phrase.Phrase) (HasTypes, bool)

var phraseToExpressionConstructor map[phrase.PhraseType]expressionConstructorForPhrase

var /* const */ skipPhraseTypes = map[phrase.PhraseType]bool{}

func init() {
	phraseToExpressionConstructor = map[phrase.PhraseType]expressionConstructorForPhrase{
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
		phrase.EncapsulatedExpression:         analyseEncapsulatedExpression,
	}
}

func scanForExpression(document *Document, node *phrase.Phrase) HasTypes {
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
	if _, ok := skipPhraseTypes[node.Type]; ok {
		for _, child := range node.Children {
			if p, ok := child.(*phrase.Phrase); ok {
				return scanForExpression(document, p)
			}
		}
	}
	if constructor, ok := phraseToExpressionConstructor[node.Type]; ok {
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
