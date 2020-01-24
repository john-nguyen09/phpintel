package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/cmd"
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
	document.Load()
	symbol := document.HasTypesAtPos(params.TextDocumentPositionParams.Position)
	var hover *protocol.Hover = nil
	// log.Printf("Hover: %T\n", symbol)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		for _, typeString := range v.Type.Resolve() {
			classes := store.GetClasses(typeString.GetFQN())
			if len(classes) > 0 {
				firstClass := classes[0]
				constructor := firstClass.GetConstructor(store)
				if constructor != nil {
					hover = cmd.MethodToHover(v, *constructor)
				} else {
					hover = cmd.ClassToHover(v, *firstClass)
				}
				break
			}
		}
	case *analysis.ClassAccess:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
			interfaces := store.GetInterfaces(typeString.GetFQN())
			if len(interfaces) > 0 {
				hover = cmd.InterfaceToHover(symbol, *interfaces[0])
				break
			}
		}
	case *analysis.InterfaceAccess:
		interfaces := []*analysis.Interface{}
		for _, typeString := range v.Type.Resolve() {
			interfaces = append(interfaces, store.GetInterfaces(typeString.GetFQN())...)
			if len(interfaces) > 0 {
				hover = cmd.InterfaceToHover(symbol, *interfaces[0])
				break
			}
		}
	case *analysis.TraitAccess:
		for _, typeString := range v.GetTypes().Resolve() {
			traits := store.GetTraits(typeString.GetFQN())
			if len(traits) > 0 {
				hover = cmd.TraitToHover(v, *traits[0])
				break
			}
		}
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		name := analysis.NewTypeString(v.Name)
		consts = append(consts, store.GetConsts(document.GetImportTable().GetConstReferenceFQN(store, name))...)
		if len(consts) > 0 {
			hover = cmd.ConstToHover(symbol, *consts[0])
			break
		}
		defines = append(defines, store.GetDefines(document.GetImportTable().GetConstReferenceFQN(store, name))...)
		if len(defines) > 0 {
			hover = cmd.DefineToHover(symbol, *defines[0])
			break
		}
	case *analysis.FunctionCall:
		functions := []*analysis.Function{}
		name := analysis.NewTypeString(v.Name)
		functions = append(functions, store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(store, name))...)
		if len(functions) > 0 {
			hover = cmd.FunctionToHover(symbol, *functions[0])
			break
		}
	case *analysis.ScopedConstantAccess:
		classConsts := []*analysis.ClassConst{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			classConsts = append(classConsts, store.GetClassConsts(
				scopeType.GetFQN(), v.Name)...)
			if len(classConsts) > 0 {
				hover = cmd.ClassConstToHover(symbol, *classConsts[0])
				break
			}
		}
	case *analysis.ScopedMethodAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope().GetFQN()
		}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			methods := []*analysis.Method{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.StaticMethodsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
			if len(methods) > 0 {
				hover = cmd.MethodToHover(symbol, *methods[0])
				break
			}
		}
	case *analysis.ScopedPropertyAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope().GetFQN()
		}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			properties := []*analysis.Property{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				properties = append(properties, analysis.GetClassProperties(store, class, v.Name,
					analysis.StaticPropsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
			if len(properties) > 0 {
				hover = cmd.PropertyToHover(symbol, *properties[0])
				break
			}
		}
	case *analysis.Variable:
		v.Resolve(store)
		hover = cmd.VariableToHover(v)
	case *analysis.PropertyAccess:
		properties := []*analysis.Property{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				properties = append(properties, analysis.GetClassProperties(store, class, "$"+v.Name,
					analysis.NewSearchOptions())...)
			}
			if len(properties) > 0 {
				hover = cmd.PropertyToHover(symbol, *properties[0])
				break
			}
		}
	case *analysis.MethodAccess:
		methods := []*analysis.Method{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.NewSearchOptions())...)
			}
			if len(methods) > 0 {
				hover = cmd.MethodToHover(symbol, *methods[0])
				break
			}
		}
	case *analysis.TypeDeclaration:
		for _, typeString := range v.Type.Resolve() {
			classes := store.GetClasses(typeString.GetFQN())
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
			interfaces := store.GetInterfaces(typeString.GetFQN())
			if len(interfaces) > 0 {
				hover = cmd.InterfaceToHover(symbol, *interfaces[0])
				break
			}
		}
	}
	if hover == nil && symbol != nil {
		symbolRange := symbol.GetLocation().Range
		hover = &protocol.Hover{
			Contents: protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: "",
			},
			Range: &symbolRange,
		}
	}
	return hover, nil
}
