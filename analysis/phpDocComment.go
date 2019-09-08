package analysis

import (
	"errors"
	"regexp"
	"strings"
)

type methodTagParam struct {
	TypeString string
	Name       string
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

var /* const */ paramOrPropPattern = regexp.MustCompile(`^(@param|@property|@property-read|@property-write)\s+(\S+)\s+(\$\S+)\s*([.\r\n]*)$`)
var /* const */ varPattern = regexp.MustCompile(`^(@var)\s+(\S+)(?:\s+(\$\S+))?\s*([.\r\n]*)$`)
var /* const */ returnPattern = regexp.MustCompile(`^(@return)\s+(\S+)\s*([.\r\n]*)$`)
var /* const */ methodPattern = regexp.MustCompile(`^(@method)\s+(?:(static)\s+)?(?:(\S+)\s+)?(\S+)\(\s*([.\r\n)]*)\s*\)\s*([.\r\n]*)$`)
var /* const */ globalPattern = regexp.MustCompile(`^(@global)\s+(\S+)(?:\s+(\$\S+))?\s*([.\r\n]*)$`)

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
		var name, typeString string
		param := whitespacePattern.Split(strings.TrimSpace(paramBit), -1)

		switch len(param) {
		case 1:
			typeString = "mixed"
			name = param[0]
		case 2:
			typeString = param[0]
			name = param[1]
		default:
			name = ""
		}

		if len(name) != 0 {
			params = append(params, methodTagParam{
				TypeString: typeString,
				Name:       name,
			})
		}
	}

	for i := len(params)/2 - 1; i >= 0; i-- {
		opp := len(params) - 1 - i
		params[i], params[opp] = params[opp], params[i]
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
}

func Parse(text string) (phpDocComment, error) {
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
	substr := text[:4]
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
	phpDoc.Methods = phpDoc.findTagsByTagName("@method")
	phpDoc.Vars = phpDoc.findTagsByTagName("@var")
	phpDoc.Globals = phpDoc.findTagsByTagName("@global")

	return phpDoc
}

func (d phpDocComment) findTagsByTagName(tagName string) []tag {
	tags := d.tags[:0]

	for _, tag := range tags {
		if tag.TagName == tagName {
			tags = append(tags, tag)
		}
	}

	return tags
}
