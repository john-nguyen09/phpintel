package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

// Expression represents a reference
type Expression struct {
	Type     TypeComposite
	Scope    hasTypes
	Location protocol.Location
	Name     string
}

type hasTypes interface {
	getTypes() TypeComposite
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

type expressionConstructorForPhrase func(*Document, *phrase.Phrase) hasTypes

var /* const */ skipPhraseTypes = map[phrase.PhraseType]bool{
	phrase.ObjectCreationExpression: true,
}

func scanForExpression(document *Document, node *phrase.Phrase) hasTypes {
	var phraseToExpressionConstructor = map[phrase.PhraseType]expressionConstructorForPhrase{
		phrase.FunctionCallExpression:         newFunctionCall,
		phrase.ConstantAccessExpression:       newConstantAccess,
		phrase.ScopedPropertyAccessExpression: newScopedPropertyAccess,
		phrase.ScopedCallExpression:           newScopedMethodAccess,
		phrase.ClassConstantAccessExpression:  newScopedConstantAccess,
		phrase.ClassTypeDesignator:            newClassTypeDesignator,
		phrase.ObjectCreationExpression:       newClassTypeDesignator,
		phrase.SimpleVariable:                 newVariableExpression,
		phrase.PropertyAccessExpression:       newPropertyAccess,
		phrase.MethodCallExpression:           newMethodAccess,
	}
	var expression hasTypes = nil
	defer func() {
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
		expression = constructor(document, node)
	}
	return expression
}

func (s *Expression) Serialise(serialiser *Serialiser) {
	s.Type.Write(serialiser)
	switch expression := s.Scope.(type) {
	case *Variable:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(variableKind))
		expression.Serialise(serialiser)
	case *ClassAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(classAccessKind))
		expression.Serialise(serialiser)
	case *ClassTypeDesignator:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(classTypeDesignatorKind))
		expression.Serialise(serialiser)
	case *ConstantAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(constantAccessKind))
		expression.Serialise(serialiser)
	case *FunctionCall:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(functionCallKind))
		expression.Serialise(serialiser)
	case *PropertyAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(propertyAccessKind))
		expression.Serialise(serialiser)
	case *ScopedConstantAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(scopedConstantAccessKind))
		expression.Serialise(serialiser)
	case *ScopedMethodAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(scopedMethodAccessKind))
		expression.Serialise(serialiser)
	case *ScopedPropertyAccess:
		serialiser.WriteBool(true)
		serialiser.WriteInt(int(scopedPropertyAccessKind))
		expression.Serialise(serialiser)
	default:
		serialiser.WriteBool(false)
	}
	serialiser.WriteLocation(s.Location)
	serialiser.WriteString(s.Name)
}

func ReadExpression(serialiser *Serialiser) Expression {
	expr := Expression{
		Type: ReadTypeComposite(serialiser),
	}
	if serialiser.ReadBool() {
		switch expressionKind(serialiser.ReadInt()) {
		case variableKind:
			expr.Scope = ReadVariable(serialiser)
		case classAccessKind:
			expr.Scope = ReadClassAccess(serialiser)
		case classTypeDesignatorKind:
			expr.Scope = ReadClassTypeDesignator(serialiser)
		case constantAccessKind:
			expr.Scope = ReadConstantAccess(serialiser)
		case functionCallKind:
			expr.Scope = ReadFunctionCall(serialiser)
		case propertyAccessKind:
			expr.Scope = ReadPropertyAccess(serialiser)
		case scopedConstantAccessKind:
			expr.Scope = ReadScopedConstantAccess(serialiser)
		case scopedMethodAccessKind:
			expr.Scope = ReadScopedMethodAccess(serialiser)
		case scopedPropertyAccessKind:
			expr.Scope = ReadScopedPropertyAccess(serialiser)
		}
	}
	expr.Location = serialiser.ReadLocation()
	expr.Name = serialiser.ReadString()
	return expr
}
