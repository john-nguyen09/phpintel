package lsp

import (
	"context"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	q := analysis.NewQuery(store)
	resolveCtx := analysis.NewResolveContext(q, document)
	document.Load()
	pos := params.TextDocumentPositionParams.Position
	symbol := document.HasTypesAtPos(pos)
	var hover *protocol.Hover = nil
	// log.Printf("Hover: %T\n", symbol)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		classes := []*analysis.Class{}
		for _, typeString := range v.GetTypes().Resolve() {
			classes = append(classes, q.GetClasses(typeString.GetFQN())...)
		}
		constructors := []analysis.MethodWithScope{}
		for _, class := range classes {
			constructor := q.GetClassConstructor(class)
			if constructor.Method != nil {
				constructors = append(constructors, constructor)
			}
		}
		if len(constructors) > 0 {
			hover = methodsToHover(v, constructors)
		} else if len(classes) > 0 {
			hover = classesToHover(v, classes)
		}
	case *analysis.ClassAccess:
		classes := []*analysis.Class{}
		interfaces := []*analysis.Interface{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, q.GetClasses(typeString.GetFQN())...)
			interfaces = append(interfaces, q.GetInterfaces(typeString.GetFQN())...)
		}
		var sb strings.Builder
		if len(classes) > 0 {
			sb.WriteString(formatClasses(classes).String())
		}
		if len(interfaces) > 0 {
			sb.WriteString(formatInterfaces(interfaces).String())
		}
		hover = hoverFromSymbol(v)
		hover.Contents = protocol.MarkupContent{
			Kind:  "markdown",
			Value: sb.String(),
		}
	case *analysis.InterfaceAccess:
		interfaces := []*analysis.Interface{}
		for _, typeString := range v.Type.Resolve() {
			interfaces = append(interfaces, q.GetInterfaces(typeString.GetFQN())...)
		}
		if len(interfaces) > 0 {
			hover = interfacesToHover(v, interfaces)
		}
	case *analysis.TraitAccess:
		traits := []*analysis.Trait{}
		for _, typeString := range v.GetTypes().Resolve() {
			traits = append(traits, q.GetTraits(typeString.GetFQN())...)
		}
		if len(traits) > 0 {
			hover = traitsToHover(v, traits)
		}
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		name := analysis.NewTypeString(v.Name)
		consts = append(consts, q.GetConsts(document.ImportTableAtPos(pos).GetConstReferenceFQN(q, name))...)
		var sb strings.Builder
		if len(consts) > 0 {
			sb.WriteString(formatConsts(consts).String())
		}
		defines = append(defines, q.GetDefines(document.ImportTableAtPos(pos).GetConstReferenceFQN(q, name))...)
		if len(defines) > 0 {
			sb.WriteString(formatDefines(defines).String())
		}
		hover = hoverFromSymbol(v)
		hover.Contents = protocol.MarkupContent{
			Kind:  "markdown",
			Value: sb.String(),
		}
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		functions := q.GetFunctions(document.ImportTableAtPos(pos).GetFunctionReferenceFQN(q, name))
		if len(functions) > 0 {
			hover = functionsToHover(symbol, functions)
			break
		}
	case *analysis.ScopedConstantAccess:
		classConsts := []*analysis.ClassConst{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			classConsts = append(classConsts, q.GetClassConsts(scopeType.GetFQN(), v.Name)...)
		}
		if len(classConsts) > 0 {
			hover = classConstsToHover(symbol, classConsts)
		}
	case *analysis.ScopedMethodAccess:
		var scopeName string
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			scopeName = hasName.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		methods := analysis.EmptyInheritedMethods()
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				methods.Merge(q.GetClassMethods(class, v.Name, methods.SearchedFQNs))
			}
		}
		if methods.Len() > 0 {
			hover = methodsToHover(symbol, methods.ReduceStatic(currentClass, scopeName))
		}
	case *analysis.ScopedPropertyAccess:
		var scopeName string
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			scopeName = hasName.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		properties := analysis.EmptyInheritedProps()
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				properties.Merge(q.GetClassProps(class, v.Name, properties.SearchedFQNs))
			}
		}
		if properties.Len() > 0 {
			hover = propertiesToHover(symbol, properties.ReduceStatic(currentClass, scopeName))
		}
	case *analysis.Variable:
		v.Resolve(resolveCtx)
		hover = variableToHover(v)
	case *analysis.PropertyAccess:
		var scopeName string
		if n, ok := v.Scope.(analysis.HasName); ok {
			scopeName = n.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		types := v.ResolveAndGetScope(resolveCtx)
		properties := analysis.EmptyInheritedProps()
		for _, scopeType := range types.Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				properties.Merge(q.GetClassProps(class, "$"+v.Name, properties.SearchedFQNs))
			}
		}
		if properties.Len() > 0 {
			hover = propertiesToHover(symbol, properties.ReduceAccess(currentClass, scopeName, types))
		}
	case *analysis.MethodAccess:
		var scopeName string
		if n, ok := v.Scope.(analysis.HasName); ok {
			scopeName = n.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		types := v.ResolveAndGetScope(resolveCtx)
		methods := analysis.EmptyInheritedMethods()
		for _, scopeType := range types.Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				methods.Merge(q.GetClassMethods(class, v.Name, methods.SearchedFQNs))
			}
			for _, theInterface := range q.GetInterfaces(scopeType.GetFQN()) {
				methods.Merge(q.GetInterfaceMethods(theInterface, v.Name, methods.SearchedFQNs))
			}
			for _, trait := range q.GetTraits(scopeType.GetFQN()) {
				methods.Merge(q.GetTraitMethods(trait, v.Name))
			}
		}
		if methods.Len() > 0 {
			hover = methodsToHover(symbol, methods.ReduceAccess(currentClass, scopeName, types))
		}
	case *analysis.TypeDeclaration:
		classes := []*analysis.Class{}
		interfaces := []*analysis.Interface{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, q.GetClasses(typeString.GetFQN())...)
			interfaces = append(interfaces, q.GetInterfaces(typeString.GetFQN())...)
		}
		var sb strings.Builder
		if len(classes) > 0 {
			sb.WriteString(formatClasses(classes).String())
		}
		if len(interfaces) > 0 {
			sb.WriteString(formatInterfaces(interfaces).String())
		}
		hover = hoverFromSymbol(v)
		hover.Contents = protocol.MarkupContent{
			Kind:  "markdown",
			Value: sb.String(),
		}
	}
	if hover == nil && symbol != nil {
		hover = hoverFromSymbol(symbol)
	}
	return hover, nil
}
