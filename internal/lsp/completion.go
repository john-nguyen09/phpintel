package lsp

import (
	"context"
	"strings"
	"time"

	"github.com/john-nguyen09/go-phpparser/phrase"
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
	document.Lock()
	defer document.Unlock()
	document.Load()
	var completionList *protocol.CompletionList = nil
	symbol := document.HasTypesAtPos(params.Position)
	word := document.WordAtPos(params.Position)
	nodes := document.NodeSpineAt(document.OffsetAtPosition(params.Position))
	parent := nodes.Parent()
	// log.Printf("Completion: %T %v", symbol, parent)
	switch parent.Type {
	case phrase.SimpleVariable:
		completionList = variableCompletion(document, params.Position, word)
	case phrase.NamespaceName:
		nodes.Parent()
		if nodes.Parent().Type == phrase.ConstantAccessExpression {
			completionList = nameCompletion(store, document, symbol, word)
		}
	case phrase.ErrorScopedAccessExpression, phrase.ClassConstantAccessExpression:
		if s, ok := symbol.(*analysis.ScopedConstantAccess); ok {
			completionList = scopedAccessCompletion(store, document, word, s.ResolveAndGetScope(store))
		}
	case phrase.ScopedCallExpression:
		if s, ok := symbol.(*analysis.ScopedMethodAccess); ok {
			completionList = scopedAccessCompletion(store, document, word, s.ResolveAndGetScope(store))
		}
	case phrase.ScopedPropertyAccessExpression:
		if s, ok := symbol.(*analysis.ScopedPropertyAccess); ok {
			completionList = scopedAccessCompletion(store, document, word, s.ResolveAndGetScope(store))
		}
	case phrase.Identifier:
		nodes.Parent()
		parent := nodes.Parent()
		switch parent.Type {
		case phrase.ClassConstantAccessExpression, phrase.ScopedCallExpression:
			symbol := document.HasTypesAt(util.FirstToken(&parent).Offset)
			if s, ok := symbol.(*analysis.ClassAccess); ok {
				s.Resolve(store)
				completionList = scopedAccessCompletion(store, document, word, s.Type)
			}
		}
	case phrase.PropertyAccessExpression:
		s := document.HasTypesBeforePos(params.Position)
		if s != nil {
			s.Resolve(store)
			completionList = memberAccessCompletion(store, document, word, s.GetTypes(), params.Position)
		}
	case phrase.MemberName:
		if nodes.Parent().Type == phrase.PropertyAccessExpression {
			if s, ok := symbol.(*analysis.PropertyAccess); ok {
				completionList = memberAccessCompletion(store, document, word, s.ResolveAndGetScope(store), params.Position)
			}
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
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:                protocol.FunctionCompletion,
			Label:               label,
			AdditionalTextEdits: textEdits,
			Documentation:       function.GetDescription(),
			Detail:              getDetailFromTextEdit(function.Name, textEdit),
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

func scopedAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope analysis.TypeComposite) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	for _, scopeType := range scope.Resolve() {
		scope := scopeType.GetFQN()
		properties, searchResult := store.SearchProperties(scope, word, baseSearchOptions)
		completionList.IsIncomplete = !searchResult.IsComplete
		for _, property := range properties {
			if !property.IsStatic {
				continue
			}
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.PropertyCompletion,
				Label:         property.GetName(),
				Documentation: property.GetDescription(),
			})
		}
		methods, searchResult := analysis.SearchClassMethods(store, scope, word, analysis.NewSearchOptions())
		completionList.IsIncomplete = !searchResult.IsComplete
		for _, method := range methods {
			if !method.IsStatic {
				continue
			}
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.MethodCompletion,
				Label:         method.GetName(),
				Documentation: method.GetDescription(),
			})
		}
		classConsts, searchResult := store.SearchClassConsts(scope, word, baseSearchOptions)
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

func memberAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope analysis.TypeComposite, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	for _, scopeType := range scope.Resolve() {
		scope := scopeType.GetFQN()
		properties, searchResult := store.SearchProperties(scope, word, baseSearchOptions)
		completionList.IsIncomplete = !searchResult.IsComplete
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
		methods, searchResult := store.SearchMethods(scope, word, analysis.NewSearchOptions())
		completionList.IsIncomplete = !searchResult.IsComplete
		for _, method := range methods {
			if method.Name == "__construct" {
				continue
			}
			completionList.Items = append(completionList.Items, protocol.CompletionItem{
				Kind:          protocol.MethodCompletion,
				Label:         method.GetName(),
				Documentation: method.GetDescription(),
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
	if textEdit == nil {
		return name.GetFQN()
	}
	return "use " + name.GetFQN()
}
