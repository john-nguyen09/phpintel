package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	document.Load()
	var completionList *protocol.CompletionList = nil
	symbol := document.SymbolAtPos(params.Position)
	// log.Printf("Completion: %T", symbol)
	switch s := symbol.(type) {
	case *analysis.Variable:
		completionList = variableCompletion(document, params.Position)
	case *analysis.ConstantAccess:
		completionList = nameCompletion(store, document, s, s.Name, params.Position)
	case *analysis.ScopedConstantAccess:
		for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
			completionList = scopedAccessCompletion(store, document, s.Name, typeString.GetFQN(), params.Position)
		}
	case *analysis.PropertyAccess:
		for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
			completionList = memberAccessCompletion(store, document, s.Name, typeString.GetFQN(), params.Position)
		}
	case *analysis.ClassTypeDesignator:
		completionList = classCompletion(store, document, s, s.Name, params.Position)
	case *analysis.TypeDeclaration:
		completionList = typeCompletion(store, document, s, s.Name, params.Position)
	}
	return completionList, nil
}

func variableCompletion(document *analysis.Document, pos protocol.Position) *protocol.CompletionList {
	varTable := document.GetVariableTableAt(pos)
	symbol := document.SymbolAtPos(pos)
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

		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.VariableCompletion,
			Label:         variable.Name,
			Documentation: variable.GetDescription(),
			Detail:        variable.GetDetail(),
		})
	}
	return completionList
}

func nameCompletion(store *analysis.Store, document *analysis.Document,
	symbol analysis.HasTypes, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	classes := store.SearchClasses(word)
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
	consts := store.SearchConsts(word)
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
	defines := store.SearchDefines(word)
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
	functions := store.SearchFunctions(word)
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
	classes := store.SearchClasses(word)
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

func scopedAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	properties := store.SearchProperties(scope, word)
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
	methods := store.SearchMethods(scope, word)
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
	classConsts := store.SearchClassConsts(scope, word)
	for _, classConst := range classConsts {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.ConstantCompletion,
			Label:         classConst.GetName(),
			Documentation: classConst.GetDescription(),
		})
	}
	return completionList
}

func memberAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	properties := store.SearchProperties(scope, word)
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
	methods := store.SearchMethods(scope, word)
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
	return completionList
}

func typeCompletion(store *analysis.Store, document *analysis.Document,
	symbol analysis.HasTypes, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	classes := store.SearchClasses(word)
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
	interfaces := store.SearchInterfaces(word)
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
