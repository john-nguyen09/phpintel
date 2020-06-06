package lsp

import (
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

var /* const */ triggerParameterHintsCommand = protocol.Command{
	Title:   "Trigger parameter hints",
	Command: "editor.action.triggerParameterHints",
}

func concatDescriptionIfAvailable(sb *strings.Builder, description string, writeLine bool) {
	if len(description) > 0 {
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(description)
		if err == nil {
			description = markdown
		}
		if writeLine {
			writeHorLine(sb)
		}
		sb.WriteString(description)
	}
}

func wrapPHPCode(sb *strings.Builder, fn func(*strings.Builder)) {
	sb.WriteString("```php\n<?php\n")
	fn(sb)
	sb.WriteString("\n```\n")
}

func wrapCode(sb *strings.Builder, fn func(*strings.Builder)) {
	sb.WriteString("```\n")
	fn(sb)
	sb.WriteString("\n```\n")
}

func writeHorLine(sb *strings.Builder) {
	sb.WriteString("\n____\n")
}

func writeScope(sb *strings.Builder, kind, name string) {
	sb.WriteString("\n`")
	sb.WriteString(kind)
	sb.WriteString(" ")
	sb.WriteString(name)
	sb.WriteString("`")
}

func concatParams(sb *strings.Builder, params []*analysis.Parameter) {
	paramContents := []string{}
	if len(params) > 0 {
		for _, param := range params {
			var paramContent strings.Builder
			if !param.Type.IsEmpty() {
				paramContent.WriteString(param.Type.ToString())
				paramContent.WriteString(" ")
			}
			paramContent.WriteString(param.Name)
			if param.HasValue() {
				paramContent.WriteString(" = ")
				paramContent.WriteString(param.Value)
			}
			paramContents = append(paramContents, paramContent.String())
		}
	}
	sb.WriteString(strings.Join(paramContents, ", "))
}

func formatClasses(classes []*analysis.Class) *strings.Builder {
	sb := &strings.Builder{}
	for _, class := range classes {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("class ")
			sb.WriteString(class.Name.GetOriginal())
			if !class.Extends.IsEmpty() {
				sb.WriteString(" extends ")
				sb.WriteString(class.Extends.ToString())
			}
			if len(class.Interfaces) > 0 {
				implements := []string{}
				for _, implement := range class.Interfaces {
					implements = append(implements, implement.GetOriginal())
				}
				sb.WriteString(" implements ")
				sb.WriteString(strings.Join(implements, ", "))
			}
		})
		concatDescriptionIfAvailable(sb, class.GetDescription(), true)
		writeHorLine(sb)
	}
	return sb
}

func classesToHover(ref analysis.HasTypes, classes []*analysis.Class) *protocol.Hover {
	sb := formatClasses(classes)
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func formatInterfaces(interfaces []*analysis.Interface) *strings.Builder {
	sb := &strings.Builder{}
	for _, inte := range interfaces {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("interface ")
			sb.WriteString(inte.Name.GetOriginal())
			extendStrings := []string{}
			for _, extend := range inte.Extends {
				if !extend.IsEmpty() {
					extendStrings = append(extendStrings, extend.GetOriginal())
				}
			}
			if len(extendStrings) > 0 {
				sb.WriteString(" extends ")
				sb.WriteString(strings.Join(extendStrings, ", "))
			}
		})
		concatDescriptionIfAvailable(sb, inte.GetDescription(), true)
		writeHorLine(sb)
	}
	return sb
}

func interfacesToHover(ref analysis.HasTypes, interfaces []*analysis.Interface) *protocol.Hover {
	content := formatInterfaces(interfaces)
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content.String(),
		},
		Range: &theRange,
	}
}

func formatConsts(constants []*analysis.Const) *strings.Builder {
	sb := &strings.Builder{}
	for _, constant := range constants {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("const ")
			sb.WriteString(constant.Name.GetOriginal())
			if len(constant.Value) > 0 {
				sb.WriteString(" = ")
				sb.WriteString(constant.Value)
			}
		})
		concatDescriptionIfAvailable(sb, constant.GetDescription(), true)
		writeHorLine(sb)
	}
	return sb
}

func formatDefines(defines []*analysis.Define) *strings.Builder {
	sb := &strings.Builder{}
	for _, define := range defines {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("define('")
			sb.WriteString(define.GetName())
			sb.WriteString("'")
			if len(define.Value) > 0 {
				sb.WriteString(", ")
				sb.WriteString(define.Value)
			}
			sb.WriteString(")")
		})
		concatDescriptionIfAvailable(sb, define.GetDescription(), true)
		writeHorLine(sb)
	}
	return sb
}

func functionsToHover(ref analysis.HasTypes, functions []*analysis.Function) *protocol.Hover {
	sb := &strings.Builder{}
	for _, fn := range functions {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("function ")
			sb.WriteString(fn.GetName().GetOriginal())
			sb.WriteString("(")
			concatParams(sb, fn.GetParams())
			sb.WriteString(")")
			if !fn.GetReturnTypes().IsEmpty() {
				sb.WriteString(": ")
				sb.WriteString(fn.GetReturnTypes().ToString())
			}
		})
		concatDescriptionIfAvailable(sb, fn.GetDescription(), true)
		writeHorLine(sb)
	}
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func traitsToHover(ref analysis.HasTypes, traits []*analysis.Trait) *protocol.Hover {
	sb := &strings.Builder{}
	for _, trait := range traits {
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("trait ")
			sb.WriteString(trait.Name.GetOriginal())
		})
		concatDescriptionIfAvailable(sb, trait.GetDescription(), true)
		writeHorLine(sb)
	}
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func classConstsToHover(ref analysis.HasTypes, classConsts []analysis.ClassConstWithScope) *protocol.Hover {
	sb := &strings.Builder{}
	for _, c := range classConsts {
		if c.Const == nil {
			continue
		}
		wrapPHPCode(sb, func(sb *strings.Builder) {
			sb.WriteString("const ")
			sb.WriteString(c.Const.Name)
			if len(c.Const.Value) > 0 {
				sb.WriteString(" = ")
				sb.WriteString(c.Const.Value)
			}
		})
		concatScopeIfAvailable(sb, c.Scope, true)
		concatDescriptionIfAvailable(sb, c.Const.GetDescription(), true)
		writeHorLine(sb)
	}
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func concatScopeIfAvailable(sb *strings.Builder, scope analysis.Symbol, writeLine bool) {
	var (
		scopeKind string
		scopeName string
	)
	switch v := scope.(type) {
	case *analysis.Class:
		scopeKind = "class"
		scopeName = v.Name.GetFQN()
	case *analysis.Interface:
		scopeKind = "interface"
		scopeName = v.Name.GetFQN()
	case *analysis.Trait:
		scopeKind = "trait"
		scopeName = v.Name.GetFQN()
	}
	if scopeKind != "" {
		if writeLine {
			writeHorLine(sb)
		}
		writeScope(sb, scopeKind, scopeName)
	}
}

func concatVisibility(sb *strings.Builder, visibility analysis.VisibilityModifierValue) {
	if visibility == analysis.Public {
		sb.WriteString("public")
	} else if visibility == analysis.Private {
		sb.WriteString("private")
	} else if visibility == analysis.Protected {
		sb.WriteString("protected")
	}
}

func methodsToHover(ref analysis.HasTypes, methods []analysis.MethodWithScope) *protocol.Hover {
	sb := &strings.Builder{}
	for _, m := range methods {
		method := m.Method
		if method == nil {
			continue
		}
		wrapPHPCode(sb, func(sb *strings.Builder) {
			concatVisibility(sb, method.VisibilityModifier)
			if method.IsStatic() {
				sb.WriteString(" static")
			}
			sb.WriteString(" function ")
			sb.WriteString(method.GetName())
			sb.WriteString("(")
			concatParams(sb, method.Params)
			sb.WriteString(")")
			if !method.GetReturnTypes().IsEmpty() {
				sb.WriteString(": ")
				sb.WriteString(method.GetReturnTypes().ToString())
			}
		})
		concatScopeIfAvailable(sb, m.Scope, true)
		concatDescriptionIfAvailable(sb, method.GetDescription(), true)
		writeHorLine(sb)
	}
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func propertiesToHover(ref analysis.HasTypes, properties []analysis.PropWithScope) *protocol.Hover {
	sb := &strings.Builder{}
	for _, p := range properties {
		property := p.Prop
		wrapPHPCode(sb, func(s *strings.Builder) {
			concatVisibility(sb, property.VisibilityModifier)
			if property.IsStatic() {
				sb.WriteString(" static")
			}
			sb.WriteString(" ")
			sb.WriteString(property.GetName())
			if !property.Types.IsEmpty() {
				sb.WriteString(": ")
				sb.WriteString(property.Types.ToString())
			}
		})
		concatScopeIfAvailable(sb, p.Scope, true)
		concatDescriptionIfAvailable(sb, property.GetDescription(), true)
		writeHorLine(sb)
	}
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func variableToHover(variable *analysis.Variable) *protocol.Hover {
	sb := &strings.Builder{}
	wrapCode(sb, func(sb *strings.Builder) {
		if t := variable.GetTypes(); !t.IsEmpty() {
			sb.WriteString(t.ToString())
			sb.WriteString(" ")
		}
		sb.WriteString(variable.Name)
	})
	concatDescriptionIfAvailable(sb, variable.GetDescription(), true)
	theRange := variable.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: sb.String(),
		},
		Range: &theRange,
	}
}

func hoverFromSymbol(s analysis.Symbol) *protocol.Hover {
	theRange := s.GetLocation().Range
	return &protocol.Hover{
		Range: &theRange,
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: "",
		},
	}
}

func hasParamsInsertText(f analysis.HasParams, label string) (string, protocol.InsertTextFormat, *protocol.Command) {
	if len(f.GetParams()) == 0 {
		return label + "()", protocol.PlainTextTextFormat, nil
	}
	return label + "($0)", protocol.SnippetTextFormat, &triggerParameterHintsCommand
}

func hasParamsDetailWithTextEdit(f analysis.HasParams, textEdit *protocol.TextEdit) string {
	var sb strings.Builder
	sb.WriteString(f.GetNameLabel())
	sb.WriteString("(")
	concatParams(&sb, f.GetParams())
	sb.WriteString(")")
	if textEdit != nil {
		sb.WriteString("\n")
		sb.WriteString(textEdit.NewText)
	}
	return sb.String()
}

func methodDocumentation(method analysis.MethodWithScope) protocol.MarkupContent {
	sb := &strings.Builder{}
	concatScopeIfAvailable(sb, method.Scope, false)
	sb.WriteString("\n\n")
	concatDescriptionIfAvailable(sb, method.Method.GetDescription(), false)
	return protocol.MarkupContent{
		Kind:  protocol.Markdown,
		Value: sb.String(),
	}
}

func propDocumentation(prop analysis.PropWithScope) protocol.MarkupContent {
	sb := &strings.Builder{}
	concatScopeIfAvailable(sb, prop.Scope, false)
	sb.WriteString("\n\n")
	concatDescriptionIfAvailable(sb, prop.Prop.GetDescription(), false)
	return protocol.MarkupContent{
		Kind:  protocol.Markdown,
		Value: sb.String(),
	}
}

func normaliseNamespaceName(name string) string {
	if len(name) > 0 && name[0] != '\\' {
		name = "\\" + name
	}
	return name
}

func namespaceDiff(full string, sub string) string {
	full = normaliseNamespaceName(full)
	sub = normaliseNamespaceName(sub)
	if strings.Index(full, sub) == 0 {
		return full[strings.LastIndex(full[0:len(sub)], "\\")+1:]
	}
	return full
}

func namespaceToCompletionItem(ns string, word string) protocol.CompletionItem {
	return protocol.CompletionItem{
		Kind:       protocol.ModuleCompletion,
		Label:      ns,
		InsertText: namespaceDiff(ns, word),
	}
}

func getDetailFromTextEdit(name analysis.TypeString, textEdit *protocol.TextEdit) string {
	detail := name.GetFQN()
	if textEdit != nil {
		detail += "\n\n" + textEdit.NewText
	}
	return detail
}

func classToCompletionItem(class *analysis.Class, label string, textEdit *protocol.TextEdit) protocol.CompletionItem {
	textEdits := []protocol.TextEdit{}
	if textEdit != nil {
		textEdits = append(textEdits, *textEdit)
	}
	return protocol.CompletionItem{
		Kind:                protocol.ClassCompletion,
		Label:               label,
		Documentation:       descriptionToMarkupContent(class.GetDescription()),
		AdditionalTextEdits: textEdits,
		Detail:              getDetailFromTextEdit(class.Name, textEdit),
	}
}

func interfaceToCompletionItem(intf *analysis.Interface, label string, textEdit *protocol.TextEdit) protocol.CompletionItem {
	textEdits := []protocol.TextEdit{}
	if textEdit != nil {
		textEdits = append(textEdits, *textEdit)
	}
	return protocol.CompletionItem{
		Kind:                protocol.InterfaceCompletion,
		Label:               label,
		Documentation:       descriptionToMarkupContent(intf.GetDescription()),
		AdditionalTextEdits: textEdits,
		Detail:              getDetailFromTextEdit(intf.Name, textEdit),
	}
}

func traitToCompletionItem(trait *analysis.Trait, label string, textEdit *protocol.TextEdit) protocol.CompletionItem {
	textEdits := []protocol.TextEdit{}
	if textEdit != nil {
		textEdits = append(textEdits, *textEdit)
	}
	return protocol.CompletionItem{
		Kind:                protocol.ClassCompletion,
		Label:               label,
		Documentation:       descriptionToMarkupContent(trait.GetDescription()),
		AdditionalTextEdits: textEdits,
		Detail:              getDetailFromTextEdit(trait.Name, textEdit),
	}
}

func methodToCompletionItem(m analysis.MethodWithScope) protocol.CompletionItem {
	method := m.Method
	insertText, textFormat, command := hasParamsInsertText(method, method.GetName())
	return protocol.CompletionItem{
		Kind:             protocol.MethodCompletion,
		Label:            method.GetName(),
		InsertText:       insertText,
		InsertTextFormat: textFormat,
		Command:          command,
		Documentation:    methodDocumentation(m),
		Detail:           hasParamsDetailWithTextEdit(method, nil),
	}
}

func propToCompletionItem(p analysis.PropWithScope) protocol.CompletionItem {
	prop := p.Prop
	return protocol.CompletionItem{
		Kind:          protocol.PropertyCompletion,
		Label:         prop.GetName(),
		Documentation: propDocumentation(p),
		Detail:        p.Prop.Types.ToString(),
	}
}

func classConstToCompletionItem(c analysis.ClassConstWithScope) protocol.CompletionItem {
	classConst := c.Const
	return protocol.CompletionItem{
		Kind:          protocol.ConstantCompletion,
		Label:         classConst.GetName(),
		Documentation: descriptionToMarkupContent(classConst.GetDescription()),
	}
}

func descriptionToMarkupContent(desc string) protocol.MarkupContent {
	return protocol.MarkupContent{
		Kind:  protocol.PlainText,
		Value: desc,
	}
}
