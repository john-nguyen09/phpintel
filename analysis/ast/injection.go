package ast

import (
	"fmt"
	"math"
	"regexp"

	"github.com/john-nguyen09/phpintel/analysis/ast/php"
	"github.com/john-nguyen09/phpintel/analysis/ast/phpdoc"
	sitter "github.com/smacker/go-tree-sitter"
)

// InjectionConfig contains information for injections
type InjectionConfig struct {
	lang                *sitter.Language
	query               *sitter.Query
	contentCaptureIndex uint32
}

// InjectionConfigCreator is a function to create the config
// from language name
type InjectionConfigCreator func(string) InjectionConfig

// NewConfig create injection config
func NewConfig(lang *sitter.Language, injectionQuery []byte) (InjectionConfig, error) {
	if injectionQuery == nil {
		return InjectionConfig{
			lang: lang,
		}, nil
	}
	query, err := sitter.NewQuery(injectionQuery, lang)
	if err != nil {
		return InjectionConfig{}, err
	}
	contentCaptureIndex := uint32(0)
	for i := uint32(0); i < query.CaptureCount(); i++ {
		switch query.CaptureNameForId(i) {
		case "injection.content":
			contentCaptureIndex = i
		}
	}
	return InjectionConfig{
		lang:                lang,
		query:               query,
		contentCaptureIndex: contentCaptureIndex,
	}, nil
}

// InjectionLayer is a layer containing information of a language
type InjectionLayer struct {
	config        InjectionConfig
	tree          *sitter.Tree
	ranges        []sitter.Range
	injectedNodes map[string]*Node
}

type configRangesTuple struct {
	config InjectionConfig
	ranges []sitter.Range
}

type languageContentNodesTuple struct {
	languageName string
	contentNodes []*Node
}

// NewInjectionLayer creates all the injection layers
func NewInjectionLayer(source []byte,
	oldLayers []*InjectionLayer, edit sitter.EditInput, injector *Injector,
	configCreator InjectionConfigCreator, config InjectionConfig, ranges []sitter.Range) []*InjectionLayer {
	result := []*InjectionLayer{}
	queue := []configRangesTuple{}
	var prevLayer *InjectionLayer = nil
	prevInjectedNodes := map[string]*Node{}
	for i := 0; ; i++ {
		var oldTree *sitter.Tree = nil
		if i < len(oldLayers) && oldLayers[i].config.lang == config.lang {
			oldTree = oldLayers[i].tree
			oldTree.Edit(edit)
		}
		injector.parser.SetIncludedRanges(ranges)
		injector.parser.SetLanguage(config.lang)
		tree := injector.parser.Parse(oldTree, source)
		rootNode := FromSitterNode(tree.RootNode())
		cursor := sitter.NewQueryCursor()
		if prevLayer != nil {
			len := int(rootNode.ChildCount())
			for i := 0; i < len; i++ {
				child := rootNode.Child(i)
				if parent, ok := prevInjectedNodes[getNodeRangeString(child)]; ok {
					prevLayer.injectedNodes[getNodeRangeString(parent)] = child
				}
			}
		}
		if query := config.query; query != nil {
			injectionsByPatternIndex := make([]languageContentNodesTuple, int(query.PatternCount()))
			cursor.Exec(config.query, rootNode.n)
			for mat, ok := cursor.NextMatch(); ok; mat, ok = cursor.NextMatch() {
				entry := &injectionsByPatternIndex[mat.PatternIndex]
				languageName, contentNode := injectionForMatch(config, query, mat, source)
				if languageName != "" {
					entry.languageName = languageName
				}
				if contentNode != nil {
					entry.contentNodes = append(entry.contentNodes, contentNode)
				}
			}
			for _, entry := range injectionsByPatternIndex {
				if entry.languageName == "" {
					continue
				}
				nextConfig := configCreator(entry.languageName)
				if nextConfig.lang == nil {
					continue
				}
				nextRanges := []sitter.Range{}
				for _, node := range entry.contentNodes {
					prevInjectedNodes[getNodeRangeString(node)] = node
					nextRanges = append(nextRanges, sitter.Range{
						StartPoint: node.StartPoint(),
						EndPoint:   node.EndPoint(),
						StartByte:  node.StartByte(),
						EndByte:    node.EndByte(),
					})
				}
				if len(nextRanges) != 0 {
					queue = append(queue, configRangesTuple{nextConfig, nextRanges})
				}
			}
		}
		layer := &InjectionLayer{
			config:        config,
			tree:          tree,
			ranges:        ranges,
			injectedNodes: map[string]*Node{},
		}
		prevLayer = layer
		result = append(result, layer)
		if len(queue) == 0 {
			break
		} else {
			var tuple configRangesTuple
			tuple, queue = queue[len(queue)-1], queue[:len(queue)-1]
			config = tuple.config
			ranges = tuple.ranges
		}
	}
	return result
}

func getNodeRangeString(n *Node) string {
	return fmt.Sprintf("%d-%d", n.StartByte(), n.EndByte())
}

func injectionForMatch(config InjectionConfig, query *sitter.Query, mat *sitter.QueryMatch, source []byte) (string, *Node) {
	contentCaptureIndex := config.contentCaptureIndex
	var contentNode *Node = nil
	for _, capture := range mat.Captures {
		if capture.Index == contentCaptureIndex {
			contentNode = FromSitterNode(capture.Node)
		}
	}
	languageName := ""
	injectionRegex := ""
	for key, value := range getQueryProperties(query) {
		switch key {
		case "injection.language":
			languageName = value
		case "injection.regex":
			injectionRegex = value
		}
	}
	if injectionRegex != "" {
		regex := regexp.MustCompile(injectionRegex)
		if !regex.MatchString(contentNode.Content(source)) {
			return "", nil
		}
	}
	return languageName, contentNode
}

// Injector is a wrapper for injection support in tree-sitter
type Injector struct {
	parser *sitter.Parser
	layers []*InjectionLayer
}

// NewPHPInjector creates an injector for PHP
func NewPHPInjector(source []byte) *Injector {
	inj := &Injector{
		parser: sitter.NewParser(),
	}
	inj.layers = NewInjectionLayer(source, nil, sitter.EditInput{}, inj, createConfig, createConfig("php"), []sitter.Range{
		sitter.Range{
			StartByte:  0,
			EndByte:    math.MaxUint32,
			StartPoint: sitter.Point{uint32(0), uint32(0)},
			EndPoint:   sitter.Point{math.MaxUint32, math.MaxUint32},
		},
	})
	return inj
}

// MainRootNode returns the main language root node
func (i *Injector) MainRootNode() *Node {
	return FromSitterNode(i.layers[0].tree.RootNode())
}

// GetInjection checks if the node has injection
func (i *Injector) GetInjection(node *Node) (*Node, bool) {
	for _, layer := range i.layers {
		if childNode, ok := layer.injectedNodes[getNodeRangeString(node)]; ok {
			return childNode, ok
		}
	}
	return nil, false
}

// Edit returns a new injector which reflects the modification
func (i *Injector) Edit(edit sitter.EditInput, source []byte) *Injector {
	inj := &Injector{
		parser: i.parser,
	}
	inj.layers = NewInjectionLayer(source, i.layers, edit, inj, createConfig, createConfig("php"), []sitter.Range{
		sitter.Range{
			StartByte:  0,
			EndByte:    math.MaxUint32,
			StartPoint: sitter.Point{uint32(0), uint32(0)},
			EndPoint:   sitter.Point{math.MaxUint32, math.MaxUint32},
		},
	})
	return inj
}

var configMap map[string]InjectionConfig = map[string]InjectionConfig{}

func init() {
	config, err := NewConfig(phpdoc.GetLanguage(), nil)
	if err != nil {
		panic(err)
	}
	configMap["phpdoc"] = config

	injectionQuery := php.GetInjectionQuery()
	config, err = NewConfig(php.GetLanguage(), injectionQuery)
	if err != nil {
		panic(err)
	}
	configMap["php"] = config
}

func createConfig(languageName string) InjectionConfig {
	if config, ok := configMap[languageName]; ok {
		return config
	}
	return InjectionConfig{}
}

func getQueryProperties(query *sitter.Query) map[string]string {
	props := map[string]string{}
	for i := uint32(0); i < query.CaptureCount(); i++ {
		predicateSteps := splitPredicateSteps(query.PredicatesForPattern(i), func(s sitter.QueryPredicateStep) bool {
			return s.Type == sitter.QueryPredicateStepTypeDone
		})
		for _, p := range predicateSteps {
			if len(p) == 0 {
				continue
			}
			if p[0].Type != sitter.QueryPredicateStepTypeString {
				continue
			}
			operatorName := query.StringValueForId(p[0].ValueId)
			switch operatorName {
			case "set!":
				key, value := parseProperty(query, p[1:])
				props[key] = value
			}
		}
	}
	return props
}

func splitPredicateSteps(steps []sitter.QueryPredicateStep, fn func(sitter.QueryPredicateStep) bool) [][]sitter.QueryPredicateStep {
	results := [][]sitter.QueryPredicateStep{}
	prevIndex := 0
	for i, s := range steps {
		if fn(s) {
			results = append(results, steps[prevIndex:i])
			prevIndex = i + 1
		}
	}
	return results
}

func parseProperty(query *sitter.Query, args []sitter.QueryPredicateStep) (string, string) {
	if len(args) == 0 {
		return "", ""
	}
	key := query.StringValueForId(args[0].ValueId)
	value := ""
	if len(args) >= 2 {
		value = query.StringValueForId(args[1].ValueId)
	}
	return key, value
}
