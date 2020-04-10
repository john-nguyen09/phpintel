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
	resolveCtx := analysis.NewResolveContext(store, document)
	document.Load()
	pos := params.TextDocumentPositionParams.Position
	symbol := document.HasTypesAtPos(pos)
	var hover *protocol.Hover = nil
	// log.Printf("Hover: %T\n", symbol)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
		}
		constructors := []*analysis.Method{}
		for _, class := range classes {
			constructor := class.GetConstructor(store)
			if constructor != nil {
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
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			interfaces = append(interfaces, store.GetInterfaces(typeString.GetFQN())...)
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
			interfaces = append(interfaces, store.GetInterfaces(typeString.GetFQN())...)
		}
		if len(interfaces) > 0 {
			hover = interfacesToHover(v, interfaces)
		}
	case *analysis.TraitAccess:
		traits := []*analysis.Trait{}
		for _, typeString := range v.GetTypes().Resolve() {
			traits = append(traits, store.GetTraits(typeString.GetFQN())...)
		}
		if len(traits) > 0 {
			hover = traitsToHover(v, traits)
		}
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		name := analysis.NewTypeString(v.Name)
		consts = append(consts, store.GetConsts(document.ImportTableAtPos(pos).GetConstReferenceFQN(store, name))...)
		var sb strings.Builder
		if len(consts) > 0 {
			sb.WriteString(formatConsts(consts).String())
		}
		defines = append(defines, store.GetDefines(document.ImportTableAtPos(pos).GetConstReferenceFQN(store, name))...)
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
		functions := store.GetFunctions(document.ImportTableAtPos(pos).GetFunctionReferenceFQN(store, name))
		if len(functions) > 0 {
			hover = functionsToHover(symbol, functions)
			break
		}
	case *analysis.ScopedConstantAccess:
		classConsts := []*analysis.ClassConst{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			classConsts = append(classConsts, store.GetClassConsts(scopeType.GetFQN(), v.Name)...)
		}
		if len(classConsts) > 0 {
			hover = classConstsToHover(symbol, classConsts)
		}
	case *analysis.ScopedMethodAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope()
		}
		methods := []*analysis.Method{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.StaticMethodsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
		}
		if len(methods) > 0 {
			hover = methodsToHover(symbol, methods)
		}
	case *analysis.ScopedPropertyAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope()
		}
		properties := []*analysis.Property{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				properties = append(properties, analysis.GetClassProperties(store, class, v.Name,
					analysis.StaticPropsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
		}
		if len(properties) > 0 {
			hover = propertiesToHover(symbol, properties)
		}
	case *analysis.Variable:
		v.Resolve(resolveCtx)
		hover = variableToHover(v)
	case *analysis.PropertyAccess:
		properties := []*analysis.Property{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				properties = append(properties, analysis.GetClassProperties(store, class, "$"+v.Name,
					analysis.PropsScopeAware(analysis.NewSearchOptions(), document, v.Scope))...)
			}
		}
		if len(properties) > 0 {
			hover = propertiesToHover(symbol, properties)
		}
	case *analysis.MethodAccess:
		methods := []*analysis.Method{}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, v.Scope))...)
			}
			for _, theInterface := range store.GetInterfaces(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetInterfaceMethods(store, theInterface, v.Name,
					analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, v.Scope))...)
			}
			for _, trait := range store.GetTraits(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetTraitMethods(store, trait, v.Name,
					analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, v.Scope))...)
			}
		}
		if len(methods) > 0 {
			hover = methodsToHover(symbol, methods)
		}
	case *analysis.TypeDeclaration:
		classes := []*analysis.Class{}
		interfaces := []*analysis.Interface{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			interfaces = append(interfaces, store.GetInterfaces(typeString.GetFQN())...)
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
