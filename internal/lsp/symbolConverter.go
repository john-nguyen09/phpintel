package lsp

import (
	"strings"

	md "github.com/evorts/html-to-markdown"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

var /* const */ triggerParameterHintsCommand = protocol.Command{
	Title:   "Trigger parameter hints",
	Command: "editor.action.triggerParameterHints",
}

func concatDescriptionIfAvailable(content string, description string) string {
	if len(description) > 0 {
		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(description)
		if err == nil {
			description = markdown
		}
		return content + "\n___\n" + description
	}
	return content
}

func paramsToString(params []*analysis.Parameter) string {
	paramContents := []string{}
	if len(params) > 0 {
		for _, param := range params {
			paramContent := ""
			if !param.Type.IsEmpty() {
				paramContent += param.Type.ToString() + " "
			}
			paramContent += param.Name
			if param.HasValue() {
				paramContent += " = " + param.Value
			}
			paramContents = append(paramContents, paramContent)
		}
	}
	return strings.Join(paramContents, ", ")
}

func ClassToHover(ref analysis.HasTypes, class analysis.Class) *protocol.Hover {
	content := "```"
	content += "class " + class.Name.GetOriginal()
	if !class.Extends.IsEmpty() {
		content += " extends " + class.Extends.GetOriginal()
	}
	if len(class.Interfaces) > 0 {
		implements := []string{}
		for _, implement := range class.Interfaces {
			implements = append(implements, implement.GetOriginal())
		}
		content += " implements " + strings.Join(implements, ", ")
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, class.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func InterfaceToHover(ref analysis.HasTypes, theInterface analysis.Interface) *protocol.Hover {
	content := "```"
	content += "interface " + theInterface.Name.GetOriginal()

	extendStrings := []string{}
	for _, extend := range theInterface.Extends {
		if extend.IsEmpty() {
			continue
		}
		extendStrings = append(extendStrings, extend.GetOriginal())
	}
	if len(extendStrings) != 0 {
		content += " extends " + strings.Join(extendStrings, ", ")
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, theInterface.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func ConstToHover(ref analysis.HasTypes, constant analysis.Const) *protocol.Hover {
	content := "```"
	content += "const " + constant.Name.GetOriginal()
	if len(constant.Value) > 0 {
		content += " = " + constant.Value
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, constant.GetDescription()) + "```"
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func DefineToHover(ref analysis.HasTypes, define analysis.Define) *protocol.Hover {
	content := "```"
	content += "define('" + define.GetName() + "'"
	if len(define.Value) > 0 {
		content += ", " + define.Value
	}
	content += ")"
	content += "```"
	content = concatDescriptionIfAvailable(content, define.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func FunctionToHover(ref analysis.HasTypes, function analysis.Function) *protocol.Hover {
	content := "```"
	content += "function " + function.GetName().GetOriginal() + "("
	content += paramsToString(function.Params)
	content += ")"
	if !function.GetReturnTypes().IsEmpty() {
		content += ": " + function.GetReturnTypes().ToString()
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, function.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func TraitToHover(ref analysis.HasTypes, trait analysis.Trait) *protocol.Hover {
	content := "```"
	content += "trait " + trait.Name.GetOriginal()
	content += "```"
	content = concatDescriptionIfAvailable(content, trait.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func ClassConstToHover(ref analysis.HasTypes, classConst analysis.ClassConst) *protocol.Hover {
	content := "```"
	content += "const " + classConst.Name
	if len(classConst.Value) > 0 {
		content += " = " + classConst.Value
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, classConst.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func concatVisibility(content string, visibility analysis.VisibilityModifierValue) string {
	if visibility == analysis.Public {
		return content + " public"
	}
	if visibility == analysis.Private {
		return content + " private"
	}
	if visibility == analysis.Protected {
		return content + " protected"
	}
	return content
}

func MethodToHover(ref analysis.HasTypes, method analysis.Method) *protocol.Hover {
	content := "```"
	content = concatVisibility(content, method.VisibilityModifier)
	if method.IsStatic {
		content += " static"
	}
	content += " function " + method.GetName() + "("
	content += paramsToString(method.Params)
	content += ")"
	if !method.GetReturnTypes().IsEmpty() {
		content += ": " + method.GetReturnTypes().ToString()
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, method.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func PropertyToHover(ref analysis.HasTypes, property analysis.Property) *protocol.Hover {
	content := "```"
	content = concatVisibility(content, property.VisibilityModifier)
	if property.IsStatic {
		content += " static"
	}
	content += " " + property.GetName()
	if !property.Types.IsEmpty() {
		content += ": " + property.Types.ToString()
	}
	content += "```"
	content = concatDescriptionIfAvailable(content, property.GetDescription())
	theRange := ref.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func VariableToHover(variable *analysis.Variable) *protocol.Hover {
	content := "```"
	if t := variable.GetTypes(); !t.IsEmpty() {
		content += t.ToString() + " "
	}
	content += variable.Name
	content += "```"
	content = concatDescriptionIfAvailable(content, variable.GetDescription())
	theRange := variable.GetLocation().Range
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func HasParamsInsertText(f analysis.HasParams, label string) (string, protocol.InsertTextFormat, *protocol.Command) {
	if len(f.GetParams()) == 0 {
		return label + "()", protocol.PlainTextTextFormat, nil
	}
	return label + "($0)", protocol.SnippetTextFormat, &triggerParameterHintsCommand
}

func HasParamsDetailWithTextEdit(f analysis.HasParams, textEdit *protocol.TextEdit) string {
	detail := f.GetNameLabel() + "(" + paramsToString(f.GetParams()) + ")"
	if textEdit != nil {
		detail += "\n" + textEdit.NewText
	}
	return detail
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
	detail := name.GetOriginal()
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
		Documentation:       class.GetDescription(),
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
		Documentation:       intf.GetDescription(),
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
		Documentation:       trait.GetDescription(),
		AdditionalTextEdits: textEdits,
		Detail:              getDetailFromTextEdit(trait.Name, textEdit),
	}
}
