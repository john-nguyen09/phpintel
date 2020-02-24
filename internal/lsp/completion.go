package lsp

import (
	"context"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

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
	symbol := document.HasTypesAtPos(pos)
	word := document.WordAtPos(pos)
	nodes := document.NodeSpineAt(document.OffsetAtPosition(pos))
	// log.Printf("Completion: %s %v %T %s", word, pos, symbol, nodes)
	parent := nodes.Parent()
	if parent != nil {
		switch parent.Type() {
		case "::":
			prev := parent.PrevSibling()
			if prev != nil {
				s := document.HasTypesAtPos(util.PointToPosition(prev.StartPoint()))
				// log.Printf("%T %v %s %v", s, s, prev.Type(), util.PointToPosition(prev.StartPoint()))
				if s != nil {
					s.Resolve(resolveCtx)
					completionList = scopedAccessCompletion(store, document, word, s)
				}
			}
		case "->":
			prev := parent.PrevSibling()
			if prev != nil {
				s := document.HasTypesAtPos(util.PointToPosition(prev.StartPoint()))
				// log.Printf("%T %v %s %v", s, s, prev.Type(), util.PointToPosition(prev.StartPoint()))
				if s != nil {
					s.Resolve(resolveCtx)
					completionList = memberAccessCompletion(store, document, word, s)
				}
			}
		case "name":
			par := nodes.Parent()
			if par != nil {
				switch par.Type() {
				case "class_constant_access_expression":
					if s, ok := symbol.(*analysis.ScopedConstantAccess); ok {
						s.Scope.Resolve(resolveCtx)
						completionList = scopedAccessCompletion(store, document, word, s.Scope)
					}
				case "scoped_call_expression":
					if s, ok := symbol.(*analysis.ScopedMethodAccess); ok {
						s.Scope.Resolve(resolveCtx)
						completionList = scopedAccessCompletion(store, document, word, s.Scope)
					}
				case "member_access_expression":
					if s, ok := symbol.(*analysis.PropertyAccess); ok {
						s.Scope.Resolve(resolveCtx)
						completionList = memberAccessCompletion(store, document, word, s.Scope)
					}
				case "member_call_expression":
					if s, ok := symbol.(*analysis.MethodAccess); ok {
						s.Scope.Resolve(resolveCtx)
						completionList = memberAccessCompletion(store, document, word, s.Scope)
					}
				case "ERROR":
					completionList = nameCompletion(store, document, symbol, word)
				case "variable_name":
					completionList = variableCompletion(document, pos, word)
				}
			}
		case "$":
			completionList = variableCompletion(document, pos, word)
		case "named_label_statement":
			completionList = nameCompletion(store, document, symbol, word)
		}
	}
	switch s := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		completionList = classCompletion(store, document, s, s.Name, params.Position)
	case *analysis.TypeDeclaration:
		completionList = typeCompletion(store, document, s, s.Name, params.Position)
	}
	return completionList, nil
}

func variableCompletion(document *analysis.Document, pos protocol.Position, word string) *protocol.CompletionList {
	varTable := document.GetVariableTableAt(pos)
	symbol := document.HasTypesAtPos(pos)
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	for _, variable := range varTable.GetVariables() {
		if variable.Name == "$" {
			continue
		}
		if symbol != nil && symbol.GetLocation().Range == variable.GetLocation().Range {
			continue
		}
		if word != "" && !strings.Contains(variable.Name, word) {
			continue
		}

		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.VariableCompletion,
			Label:         variable.Name,
			Documentation: variable.GetDescription(),
			Detail:        variable.GetDetail(),
		})
	}
	return completionList
}

func nameCompletion(store *analysis.Store, document *analysis.Document, symbol analysis.HasTypes, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	classes, searchResult := store.SearchClasses(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := document.GetImportTable()
	for _, class := range classes {
		label, textEdit := importTable.ResolveToQualified(document, class, class.Name, word)
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
	consts, searchResult := store.SearchConsts(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, constant := range consts {
		label, textEdit := importTable.ResolveToQualified(document, constant, constant.Name, word)
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
	defines, searchResult := store.SearchDefines(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, define := range defines {
		label, textEdit := importTable.ResolveToQualified(document, define, define.Name, word)
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
	functions, searchResult := store.SearchFunctions(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, function := range functions {
		label, textEdit := importTable.ResolveToQualified(document, function, function.Name, word)
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
			Detail:              HasParamsDetailWithTextEdit(function, textEdit),
		})
	}
	return completionList
}

func classCompletion(store *analysis.Store, document *analysis.Document,
	symbol analysis.HasTypes, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	classes, searchResult := store.SearchClasses(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := document.GetImportTable()
	for _, class := range classes {
		name, textEdit := importTable.ResolveToQualified(document, class, class.Name, word)
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

func scopedAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope analysis.HasTypes) *protocol.CompletionList {
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
		for _, class := range store.GetClasses(scopeTypeFQN) {
			methods = append(methods, analysis.SearchClassMethods(store, class, word,
				analysis.StaticMethodsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			props = append(props, analysis.SearchClassProperties(store, class, word,
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
				Detail:           HasParamsDetailWithTextEdit(method, nil),
			})
		}
		classConsts, searchResult := store.SearchClassConsts(scopeTypeFQN, word, baseSearchOptions)
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

func memberAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope analysis.HasTypes) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	for _, scopeType := range scope.GetTypes().Resolve() {
		properties := []*analysis.Property{}
		methods := []*analysis.Method{}
		for _, class := range store.GetClasses(scopeType.GetFQN()) {
			methods = append(methods, analysis.SearchClassMethods(store, class, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, scope))...)
			properties = append(properties, analysis.SearchClassProperties(store, class, word,
				analysis.PropsScopeAware(analysis.NewSearchOptions(), document, scope))...)
		}
		for _, theInterface := range store.GetInterfaces(scopeType.GetFQN()) {
			methods = append(methods, analysis.SearchInterfaceMethods(store, theInterface, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, scope))...)
		}
		for _, trait := range store.GetTraits(scopeType.GetFQN()) {
			methods = append(methods, analysis.GetTraitMethods(store, trait, word,
				analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, scope))...)
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
				Detail:           HasParamsDetailWithTextEdit(method, nil),
			})
		}
	}
	return completionList
}

func typeCompletion(store *analysis.Store, document *analysis.Document,
	symbol analysis.HasTypes, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	classes, searchResult := store.SearchClasses(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := document.GetImportTable()
	for _, class := range classes {
		label, textEdit := importTable.ResolveToQualified(document, class, class.Name, word)
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
	interfaces, searchResult := store.SearchInterfaces(word, baseSearchOptions)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, theInterface := range interfaces {
		label, textEdit := importTable.ResolveToQualified(document, theInterface, theInterface.Name, word)
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
	return completionList
}

func getDetailFromTextEdit(name analysis.TypeString, textEdit *protocol.TextEdit) string {
	detail := name.GetOriginal()
	if textEdit != nil {
		detail += "\n\n" + textEdit.NewText
	}
	return detail
}
