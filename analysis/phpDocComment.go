package analysis

import (
	"errors"
	"regexp"
	"strings"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/phpdoc"
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

func paramOrPropTypeTag(tagName string, input []byte, node *sitter.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(input, child)
		case "type":
			ts = append(ts, child.Content(input))
		case "variable_name":
			name = child.Content(input)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func varTag(tagName string, input []byte, node *sitter.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(input, child)
		case "type":
			ts = append(ts, child.Content(input))
		case "variable_name":
			name = child.Content(input)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func returnTag(tagName string, input []byte, node *sitter.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			ts = append(ts, child.Content(input))
		case "description":
			description = readDescriptionNode(input, child)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func methodTag(tagName string, input []byte, node *sitter.Node) tag {
	ts := []string{}
	isStatic := false
	name := ""
	params := []methodTagParam{}
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(input, child)
		case "type":
			ts = append(ts, child.Content(input))
		case "static":
			isStatic = true
		case "name":
			name = child.Content(input)
		case "param":
			params = append(params, methodParam(input, child))
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

func methodParam(input []byte, node *sitter.Node) methodTagParam {
	ts := []string{}
	name := ""
	value := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			ts = append(ts, child.Content(input))
		case "variable_name":
			name = child.Content(input)
		case "param_value":
			value = child.Content(input)
		}
	}
	return methodTagParam{
		TypeString: strings.Join(ts, "|"),
		Name:       name,
		Value:      value,
	}
}

func globalTag(tagName string, input []byte, node *sitter.Node) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "type":
			ts = append(ts, child.Content(input))
		case "variable_name":
			name = child.Content(input)
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

func readDescriptionNode(input []byte, node *sitter.Node) string {
	desc := node.Content(input)
	return strings.TrimSpace(stripPattern.ReplaceAllString(desc, ""))
}

func parse(text string) (phpDocComment, error) {
	if len(text) == 0 {
		return phpDocComment{}, errors.New("Text is zero")
	}
	description := ""
	tags := []tag{}

	parser := sitter.NewParser()
	parser.SetLanguage(phpdoc.GetLanguage())
	input := []byte(text)
	tree := parser.Parse(nil, input)
	node := tree.RootNode()
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		switch child.Type() {
		case "description":
			description = readDescriptionNode(input, child)
		case "tag":
			tag, err := parseTag(input, child)
			if err == nil {
				tags = append(tags, tag)
			}
		}
	}

	return newPhpDoc(description, tags), nil
}

func getTagName(input []byte, node *sitter.Node) string {
	traverser := util.NewTraverser(node)
	tagName := ""
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if child.Type() == "tag_name" {
			tagName = child.Content(input)
		}
	}
	return tagName
}

func parseTag(input []byte, node *sitter.Node) (tag, error) {
	tagName := getTagName(input, node)
	switch tagName {
	case "@param", "@property", "@property-read", "@property-write":
		return paramOrPropTypeTag(tagName, input, node), nil
	case "@var":
		return varTag(tagName, input, node), nil
	case "@return":
		return returnTag(tagName, input, node), nil
	case "@method":
		return methodTag(tagName, input, node), nil
	case "@global":
		return globalTag(tagName, input, node), nil
	}
	return tag{}, errors.New("Unexpected tag")
}

func newPhpDoc(description string, tags []tag) phpDocComment {
	phpDoc := phpDocComment{
		Description: description,
		tags:        tags,
	}

	phpDoc.Returns = phpDoc.findTagsByTagName("@return")
	phpDoc.Properties = phpDoc.findTagsByTagName("@property")
	phpDoc.PropertyReads = phpDoc.findTagsByTagName("@property-read")
	phpDoc.PropertyWrites = phpDoc.findTagsByTagName("@property-write")
	phpDoc.Methods = phpDoc.findTagsByTagName("@method")
	phpDoc.Vars = phpDoc.findTagsByTagName("@var")
	phpDoc.Globals = phpDoc.findTagsByTagName("@global")

	return phpDoc
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

func newPhpDocFromNode(document *Document, token *sitter.Node) Symbol {
	text := document.GetNodeText(token)
	if !phpDocFirstLineRegex.MatchString(text) {
		return nil
	}
	phpDocComment, err := parse(text)
	if err != nil {
		return nil
	}
	phpDocComment.location = document.GetNodeLocation(token)
	return &phpDocComment
}
