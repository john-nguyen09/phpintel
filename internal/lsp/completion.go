package lsp

import (
	"context"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type completionContext struct {
	doc   *analysis.Document
	store *analysis.Store
	pos   protocol.Position
}

func (s *Server) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	defer util.TimeTrack(time.Now(), "completion")
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	document.Lock()
	defer document.Unlock()
	document.Load()
	var completionList *protocol.CompletionList = nil
	pos := params.Position
	resolveCtx := analysis.NewResolveContext(store, document)
	completionCtx := &completionContext{document, store, pos}
	symbol := document.HasTypesAtPos(pos)
	word := document.WordAtPos(pos)
	nodes := document.NodeSpineAt(document.OffsetAtPosition(pos))
	// log.Printf("Completion: %s %v %T %s, kind: %v", word, pos, symbol, nodes, params.Context.TriggerKind)
	parent := nodes.Parent()
	if parent != nil {
		switch parent.Type() {
		case "::":
			prev := parent.PrevSibling()
			if prev != nil {
				s := document.HasTypesAtPos(util.PointToPosition(prev.EndPoint()))
				// log.Printf("%T %v %s %v", s, s, prev.Type(), util.PointToPosition(prev.EndPoint()))
				if s != nil {
					s.Resolve(resolveCtx)
					completionList = scopedAccessCompletion(completionCtx, word, s)
				}
			}
		case "->":
			prev := parent.PrevSibling()
			if prev != nil {
				s := document.HasTypesAtPos(util.PointToPosition(prev.EndPoint()))
				// log.Printf("%T %s %v", s, prev.Type(), util.PointToPosition(prev.EndPoint()))
				if s != nil {
					s.Resolve(resolveCtx)
					completionList = memberAccessCompletion(completionCtx, word, s)
				}
			}
		case "name":
			par := nodes.Parent()
			prev := parent.PrevSibling()
			if (prev != nil && prev.Type() == "use") || (par != nil && par.Type() == "namespace_name") {
				completionList = useCompletion(completionCtx, word)
				break
			}
			if par != nil {
				switch par.Type() {
				case "class_constant_access_expression":
					if s, ok := symbol.(*analysis.ScopedConstantAccess); ok {
						if s.Scope != nil {
							s.Scope.Resolve(resolveCtx)
							completionList = scopedAccessCompletion(completionCtx, word, s.Scope)
						}
					}
				case "scoped_call_expression":
					if s, ok := symbol.(*analysis.ScopedMethodAccess); ok {
						if s.Scope != nil {
							s.Scope.Resolve(resolveCtx)
							completionList = scopedAccessCompletion(completionCtx, word, s.Scope)
						}
					}
				case "member_access_expression":
					if s, ok := symbol.(*analysis.PropertyAccess); ok {
						if s.Scope != nil {
							s.Scope.Resolve(resolveCtx)
							completionList = memberAccessCompletion(completionCtx, word, s.Scope)
						}
					}
				case "member_call_expression":
					if s, ok := symbol.(*analysis.MethodAccess); ok {
						if s.Scope != nil {
							s.Scope.Resolve(resolveCtx)
							completionList = memberAccessCompletion(completionCtx, word, s.Scope)
						}
					}
				case "named_label_statement":
					completionList = nameCompletion(completionCtx, symbol, word)
				case "qualified_name":
					par = nodes.Parent()
					if par != nil && par.Type() == "namespace_use_clause" {
						completionList = useCompletion(completionCtx, word)
						break
					}
					completionList = nameCompletion(completionCtx, symbol, word)
				case "ERROR":
					parPrev := par.PrevSibling()
					par = nodes.Parent()
					if parPrev != nil {
						t := parPrev.Type()
						if t == "php_tag" {
							completionList = phpTagCompletion(word)
							break
						}
						if par != nil && par.Type() == "declaration_list" {
							completionList = keywordCompletion(completionCtx, word)
							break
						}
					}
					prev := parent.PrevSibling()
					if par.Type() == "formal_parameters" || (prev != nil && prev.Type() == "(") {
						completionList = typeCompletion(completionCtx, word)
						break
					}
					if par.Type() == "member_access_expression" {
						if parPrev != nil && parPrev.Type() == "->" {
							prev = parPrev.PrevSibling()
							if prev != nil {
								s := document.HasTypesAtPos(util.PointToPosition(prev.EndPoint()))
								if s != nil {
									s.Resolve(resolveCtx)
									completionList = memberAccessCompletion(completionCtx, word, s)
								}
							}
						}
						break
					}
					if par != nil && par.Type() == "dynamic_variable_name" {
						completionList = variableCompletion(completionCtx, resolveCtx, word)
						break
					}
					completionList = nameCompletion(completionCtx, symbol, word)
				case "variable_name":
					completionList = variableCompletion(completionCtx, resolveCtx, word)
				}
			}
		case "type":
			completionList = nameCompletion(completionCtx, symbol, word)
		case "$":
			completionList = variableCompletion(completionCtx, resolveCtx, word)
		case "named_label_statement", "\\":
			par := parent.Parent()
			if par != nil && par.Type() == "namespace_name" {
				completionList = useCompletion(completionCtx, word)
				break
			}
			completionList = nameCompletion(completionCtx, symbol, word)
		}
	}
	switch s := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		completionList = classCompletion(completionCtx, s, s.Name)
	case *analysis.TypeDeclaration:
		completionList = typeCompletion(completionCtx, s.Name)
	}
	return completionList, nil
}

func variableCompletion(ctx *completionContext, resolveCtx analysis.ResolveContext, word string) *protocol.CompletionList {
	varTable := ctx.doc.GetVariableTableAt(ctx.pos)
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	if varTable == nil {
		return completionList
	}
	symbol := ctx.doc.HasTypesAtPos(ctx.pos)
	for _, variable := range varTable.GetVariables(ctx.pos) {
		if variable.Name == "$" {
			continue
		}
		if symbol != nil && symbol.GetLocation().Range == variable.GetLocation().Range {
			continue
		}
		if word != "" && !strings.Contains(variable.Name, word) {
			continue
		}

		variable.Resolve(resolveCtx)
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.VariableCompletion,
			Label:         variable.Name,
			Documentation: variable.GetDescription(),
			Detail:        variable.GetDetail(),
		})
	}
	return completionList
}

func nameCompletion(ctx *completionContext, symbol analysis.HasTypes, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	opts := baseSearchOptions()
	classes, searchResult := ctx.store.SearchClasses(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := ctx.doc.ImportTableAtPos(ctx.pos)
	for _, class := range classes {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, class, class.Name, word)
		completionList.Items = append(completionList.Items, classToCompletionItem(class, label, textEdit))
	}
	interfaces, searchResult := ctx.store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, intf := range interfaces {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, intf, intf.Name, word)
		completionList.Items = append(completionList.Items, interfaceToCompletionItem(intf, label, textEdit))
	}
	consts, searchResult := ctx.store.SearchConsts(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, constant := range consts {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, constant, constant.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.ConstantCompletion,
			Label:               label,
			AdditionalTextEdits: textEdits,
			Documentation:       constant.GetDescription(),
			Detail:              getDetailFromTextEdit(constant.Name, textEdit),
		})
	}
	defines, searchResult := ctx.store.SearchDefines(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, define := range defines {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, define, define.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.ConstantCompletion,
			Label:               label,
			InsertText:          label,
			Documentation:       define.GetDescription(),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(define.Name, textEdit),
		})
	}
	functions, searchResult := ctx.store.SearchFunctions(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, function := range functions {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, function, function.Name, word)
		insertText, textFormat, command := HasParamsInsertText(function, label)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			InsertText:          insertText,
			InsertTextFormat:    textFormat,
			Command:             command,
			Kind:                protocol.FunctionCompletion,
			Label:               label,
			AdditionalTextEdits: textEdits,
			Documentation:       function.GetDescription(),
			Detail:              hasParamsDetailWithTextEdit(function, textEdit),
		})
	}
	if analysis.IsFQN(word) {
		namespaces, _ := ctx.store.SearchNamespaces(word, opts)
		for _, ns := range namespaces {
			completionList.Items = append(completionList.Items, namespaceToCompletionItem(ns, word))
		}
	}
	return completionList
}

func classCompletion(ctx *completionContext, symbol analysis.HasTypes, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	classes, searchResult := ctx.store.SearchClasses(word, baseSearchOptions())
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := ctx.doc.ImportTableAtPos(ctx.pos)
	for _, class := range classes {
		name, textEdit := importTable.ResolveToQualified(ctx.doc, class, class.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.ClassCompletion,
			Label:               name,
			Documentation:       class.GetDescription(),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(class.Name, textEdit),
		})
	}
	return completionList
}

func scopedAccessCompletion(ctx *completionContext, word string, scope analysis.HasTypes) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	name := ""
	classScope := ""
	if hasName, ok := scope.(analysis.HasName); ok {
		name = hasName.GetName()
	}
	if hasScope, ok := scope.(analysis.HasScope); ok {
		classScope = hasScope.GetScope()
	}
	for _, scopeType := range scope.GetTypes().Resolve() {
		scopeTypeFQN := scopeType.GetFQN()
		props := []*analysis.Property{}
		methods := []*analysis.Method{}
		for _, class := range ctx.store.GetClasses(scopeTypeFQN) {
			methods = append(methods, analysis.SearchClassMethods(ctx.store, class, word,
				analysis.StaticMethodsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			props = append(props, analysis.SearchClassProperties(ctx.store, class, word,
				analysis.StaticPropsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
		}
		for _, property := range props {
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.PropertyCompletion,
				Label:         property.GetName(),
				Documentation: property.GetDescription(),
			})
		}
		for _, method := range methods {
			insertText, textFormat, command := HasParamsInsertText(method, method.GetName())
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:             protocol.MethodCompletion,
				Label:            method.GetName(),
				InsertText:       insertText,
				InsertTextFormat: textFormat,
				Command:          command,
				Documentation:    method.GetDescription(),
				Detail:           hasParamsDetailWithTextEdit(method, nil),
			})
		}
		classConsts, searchResult := ctx.store.SearchClassConsts(scopeTypeFQN, word, baseSearchOptions())
		completionList.IsIncomplete = !searchResult.IsComplete
		for _, classConst := range classConsts {
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.ConstantCompletion,
				Label:         classConst.GetName(),
				Documentation: classConst.GetDescription(),
			})
		}
	}
	return completionList
}

func memberAccessCompletion(ctx *completionContext, word string, scope analysis.HasTypes) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	for _, scopeType := range scope.GetTypes().Resolve() {
		properties := []*analysis.Property{}
		methods := []*analysis.Method{}
		for _, class := range ctx.store.GetClasses(scopeType.GetFQN()) {
			methods = append(methods, analysis.SearchClassMethods(ctx.store, class, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), ctx.doc, scope))...)
			properties = append(properties, analysis.SearchClassProperties(ctx.store, class, word,
				analysis.PropsScopeAware(analysis.NewSearchOptions(), ctx.doc, scope))...)
		}
		for _, theInterface := range ctx.store.GetInterfaces(scopeType.GetFQN()) {
			methods = append(methods, analysis.SearchInterfaceMethods(ctx.store, theInterface, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), ctx.doc, scope))...)
		}
		for _, trait := range ctx.store.GetTraits(scopeType.GetFQN()) {
			methods = append(methods, analysis.GetTraitMethods(ctx.store, trait, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), ctx.doc, scope))...)
		}
		for _, property := range properties {
			name := property.GetName()
			if !property.IsStatic {
				name = string([]rune(name)[1:])
			}
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.PropertyCompletion,
				Label:         name,
				Documentation: property.GetDescription(),
			})
		}
		for _, method := range methods {
			if method.Name == "__construct" {
				continue
			}
			insertText, textFormat, command := HasParamsInsertText(method, method.GetName())
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:             protocol.MethodCompletion,
				Label:            method.GetName(),
				InsertText:       insertText,
				InsertTextFormat: textFormat,
				Command:          command,
				Documentation:    method.GetDescription(),
				Detail:           hasParamsDetailWithTextEdit(method, nil),
			})
		}
	}
	return completionList
}

func typeCompletion(ctx *completionContext, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	opts := baseSearchOptions()
	classes, searchResult := ctx.store.SearchClasses(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := ctx.doc.ImportTableAtPos(ctx.pos)
	for _, class := range classes {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, class, class.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.ClassCompletion,
			Label:               label,
			Documentation:       class.GetDescription(),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(class.Name, textEdit),
		})
	}
	interfaces, searchResult := ctx.store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, theInterface := range interfaces {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, theInterface, theInterface.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.ClassCompletion,
			Label:               label,
			Documentation:       theInterface.GetDescription(),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(theInterface.Name, textEdit),
		})
	}
	if analysis.IsFQN(word) {
		namespaces, _ := ctx.store.SearchNamespaces(word, opts)
		for _, ns := range namespaces {
			completionList.Items = append(completionList.Items, namespaceToCompletionItem(ns, word))
		}
	}
	return completionList
}

func keywordCompletion(ctx *completionContext, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	return completionList
}

func phpTagCompletion(word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	completion := "php"
	word = strings.ToLower(word)
	for i := 1; i < len(completion); i++ {
		if word == completion[:i] {
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:  protocol.KeywordCompletion,
				Label: completion,
			})
			break
		}
	}
	return completionList
}

func useCompletion(ctx *completionContext, word string) *protocol.CompletionList {
	t := analysis.NewTypeString(word)
	t.SetNamespace("")
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	opts := baseSearchOptions()
	namespaces, _ := ctx.store.SearchNamespaces(t.GetFQN(), opts)
	for _, ns := range namespaces {
		completionList.Items = append(completionList.Items, namespaceToCompletionItem(ns, word))
	}
	classes, searchResult := ctx.store.SearchClasses(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, class := range classes {
		completionList.Items = append(completionList.Items, classToCompletionItem(class, class.Name.GetOriginal(), nil))
	}
	interfaces, searchResult := ctx.store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, intf := range interfaces {
		completionList.Items = append(completionList.Items, interfaceToCompletionItem(intf, intf.Name.GetOriginal(), nil))
	}
	traits, searchResult := ctx.store.SearchTraits(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, trait := range traits {
		completionList.Items = append(completionList.Items, traitToCompletionItem(trait, trait.Name.GetOriginal(), nil))
	}
	return completionList
}
