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
		return nil, nil
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, nil
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
		currentClass := document.GetClassScopeAtSymbol(v)
		var classConsts []analysis.ClassConstWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ccs := analysis.EmptyInheritedClassConst()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ccs.Merge(q.GetClassClassConsts(class, v.Name, ccs.SearchedFQNs))
			}
			for _, intf := range q.GetInterfaces(scopeType.GetFQN()) {
				ccs.Merge(q.GetInterfaceClassConsts(intf, v.Name, ccs.SearchedFQNs))
			}
			classConsts = analysis.MergeClassConstWithScope(classConsts, ccs.ReduceStatic(currentClass, v))
		}
		if len(classConsts) > 0 {
			hover = classConstsToHover(symbol, classConsts)
		}
	case *analysis.ScopedMethodAccess:
		currentClass := document.GetClassScopeAtSymbol(v)
		var methods []analysis.MethodWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ms := analysis.EmptyInheritedMethods()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ms.Merge(q.GetClassMethods(class, v.Name, ms.SearchedFQNs))
			}
			methods = analysis.MergeMethodWithScope(methods, ms.ReduceStatic(currentClass, v))
		}
		if len(methods) > 0 {
			hover = methodsToHover(symbol, methods)
		}
	case *analysis.ScopedPropertyAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var props []analysis.PropWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ps := analysis.EmptyInheritedProps()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ps.Merge(q.GetClassProps(class, v.Name, ps.SearchedFQNs))
			}
			props = analysis.MergePropWithScope(props, ps.ReduceStatic(currentClass, v))
		}
		if len(props) > 0 {
			hover = propertiesToHover(symbol, props)
		}
	case *analysis.Variable:
		v.Resolve(resolveCtx)
		hover = variableToHover(v)
	case *analysis.PropertyAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var props []analysis.PropWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ps := analysis.EmptyInheritedProps()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ps.Merge(q.GetClassProps(class, "$"+v.Name, ps.SearchedFQNs))
			}
			props = analysis.MergePropWithScope(props, ps.ReduceAccess(currentClass, v))
		}
		if len(props) > 0 {
			hover = propertiesToHover(symbol, props)
		}
	case *analysis.MethodAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var methods []analysis.MethodWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ms := analysis.EmptyInheritedMethods()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ms.Merge(q.GetClassMethods(class, v.Name, ms.SearchedFQNs))
			}
			for _, theInterface := range q.GetInterfaces(scopeType.GetFQN()) {
				ms.Merge(q.GetInterfaceMethods(theInterface, v.Name, ms.SearchedFQNs))
			}
			for _, trait := range q.GetTraits(scopeType.GetFQN()) {
				ms.Merge(q.GetTraitMethods(trait, v.Name))
			}
			methods = analysis.MergeMethodWithScope(methods, ms.ReduceAccess(currentClass, v))
		}
		if len(methods) > 0 {
			hover = methodsToHover(symbol, methods)
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
