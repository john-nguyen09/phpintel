package cmd

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func concatDescriptionIfAvailable(content string, description string) string {
	if len(description) > 0 {
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
			if len(param.Value) > 0 {
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

// TODO: Implement TraitUseClause to use this function
func TraitToHover(ref analysis.HasTypes, trait analysis.Trait) *protocol.Hover {
	content := "```"
	content += "trait " + trait.Name.GetOriginal()
	content += "```"
	content = concatDescriptionIfAvailable(content, trait.GetDescription()) + "```"
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
	if !variable.GetTypes().IsEmpty() {
		content += variable.GetTypes().ToString() + " "
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
