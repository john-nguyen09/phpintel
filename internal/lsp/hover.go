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
	symbol := document.SymbolAtPos(params.TextDocumentPositionParams.Position)
	var hover *protocol.Hover = nil
	// log.Printf("Hover: %T\n", symbol)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
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
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		for _, typeString := range v.Type.Resolve() {
			consts = append(consts, store.GetConsts(typeString.GetFQN())...)
			if len(consts) > 0 {
				hover = cmd.ConstToHover(symbol, *consts[0])
				break
			}
			defines = append(defines, store.GetDefines(typeString.GetFQN())...)
			if len(defines) > 0 {
				hover = cmd.DefineToHover(symbol, *defines[0])
				break
			}
		}
	case *analysis.FunctionCall:
		functions := []*analysis.Function{}
		name := analysis.NewTypeString(v.Name)
		functions = append(functions, store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(name))...)
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
		methods := []*analysis.Method{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, method := range store.GetMethods(scopeType.GetFQN(), v.Name) {
				if !method.IsStatic {
					continue
				}
				methods = append(methods, method)
			}
			if len(methods) > 0 {
				hover = cmd.MethodToHover(symbol, *methods[0])
				break
			}
		}
	case *analysis.ScopedPropertyAccess:
		properties := []*analysis.Property{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, property := range store.GetProperties(scopeType.GetFQN(), v.Name) {
				if !property.IsStatic {
				}
				properties = append(properties, property)
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
			for _, property := range store.GetProperties(scopeType.GetFQN(), "$"+v.Name) {
				properties = append(properties, property)
			}
			if len(properties) > 0 {
				hover = cmd.PropertyToHover(symbol, *properties[0])
				break
			}
		}
	case *analysis.MethodAccess:
		methods := []*analysis.Method{}
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, method := range store.GetMethods(scopeType.GetFQN(), v.Name) {
				methods = append(methods, method)
			}
			if len(methods) > 0 {
				hover = cmd.MethodToHover(symbol, *methods[0])
				break
			}
		}
	}
	return hover, nil
}
