package analysis

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
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

	nameLocation protocol.Location
}

var /* const */ phpDocFirstLineRegex = regexp.MustCompile(`^\/\*\*`)
var /* const */ stripPattern = regexp.MustCompile(`(?m)^\/\*\*[ \t]*|\s*\*\/$|^[ \t]*\*[ \t]*`)

func processTypeNode(document *Document, node *phrase.Phrase) string {
	text := document.GetNodeText(node)
	if text == "" {
		return text
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	if p, ok := child.(*phrase.Phrase); ok && (p.Type == phrase.QualifiedName ||
		p.Type == phrase.FullyQualifiedName) {
		typeDecl := newTypeDeclaration(document, node)
		document.addSymbol(typeDecl)
	}
	return text
}

func processTypeUnionNode(document *Document, node *phrase.Phrase) []string {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	texts := []string{}
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.TypeDeclaration {
			texts = append(texts, processTypeNode(document, p))
		}
		child = traverser.Advance()
	}
	return texts
}

func paramOrPropTypeTag(tagName string, document *Document, p *phrase.Phrase) tag {
	ts := []string{}
	name := ""
	nameLocation := protocol.Location{}
	description := ""
	traverser := util.NewTraverser(p)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.DocumentCommentDescription:
				description = readDescriptionNode(document, p)
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			}
		} else if t, ok := child.(*lexer.Token); ok && t.Type == lexer.VariableName {
			name = document.getTokenText(t)
			nameLocation = document.GetNodeLocation(t)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),

		nameLocation: nameLocation,
	}
}

func varTag(tagName string, document *Document, node *phrase.Phrase) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.DocumentCommentDescription:
				description = readDescriptionNode(document, p)
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			}
		} else if t, ok := child.(*lexer.Token); ok && t.Type == lexer.VariableName {
			name = document.getTokenText(t)
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func returnTag(tagName string, document *Document, node *phrase.Phrase) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.DocumentCommentDescription:
				description = readDescriptionNode(document, p)
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			}
		}
	}
	return tag{
		TagName:     tagName,
		Name:        name,
		Description: description,
		TypeString:  strings.Join(ts, "|"),
	}
}

func processParamList(document *Document, node *phrase.Phrase) []methodTagParam {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	params := []methodTagParam{}
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ParameterDeclaration:
				params = append(params, methodParam(document, p))
			}
		}
		child = traverser.Advance()
	}
	return params
}

func methodTag(tagName string, document *Document, node *phrase.Phrase) tag {
	ts := []string{}
	isStatic := false
	name := ""
	nameLocation := protocol.Location{}
	params := []methodTagParam{}
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.DocumentCommentDescription:
				description = readDescriptionNode(document, p)
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			case phrase.Identifier:
				name = document.getPhraseText(p)
				nameLocation = document.GetNodeLocation(p)
			case phrase.ParameterDeclarationList:
				params = processParamList(document, p)
			}
		} else if t, ok := child.(*lexer.Token); ok {
			switch t.Type {
			case lexer.Static:
				isStatic = true
			}
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

		nameLocation: nameLocation,
	}
}

func methodParam(document *Document, node *phrase.Phrase) methodTagParam {
	ts := []string{}
	name := ""
	value := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			case phrase.ParameterValue:
				value = document.getPhraseText(p)
			}
		} else if t, ok := child.(*lexer.Token); ok && t.Type == lexer.VariableName {
			name = document.getTokenText(t)
		}
	}
	return methodTagParam{
		TypeString: strings.Join(ts, "|"),
		Name:       name,
		Value:      value,
	}
}

func globalTag(tagName string, document *Document, node *phrase.Phrase) tag {
	ts := []string{}
	name := ""
	description := ""
	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.TypeDeclaration:
				t := processTypeNode(document, p)
				if t != "" {
					ts = append(ts, t)
				}
			case phrase.TypeUnion:
				ts = append(ts, processTypeUnionNode(document, p)...)
			}
		} else if t, ok := child.(*lexer.Token); ok && t.Type == lexer.VariableName {
			name = document.getTokenText(t)
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

func readDescriptionNode(document *Document, node *phrase.Phrase) string {
	desc := document.getPhraseText(node)
	return strings.TrimSpace(stripPattern.ReplaceAllString(desc, ""))
}

func getTagName(document *Document, p *phrase.Phrase) string {
	traverser := util.NewTraverser(p)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if t, ok := child.(*lexer.Token); ok &&
			t.Type > lexer.DocumentCommentTagNameAnchorStart && t.Type < lexer.DocumentCommentTagNameAnchorEnd {
			return document.getTokenText(t)
		}
	}
	return ""
}

func parseTag(document *Document, p *phrase.Phrase) (tag, error) {
	tagName := getTagName(document, p)
	switch tagName {
	case "@param", "@property", "@property-read", "@property-write":
		paramOrProp := paramOrPropTypeTag(tagName, document, p)
		if tagName != "@param" && paramOrProp.Name == "" {
			return tag{}, fmt.Errorf("@property tags with no name")
		}
		return paramOrProp, nil
	case "@var":
		return varTag(tagName, document, p), nil
	case "@return":
		return returnTag(tagName, document, p), nil
	case "@method":
		return methodTag(tagName, document, p), nil
	case "@global":
		return globalTag(tagName, document, p), nil
	}
	return tag{}, fmt.Errorf("Unexpected tag: %s", tagName)
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

func newPhpDocFromNode(a analyser, document *Document, node *phrase.Phrase) Symbol {
	phpDoc := phpDocComment{
		location:    document.GetNodeLocation(node),
		Description: "",
		tags:        []tag{},
	}

	traverser := util.NewTraverser(node)
	for child := traverser.Advance(); child != nil; child = traverser.Advance() {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.DocumentCommentDescription:
				phpDoc.Description = readDescriptionNode(document, p)
			}
			if p.Type > phrase.DocumentCommentTagAnchorStart && p.Type < phrase.DocumentCommentTagAnchorEnd {
				tag, err := parseTag(document, p)
				if err == nil {
					phpDoc.tags = append(phpDoc.tags, tag)
				}
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
