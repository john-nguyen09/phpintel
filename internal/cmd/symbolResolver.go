package cmd

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func concatDescriptionIfAvailable(content string, description string) string {
	if len(description) > 0 {
		return content + "\n" + description
	}
	return content
}

func paramsToString(params []analysis.Parameter) string {
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

func ClassToHover(ref analysis.Symbol, class analysis.Class) protocol.Hover {
	content := "# class " + class.Name.GetOriginal()
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
	content = concatDescriptionIfAvailable(content, class.GetDescription())
	theRange := ref.GetLocation().Range
	return protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:    "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func ConstToHover(ref analysis.Symbol, constant analysis.Const) protocol.Hover {
	content := "# const " + constant.Name
	if len(constant.Value) > 0 {
		content += " = " + constant.Value
	}
	content = concatDescriptionIfAvailable(content, constant.GetDescription())
	theRange := ref.GetLocation().Range
	return protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind: "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func DefineToHover(ref analysis.Symbol, define analysis.Define) protocol.Hover {
	content := "# define('" + define.GetName() + "'"
	if len(define.Value) > 0 {
		content += ", " + define.Value
	}
	content += ")"
	content = concatDescriptionIfAvailable(content, define.GetDescription())
	theRange := ref.GetLocation().Range
	return protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind: "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

func FunctionToHover(ref analysis.Symbol, function analysis.Function) protocol.Hover {
	content := "# function " + function.GetName() + "("
	content += paramsToString(function.Params)
	content += ")"
	content = concatDescriptionIfAvailable(content, function.GetDescription())
	theRange := ref.GetLocation().Range
	return protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind: "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}

// TODO: Implement TraitUseClause to use this function
func TraitToHover(ref analysis.Symbol, trait analysis.Trait) protocol.Hover {
	content := "# trait " + trait.Name.GetOriginal()
	content = concatDescriptionIfAvailable(content, trait.GetDescription())
	theRange := ref.GetLocation().Range
	return protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind: "markdown",
			Value: content,
		},
		Range: &theRange,
	}
}
