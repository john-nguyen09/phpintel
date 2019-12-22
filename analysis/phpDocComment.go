package analysis

import (
	"errors"
	"regexp"
	"strings"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
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

var /* const */ stripPattern = regexp.MustCompile(`(?m)^\/\*\*[ \t]*|\s*\*\/$|^[ \t]*\*[ \t]*`)
var /* const */ tagBoundaryPattern = regexp.MustCompile(`(?:\r\n|\r|\n)@`)
var /* const */ whitespacePattern = regexp.MustCompile(`\s+`)

var /* const */ paramOrPropPattern = regexp.MustCompile(`^(@param|@property|@property-read|@property-write)\s+(\S+)\s+(\$\S+)\s*(.*)$`)
var /* const */ varPattern = regexp.MustCompile(`^(@var)\s+(\S+)(?:\s+(\$\S+))?\s*(.*)$`)
var /* const */ returnPattern = regexp.MustCompile(`^(@return)\s+(\S+)\s*(.*)$`)
var /* const */ methodPattern = regexp.MustCompile(`^(@method)\s+(?:(static)\s+)?(?:(\S+)\s+)?(\S+)\(\s*(.*)\s*\)\s*(.*)$`)
var /* const */ globalPattern = regexp.MustCompile(`^(@global)\s+(\S+)(?:\s+(\$\S+))?\s*(.*)$`)

func typeTag(tagName string, typeString string, name string, description string) tag {
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  typeString,
	}
}

func methodTag(tagName string, visibility string, returnTypeString string, name string, parameters []methodTagParam, description string) tag {
	if len(returnTypeString) == 0 {
		returnTypeString = "void"
	}

	return tag{
		TagName:     tagName,
		IsStatic:    visibility == "static",
		TypeString:  returnTypeString,
		Name:        name,
		Parameters:  parameters,
		Description: description,
	}
}

func methodParams(text string) []methodTagParam {
	params := []methodTagParam{}

	if len(text) == 0 {
		return params
	}

	paramSplit := strings.Split(text, ",")

	for _, paramBit := range paramSplit {
		var name, typeString, value string
		param := whitespacePattern.Split(strings.TrimSpace(paramBit), -1)

		switch len(param) {
		case 1:
			typeString = "mixed"
			name = param[0]
		case 2:
			typeString = param[0]
			name = param[1]
		case 4:
			typeString = param[0]
			name = param[1]
			value = param[3]
		default:
			name = ""
		}

		if len(name) != 0 {
			params = append(params, methodTagParam{
				TypeString: typeString,
				Name:       name,
				Value:      value,
			})
		}
	}

	return params
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

func parse(text string) (phpDocComment, error) {
	if len(text) == 0 {
		return phpDocComment{}, errors.New("Text is zero")
	}

	stripped := stripPattern.ReplaceAllString(text, "")
	boundaries := tagBoundaryPattern.FindAllStringIndex(stripped, -1)

	split := []string{}
	start := 0
	for _, indexRange := range boundaries {
		split = append(split, stripped[start:indexRange[0]])
		start = indexRange[1]
		if stripped[indexRange[1]-1] == '@' {
			start--
		}
	}
	if start < len(stripped) {
		split = append(split, stripped[start:])
	}

	description := ""
	if len(split) > 0 && strings.Index(split[0], "@") != 0 {
		description, split = strings.TrimSpace(split[0]), split[1:]
	}

	tags := []tag{}

	for _, part := range split {
		tag, err := parseTag(part)
		if err == nil {
			tags = append(tags, tag)
		}
	}

	if len(description) == 0 && len(tags) == 0 {
		return phpDocComment{}, errors.New("Invalid PhpDoc syntax")
	}

	return newPhpDoc(description, tags), nil
}

func parseTag(text string) (tag, error) {
	min := 4
	if min >= len(text) {
		min = len(text) - 1
	}
	substr := text[:min]
	switch substr {
	case "@par", "@pro":
		if matches := paramOrPropPattern.FindStringSubmatch(text); len(matches) > 0 {
			return typeTag(matches[1], matches[2], matches[3], matches[4]), nil
		}
	case "@var":
		if matches := varPattern.FindStringSubmatch(text); len(matches) > 0 {
			return typeTag(matches[1], matches[2], matches[3], matches[4]), nil
		}
	case "@ret":
		if matches := returnPattern.FindStringSubmatch(text); len(matches) > 0 {
			return typeTag(matches[1], matches[2], "", matches[3]), nil
		}
	case "@met":
		if matches := methodPattern.FindStringSubmatch(text); len(matches) > 0 {
			return methodTag(matches[1], matches[2], matches[3], matches[4], methodParams(matches[5]), matches[6]), nil
		}
	case "@glo":
		if matches := globalPattern.FindStringSubmatch(text); len(matches) > 0 {
			return typeTag(matches[1], matches[2], matches[3], matches[4]), nil
		}
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

func newPhpDocFromNode(document *Document, token *lexer.Token) Symbol {
	phpDocComment, err := parse(document.GetTokenText(token))
	if err != nil {
		return nil
	}
	phpDocComment.location = document.GetNodeLocation(token)
	return &phpDocComment
}
