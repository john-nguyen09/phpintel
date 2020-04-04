package analysis

import (
	"errors"
	"regexp"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/ast"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type methodTagParam struct {
	TypeString string
	Name       string
	Value      string
}

type tag struct {
	TagName     string
	Name        string
	Description string
	TypeString  string
	Parameters  []methodTagParam
	IsStatic    bool
}

var /* const */ phpDocFirstLineRegex = regexp.MustCompile(`^\/\*\*`)
var /* const */ stripPattern = regexp.MustCompile(`(?m)^\/\*\*[ \t]*|\s*\*\/$|^[ \t]*\*[ \t]*`)

func processTypeNode(document *Document, node *ast.Node) string {
	text := document.GetNodeText(node)
	if text == "" {
		return text
	}
	typeDecl := newTypeDeclaration(document, node)
	document.addSymbol(typeDecl)
	return text
}

func paramOrPropTypeTag(tagName string, document *Document, node *ast.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(document, child)
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "variable_name":
			name = document.GetNodeText(child)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func varTag(tagName string, document *Document, node *ast.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(document, child)
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "variable_name":
			name = document.GetNodeText(child)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func returnTag(tagName string, document *Document, node *ast.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "description":
			description = readDescriptionNode(document, child)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func methodTag(tagName string, document *Document, node *ast.Node) tag {
	ts := []string{}
	isStatic := false
	name := ""
	params := []methodTagParam{}
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(document, child)
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "static":
			isStatic = true
		case "name":
			name = document.GetNodeText(child)
		case "param":
			params = append(params, methodParam(document, child))
		}
	}

	t := strings.Join(ts, "|")
	if t == "" {
		t = "void"
	}
	return tag{
		TagName:     tagName,
		IsStatic:    isStatic,
		TypeString:  t,
		Name:        name,
		Parameters:  params,
		Description: description,
	}
}

func methodParam(document *Document, node *ast.Node) methodTagParam {
	ts := []string{}
	name := ""
	value := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "variable_name":
			name = document.GetNodeText(child)
		case "param_value":
			value = document.GetNodeText(child)
		}
	}
	return methodTagParam{
		TypeString: strings.Join(ts, "|"),
		Name:       name,
		Value:      value,
	}
}

func globalTag(tagName string, document *Document, node *ast.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			t := processTypeNode(document, child)
			if t != "" {
				ts = append(ts, t)
			}
		case "variable_name":
			name = document.GetNodeText(child)
		}
	}
	return tag{
		TagName:     tagName,
		TypeString:  strings.Join(ts, "|"),
		Name:        name,
		Description: description,
	}
}

type phpDocComment struct {
	Description string
	tags        []tag
	Returns     []tag
	Properties  []tag
	Methods     []tag
	Vars        []tag
	Globals     []tag
	location    protocol.Location

	PropertyReads  []tag
	PropertyWrites []tag
}

func readDescriptionNode(document *Document, node *ast.Node) string {
	desc := document.GetNodeText(node)
	return strings.TrimSpace(stripPattern.ReplaceAllString(desc, ""))
}

func getTagName(document *Document, node *ast.Node) string {
	traverser := util.NewTraverser(node)
	tagName := ""
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if child.Type() == "tag_name" {
			tagName = document.GetNodeText(child)
		}
	}
	return tagName
}

func parseTag(document *Document, node *ast.Node) (tag, error) {
	tagName := getTagName(document, node)
	switch tagName {
	case "@param", "@property", "@property-read", "@property-write":
		return paramOrPropTypeTag(tagName, document, node), nil
	case "@var":
		return varTag(tagName, document, node), nil
	case "@return":
		return returnTag(tagName, document, node), nil
	case "@method":
		return methodTag(tagName, document, node), nil
	case "@global":
		return globalTag(tagName, document, node), nil
	}
	return tag{}, errors.New("Unexpected tag")
}

func (d phpDocComment) findTagsByTagName(tagName string) []tag {
	tags := []tag{}

	for _, tag := range d.tags {
		if tag.TagName == tagName {
			tags = append(tags, tag)
		}
	}

	return tags
}

func (d phpDocComment) findParamTag(name string) *tag {
	for _, tag := range d.tags {
		if tag.TagName == "@param" && tag.Name == name {
			return &tag
		}
	}
	return nil
}

func (d *phpDocComment) GetLocation() protocol.Location {
	return d.location
}

func newPhpDocFromNode(document *Document, node *ast.Node) Symbol {
	if node, ok := document.injector.GetInjection(node); ok {
		phpDoc := phpDocComment{
			location:    document.GetNodeLocation(node),
			Description: "",
			tags:        []tag{},
		}

		traverser := util.NewTraverser(node)
		for child := traverser.Advance(); child != nil; child = traverser.Advance() {
			switch child.Type() {
			case "description":
				phpDoc.Description = readDescriptionNode(document, child)
			case "tag":
				tag, err := parseTag(document, child)
				if err == nil {
					phpDoc.tags = append(phpDoc.tags, tag)
				}
			}
		}

		phpDoc.Returns = phpDoc.findTagsByTagName("@return")
		phpDoc.Properties = phpDoc.findTagsByTagName("@property")
		phpDoc.PropertyReads = phpDoc.findTagsByTagName("@property-read")
		phpDoc.PropertyWrites = phpDoc.findTagsByTagName("@property-write")
		phpDoc.Methods = phpDoc.findTagsByTagName("@method")
		phpDoc.Vars = phpDoc.findTagsByTagName("@var")
		phpDoc.Globals = phpDoc.findTagsByTagName("@global")

		return &phpDoc
	}
	return nil
}
