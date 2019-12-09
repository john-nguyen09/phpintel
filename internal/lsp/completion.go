package lsp

import (
	"context"
	"log"

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
	log.Printf("Completion: %T", symbol)
	switch s := symbol.(type) {
	case *analysis.Variable:
		completionList = variableCompletion(document, params.Position)
	case *analysis.ConstantAccess:
		completionList = nameCompletion(store, document, s.Name, params.Position)
	case *analysis.ScopedConstantAccess:
		for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
			completionList = scopedAccessCompletion(store, document, s.Name, typeString.GetFQN(), params.Position)
		}
	case *analysis.PropertyAccess:
		for _, typeString := range s.ResolveAndGetScope(store).Resolve() {
			completionList = memberAccessCompletion(store, document, s.Name, typeString.GetFQN(), params.Position)
		}
	case *analysis.ClassTypeDesignator:
		completionList = classCompletion(store, document, s.Name, params.Position)
	}
	return completionList, nil
}

func variableCompletion(document *analysis.Document, pos protocol.Position) *protocol.CompletionList {
	varTable := document.GetVariableTableAt(pos)
	completionList := &protocol.CompletionList{}
	for _, variable := range varTable.GetVariables() {
		if variable.Name == "$" {
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

func nameCompletion(store *analysis.Store, document *analysis.Document, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{}
	classes := store.SearchClasses(word)
	importTable := document.GetImportTable()
	for _, class := range classes {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.ClassCompletion,
			Label:         importTable.ResolveToQualified(class.Name),
			Documentation: class.GetDescription(),
		})
	}
	consts := store.SearchConsts(word)
	for _, constant := range consts {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.ConstantCompletion,
			Label:         importTable.ResolveToQualified(constant.Name),
			Documentation: constant.GetDescription(),
			Detail:        constant.Value,
		})
	}
	defines := store.SearchDefines(word)
	for _, define := range defines {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.ConstantCompletion,
			Label:         importTable.ResolveToQualified(define.Name),
			Documentation: define.GetDescription(),
			Detail:        define.Value,
		})
	}
	functions := store.SearchFunctions(word)
	for _, function := range functions {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.FunctionCompletion,
			Label:         importTable.ResolveToQualified(function.Name),
			Documentation: function.GetDescription(),
			Detail:        function.GetDetail(),
		})
	}
	return completionList
}

func classCompletion(store *analysis.Store, document *analysis.Document, word string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{}
	classes := store.SearchClasses(word)
	for _, class := range classes {
		completionList.Items = append(completionList.Items, protocol.CompletionItem{
			Kind:          protocol.ClassCompletion,
			Label:         class.GetName(),
			Documentation: class.GetDescription(),
		})
	}
	return completionList
}

func scopedAccessCompletion(store *analysis.Store, document *analysis.Document, word string, scope string, pos protocol.Position) *protocol.CompletionList {
	completionList := &protocol.CompletionList{}
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
	completionList := &protocol.CompletionList{}
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
		if method.IsStatic {
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
