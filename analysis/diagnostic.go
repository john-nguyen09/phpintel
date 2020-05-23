package analysis

import (
	"time"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

const source = "phpintel"

// GetParserDiagnostics returns the diagnostics for the syntax error
func GetParserDiagnostics(document *Document) []protocol.Diagnostic {
	rootNode := document.GetRootNode()
	diagnostics := []protocol.Diagnostic{}
	traverser := util.NewTraverser(rootNode)
	traverser.Traverse(func(node phrase.AstNode, _ []*phrase.Phrase) util.VisitorContext {
		if p, ok := node.(*phrase.Phrase); ok && p.Type == phrase.DocumentComment {
			return util.VisitorContext{ShouldAscend: false}
		}
		if err, ok := node.(*phrase.ParseError); ok {
			diagnostics = append(diagnostics, parserErrorToDiagnostic(document, err))
		}
		return util.VisitorContext{ShouldAscend: true}
	})
	return diagnostics
}

func parserErrorToDiagnostic(document *Document, err *phrase.ParseError) protocol.Diagnostic {
	message := "Unexpected " + err.Type.String() + "."
	if err.Expected != lexer.Undefined {
		message += " Expected " + err.Expected.String() + "."
	}

	return protocol.Diagnostic{
		Range:    document.errorRange(err),
		Message:  message,
		Severity: protocol.SeverityError,
		Source:   source,
	}
}

// UnusedDiagnostics returns the diagnostics for unused variables or imports
// TODO: provide unused imports
func UnusedDiagnostics(document *Document) []protocol.Diagnostic {
	diagnostics := []protocol.Diagnostic{}
	unusedVariables := document.UnusedVariables()
	for _, unusedVariable := range unusedVariables {
		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range:    unusedVariable.GetLocation().Range,
			Message:  unusedVariable.Name + " is declared but its value is never read.",
			Source:   source,
			Severity: protocol.SeverityHint,
			Tags:     []protocol.DiagnosticTag{protocol.Unnecessary},
		})
	}
	for _, importTable := range document.importTables {
		for _, unusedImport := range importTable.unusedImportItems() {
			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range:    unusedImport.locationRange,
				Message:  unusedImport.name + " is declared but its value is never used.",
				Source:   source,
				Severity: protocol.SeverityHint,
				Tags:     []protocol.DiagnosticTag{protocol.Unnecessary},
			})
		}
	}
	return diagnostics
}

// DeprecatedDiagnostics returns the diagnostics for deprecated references
func DeprecatedDiagnostics(ctx ResolveContext) []protocol.Diagnostic {
	defer util.TimeTrack(time.Now(), "DeprecatedDiagnostics")
	doc := ctx.document
	store := ctx.store
	create := func(r protocol.Range, message string) protocol.Diagnostic {
		return protocol.Diagnostic{
			Range:    r,
			Message:  message,
			Source:   source,
			Severity: protocol.SeverityHint,
			Tags:     []protocol.DiagnosticTag{protocol.Deprecated},
		}
	}
	diagnostics := []protocol.Diagnostic{}
	TraverseDocument(doc, func(s Symbol) {
		switch v := s.(type) {
		case *FunctionCall:
			t := NewTypeString(v.Name)
			fqn := doc.ImportTableAtPos(v.Location.Range.Start).GetFunctionReferenceFQN(store, t)
			for _, f := range store.GetFunctions(fqn) {
				if f.deprecatedTag != nil {
					diagnostics = append(diagnostics, create(
						v.Location.Range,
						deprecatedDescription(v.Name+" is deprecated", f.deprecatedTag),
					))
					break
				}
			}
		case *ClassTypeDesignator:
			v.Resolve(ctx)
		LClassTypeDesignator:
			for _, t := range v.GetTypes().Resolve() {
				for _, c := range store.GetClasses(t.GetFQN()) {
					if c.deprecatedTag != nil {
						diagnostics = append(diagnostics, create(
							v.Location.Range,
							deprecatedDescription(v.Name+" is deprecated", c.deprecatedTag),
						))
						break LClassTypeDesignator
					}
				}
			}
		case *ClassAccess:
			v.Resolve(ctx)
		LClass:
			for _, t := range v.GetTypes().Resolve() {
				for _, c := range ctx.store.GetClasses(t.GetFQN()) {
					if c.deprecatedTag != nil {
						diagnostics = append(diagnostics, create(
							v.Location.Range,
							deprecatedDescription(v.Name+" is deprecated", c.deprecatedTag),
						))
						break LClass
					}
				}
				for _, i := range ctx.store.GetInterfaces(t.GetFQN()) {
					if i.deprecatedTag != nil {
						diagnostics = append(diagnostics, create(
							v.Location.Range,
							deprecatedDescription(v.Name+" is deprecated", i.deprecatedTag),
						))
						break LClass
					}
				}
			}
		case *InterfaceAccess:
			v.Resolve(ctx)
		LInterface:
			for _, t := range v.GetTypes().Resolve() {
				for _, i := range ctx.store.GetInterfaces(t.GetFQN()) {
					if i.deprecatedTag != nil {
						diagnostics = append(diagnostics, create(
							v.Location.Range,
							deprecatedDescription(v.Name+" is deprecated", i.deprecatedTag),
						))
						break LInterface
					}
				}
			}
		case *ConstantAccess:
			v.Resolve(ctx)
			name := NewTypeString(v.Name)
			fqn := doc.ImportTableAtPos(v.Location.Range.Start).GetConstReferenceFQN(store, name)
			var shouldStop bool
			for _, c := range store.GetConsts(fqn) {
				if c.deprecatedTag != nil {
					diagnostics = append(diagnostics, create(
						v.Location.Range,
						deprecatedDescription(v.Name+" is deprecated", c.deprecatedTag),
					))
					shouldStop = true
					break
				}
			}
			if shouldStop {
				break
			}
			for _, c := range store.GetDefines(fqn) {
				if c.deprecatedTag != nil {
					diagnostics = append(diagnostics, create(
						v.Location.Range,
						deprecatedDescription(v.Name+" is deprecated", c.deprecatedTag),
					))
					break
				}
			}
		case *ScopedConstantAccess:
		LScopedConstant:
			for _, scopeType := range v.ResolveAndGetScope(ctx).Resolve() {
				for _, c := range store.GetClassConsts(scopeType.GetFQN(), v.Name) {
					if c.deprecatedTag != nil {
						diagnostics = append(diagnostics, create(
							v.Location.Range,
							deprecatedDescription(c.ReferenceFQN()+" is deprecated", c.deprecatedTag),
						))
						break LScopedConstant
					}
				}
			}
		case *ScopedMethodAccess:
			name := ""
			classScope := ""
			if hasName, ok := v.Scope.(HasName); ok {
				name = hasName.GetName()
			}
			if hasScope, ok := v.Scope.(HasScope); ok {
				classScope = hasScope.GetScope()
			}
		LScopedMethod:
			for _, scopeType := range v.ResolveAndGetScope(ctx).Resolve() {
				for _, class := range store.GetClasses(scopeType.GetFQN()) {
					for _, m := range GetClassMethods(store, class, v.Name,
						StaticMethodsScopeAware(NewSearchOptions(), classScope, name)) {
						if m.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(m.ReferenceFQN()+" is deprecated", m.deprecatedTag),
							))
							break LScopedMethod
						}
					}
				}
			}
		case *ScopedPropertyAccess:
			name := ""
			classScope := ""
			if hasName, ok := v.Scope.(HasName); ok {
				name = hasName.GetName()
			}
			if hasScope, ok := v.Scope.(HasScope); ok {
				classScope = hasScope.GetScope()
			}
		LScopedProp:
			for _, scopeType := range v.ResolveAndGetScope(ctx).Resolve() {
				for _, class := range store.GetClasses(scopeType.GetFQN()) {
					for _, p := range GetClassProperties(store, class, v.Name,
						StaticPropsScopeAware(NewSearchOptions(), classScope, name)) {
						if p.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(p.ReferenceFQN()+" is deprecated", p.deprecatedTag),
							))
							break LScopedProp
						}
					}
				}
			}
		case *PropertyAccess:
		LProp:
			for _, scopeType := range v.ResolveAndGetScope(ctx).Resolve() {
				for _, class := range store.GetClasses(scopeType.GetFQN()) {
					for _, p := range GetClassProperties(store, class, "$"+v.Name,
						PropsScopeAware(NewSearchOptions(), doc, v.Scope)) {
						if p.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(p.ReferenceFQN()+" is deprecated", p.deprecatedTag),
							))
							break LProp
						}
					}
				}
			}
		case *MethodAccess:
		LMethod:
			for _, scopeType := range v.ResolveAndGetScope(ctx).Resolve() {
				for _, class := range store.GetClasses(scopeType.GetFQN()) {
					for _, m := range GetClassMethods(store, class, v.Name,
						MethodsScopeAware(NewSearchOptions(), doc, v.Scope)) {
						if m.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(m.ReferenceFQN()+" is deprecated", m.deprecatedTag),
							))
							break LMethod
						}
					}
				}
				for _, theInterface := range store.GetInterfaces(scopeType.GetFQN()) {
					for _, m := range GetInterfaceMethods(store, theInterface, v.Name,
						MethodsScopeAware(NewSearchOptions(), doc, v.Scope)) {
						if m.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(m.ReferenceFQN()+" is deprecated", m.deprecatedTag),
							))
							break LMethod
						}
					}
				}
				for _, trait := range store.GetTraits(scopeType.GetFQN()) {
					for _, m := range GetTraitMethods(store, trait, v.Name,
						MethodsScopeAware(NewSearchOptions(), doc, v.Scope)) {
						if m.deprecatedTag != nil {
							diagnostics = append(diagnostics, create(
								v.Location.Range,
								deprecatedDescription(m.ReferenceFQN()+" is deprecated", m.deprecatedTag),
							))
							break LMethod
						}
					}
				}
			}
		}
	}, nil)
	return diagnostics
}

func deprecatedDescription(msg string, t *tag) string {
	description := ""
	if t.Name != "" {
		description += "Since " + t.Name
		if t.Description != "" {
			description += " - "
		}
	}
	if t.Description != "" {
		description += t.Description
	}
	if description != "" {
		return msg + ": " + description
	}
	return msg
}
