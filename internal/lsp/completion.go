package lsp

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type completionContext struct {
	doc   *analysis.Document
	query *analysis.Query
	pos   protocol.Position
}

type incompleteMemberAccess struct {
	location   protocol.Location
	scopeTypes analysis.TypeComposite
	scopeName  string
}

var _ analysis.MemberAccess = (*incompleteMemberAccess)(nil)

func fromHasTypes(h analysis.HasTypes) incompleteMemberAccess {
	var scopeName string
	if n, ok := h.(analysis.HasName); ok {
		scopeName = n.GetName()
	}
	return incompleteMemberAccess{
		location:   h.GetLocation(),
		scopeTypes: h.GetTypes(),
		scopeName:  scopeName,
	}
}

func (i incompleteMemberAccess) GetLocation() protocol.Location {
	return i.location
}

func (i incompleteMemberAccess) GetTypes() analysis.TypeComposite {
	return analysis.TypeComposite{}
}

func (i incompleteMemberAccess) Resolve(analysis.ResolveContext) {}

func (i incompleteMemberAccess) ScopeTypes() analysis.TypeComposite {
	return i.scopeTypes
}

func (i incompleteMemberAccess) ScopeName() string {
	return i.scopeName
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
	var completionList *protocol.CompletionList = nil
	pos := params.Position
	q := analysis.NewQuery(store)
	resolveCtx := analysis.NewResolveContext(q, document)
	completionCtx := &completionContext{document, q, pos}
	symbol := document.HasTypesAtPos(pos)
	word := document.WordAtPos(pos)
	nodes := document.NodeSpineAt(document.OffsetAtPosition(pos))
	// log.Printf("Completion: %s %v %T %s, kind: %v", word, pos, symbol, nodes, params.Context.TriggerKind)
	parent := nodes.Parent()
	switch parent.Type {
	case phrase.SimpleVariable:
		if nodes.Parent().Type == phrase.ScopedMemberName {
			if s, ok := symbol.(*analysis.ScopedPropertyAccess); ok {
				completionList = scopedAccessCompletion(completionCtx, word, s)
			}
			break
		}
		completionList = variableCompletion(completionCtx, resolveCtx, word)
	case phrase.NamespaceName:
		nodes.Parent()
		if nodes.Parent().Type == phrase.ConstantAccessExpression {
			completionList = nameCompletion(completionCtx, symbol, word)
		}
	case phrase.ErrorScopedAccessExpression, phrase.ClassConstantAccessExpression:
		if s, ok := symbol.(*analysis.ScopedConstantAccess); ok {
			if s.Scope != nil {
				completionList = scopedAccessCompletion(completionCtx, word, s)
			}
		}
	case phrase.ScopedCallExpression:
		if s, ok := symbol.(*analysis.ScopedMethodAccess); ok {
			if s.Scope != nil {
				completionList = scopedAccessCompletion(completionCtx, word, s)
			}
		}
	case phrase.ScopedPropertyAccessExpression:
		if s, ok := symbol.(*analysis.ScopedPropertyAccess); ok {
			if s.Scope != nil {
				completionList = scopedAccessCompletion(completionCtx, word, s)
			}
		}
	case phrase.Identifier:
		nodes.Parent()
		parent := nodes.Parent()
		switch parent.Type {
		case phrase.ClassConstantAccessExpression, phrase.ScopedCallExpression:
			symbol := document.HasTypesAt(util.FirstToken(&parent).Offset)
			if s, ok := symbol.(*analysis.ClassAccess); ok {
				s.Resolve(resolveCtx)
				completionList = scopedAccessCompletion(completionCtx, word, fromHasTypes(s))
			}
		}
	case phrase.PropertyAccessExpression:
		s := document.HasTypesBeforePos(params.Position)
		if s != nil {
			s.Resolve(resolveCtx)
			completionList = memberAccessCompletion(completionCtx, word, fromHasTypes(s))
		}
	case phrase.MemberName:
		parent := nodes.Parent()
		switch parent.Type {
		case phrase.PropertyAccessExpression:
			if s, ok := symbol.(*analysis.PropertyAccess); ok {
				if s.Scope != nil {
					s.Scope.Resolve(resolveCtx)
					completionList = memberAccessCompletion(completionCtx, word, s)
				}
			}
		case phrase.MethodCallExpression:
			if s, ok := symbol.(*analysis.MethodAccess); ok {
				if s.Scope != nil {
					s.Scope.Resolve(resolveCtx)
					completionList = memberAccessCompletion(completionCtx, word, s)
				}
			}
		}
	}
	switch s := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		completionList = classCompletion(completionCtx, s, s.Name)
	case *analysis.TypeDeclaration:
		completionList = typeCompletion(completionCtx, s.Name)
	case *analysis.InterfaceAccess:
		completionList = interfaceCompletion(completionCtx, word)
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
			Documentation: descriptionToMarkupContent(variable.GetDescription()),
			Detail:        variable.GetDetail(),
		})
	}
	return completionList
}

func nameCompletion(ctx *completionContext, symbol analysis.HasTypes, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	store := ctx.query.Store()
	opts := baseSearchOptions()
	classes, searchResult := store.SearchClasses(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := ctx.doc.ImportTableAtPos(ctx.pos)
	for _, class := range classes {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, class, class.Name, word)
		completionList.Items = append(completionList.Items, classToCompletionItem(class, label, textEdit))
	}
	interfaces, searchResult := store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, intf := range interfaces {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, intf, intf.Name, word)
		completionList.Items = append(completionList.Items, interfaceToCompletionItem(intf, label, textEdit))
	}
	consts, searchResult := store.SearchConsts(word, opts)
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
			Documentation:       descriptionToMarkupContent(constant.GetDescription()),
			Detail:              getDetailFromTextEdit(constant.Name, textEdit),
		})
	}
	defines, searchResult := store.SearchDefines(word, opts)
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
			Documentation:       descriptionToMarkupContent(define.GetDescription()),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(define.Name, textEdit),
		})
	}
	functions, searchResult := store.SearchFunctions(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, function := range functions {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, function, function.Name, word)
		insertText, textFormat, command := hasParamsInsertText(function, label)
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
			Documentation:       descriptionToMarkupContent(function.GetDescription()),
			Detail:              hasParamsDetailWithTextEdit(function, textEdit),
		})
	}
	if analysis.IsFQN(word) {
		namespaces, _ := store.SearchNamespaces(word, opts)
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
	store := ctx.query.Store()
	classes, searchResult := store.SearchClasses(word, baseSearchOptions())
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
			Documentation:       descriptionToMarkupContent(class.GetDescription()),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(class.Name, textEdit),
		})
	}
	return completionList
}

func scopedAccessCompletion(ctx *completionContext, word string, access analysis.MemberAccess) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	currentClass := ctx.doc.GetClassScopeAtSymbol(access)
	var (
		methods     []analysis.MethodWithScope
		props       []analysis.PropWithScope
		classConsts []analysis.ClassConstWithScope
	)
	for _, scopeType := range access.ScopeTypes().Resolve() {
		scopeTypeFQN := scopeType.GetFQN()
		ms := analysis.EmptyInheritedMethods()
		ps := analysis.EmptyInheritedProps()
		ccs := analysis.EmptyInheritedClassConst()
		for _, class := range ctx.query.GetClasses(scopeTypeFQN) {
			ms.Merge(ctx.query.SearchClassMethods(class, word, ms.SearchedFQNs))
			ps.Merge(ctx.query.SearchClassProps(class, word, ps.SearchedFQNs))
			ccs.Merge(ctx.query.SearchClassClassConsts(class, word, ccs.SearchedFQNs))
		}
		for _, intf := range ctx.query.GetInterfaces(scopeTypeFQN) {
			ms.Merge(ctx.query.SearchInterfaceMethods(intf, word, ms.SearchedFQNs))
			ps.Merge(ctx.query.SearchInterfaceProps(intf, word, ps.SearchedFQNs))
			ccs.Merge(ctx.query.SearchInterfaceClassConsts(intf, word, ccs.SearchedFQNs))
		}
		methods = analysis.MergeMethodWithScope(methods, ms.ReduceStatic(currentClass, access))
		props = analysis.MergePropWithScope(props, ps.ReduceStatic(currentClass, access))
		classConsts = analysis.MergeClassConstWithScope(classConsts, ccs.ReduceStatic(currentClass, access))
	}
	var scores []int
	for _, m := range methods {
		completionList.Items = append(completionList.Items, methodToCompletionItem(m))
		scores = append(scores, m.Score)
	}
	for _, p := range props {
		completionList.Items = append(completionList.Items, propToCompletionItem(p))
		scores = append(scores, p.Score)
	}
	for _, c := range classConsts {
		completionList.Items = append(completionList.Items, classConstToCompletionItem(c))
		scores = append(scores, c.Score)
	}
	sort.SliceStable(completionList.Items, func(i, j int) bool {
		return scores[i] < scores[j]
	})
	return completionList
}

func memberAccessCompletion(ctx *completionContext, word string, access analysis.MemberAccess) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	currentClass := ctx.doc.GetClassScopeAtSymbol(access)
	var (
		methods []analysis.MethodWithScope
		props   []analysis.PropWithScope
	)
	for _, scopeType := range access.ScopeTypes().Resolve() {
		scopeTypeFQN := scopeType.GetFQN()
		ms := analysis.EmptyInheritedMethods()
		ps := analysis.EmptyInheritedProps()
		for _, class := range ctx.query.GetClasses(scopeTypeFQN) {
			ms.Merge(ctx.query.SearchClassMethods(class, word, ms.SearchedFQNs))
			ps.Merge(ctx.query.SearchClassProps(class, word, ps.SearchedFQNs))
		}
		for _, intf := range ctx.query.GetInterfaces(scopeTypeFQN) {
			ms.Merge(ctx.query.SearchInterfaceMethods(intf, word, ms.SearchedFQNs))
			ps.Merge(ctx.query.SearchInterfaceProps(intf, word, ps.SearchedFQNs))
		}
		methods = analysis.MergeMethodWithScope(methods, ms.ReduceAccess(currentClass, access))
		props = analysis.MergePropWithScope(props, ps.ReduceAccess(currentClass, access))
	}
	var scores []int
	for _, m := range methods {
		completionList.Items = append(completionList.Items, methodToCompletionItem(m))
		scores = append(scores, m.Score)
	}
	for _, p := range props {
		item := propToCompletionItem(p)
		item.Label = item.Label[1:]
		completionList.Items = append(completionList.Items, item)
		scores = append(scores, p.Score)
	}
	sort.SliceStable(completionList.Items, func(i, j int) bool {
		return scores[i] < scores[j]
	})
	return completionList
}

func typeCompletion(ctx *completionContext, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: true,
	}
	opts := baseSearchOptions()
	store := ctx.query.Store()
	classes, searchResult := store.SearchClasses(word, opts)
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
			Documentation:       descriptionToMarkupContent(class.GetDescription()),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(class.Name, textEdit),
		})
	}
	interfaces, searchResult := store.SearchInterfaces(word, opts)
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
			Documentation:       descriptionToMarkupContent(theInterface.GetDescription()),
			AdditionalTextEdits: textEdits,
			Detail:              getDetailFromTextEdit(theInterface.Name, textEdit),
		})
	}
	if analysis.IsFQN(word) {
		namespaces, _ := store.SearchNamespaces(word, opts)
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
	store := ctx.query.Store()
	namespaces, _ := store.SearchNamespaces(t.GetFQN(), opts)
	for _, ns := range namespaces {
		completionList.Items = append(completionList.Items, namespaceToCompletionItem(ns, word))
	}
	classes, searchResult := store.SearchClasses(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, class := range classes {
		completionList.Items = append(completionList.Items, classToCompletionItem(class, class.Name.GetOriginal(), nil))
	}
	interfaces, searchResult := store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, intf := range interfaces {
		completionList.Items = append(completionList.Items, interfaceToCompletionItem(intf, intf.Name.GetOriginal(), nil))
	}
	traits, searchResult := store.SearchTraits(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	for _, trait := range traits {
		completionList.Items = append(completionList.Items, traitToCompletionItem(trait, trait.Name.GetOriginal(), nil))
	}
	return completionList
}

func interfaceCompletion(ctx *completionContext, word string) *protocol.CompletionList {
	completionList := &protocol.CompletionList{
		IsIncomplete: false,
	}
	opts := baseSearchOptions()
	store := ctx.query.Store()
	interfaces, searchResult := store.SearchInterfaces(word, opts)
	completionList.IsIncomplete = !searchResult.IsComplete
	importTable := ctx.doc.ImportTableAtPos(ctx.pos)
	for _, intf := range interfaces {
		label, textEdit := importTable.ResolveToQualified(ctx.doc, intf, intf.Name, word)
		completionList.Items = append(completionList.Items, interfaceToCompletionItem(intf, label, textEdit))
	}
	return completionList
}
