package analysis

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestNestedArgumentList(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/nestedArgs.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	testOffsets := []int{
		308,
		345,
	}
	for _, testOffset := range testOffsets {
		argumentList, hasParamsResolvable := document.ArgumentListAndFunctionCallAt(document.positionAt(testOffset))
		t.Run(strconv.Itoa(testOffset), func(t *testing.T) {
			cupaloy.SnapshotT(t, argumentList, hasParamsResolvable)
		})
	}
}

func TestArgumentLists(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
$var1 = 0;
testFunction1()
++$var1;
`))
	doc.Load()
	argRanges := []protocol.Range{}
	TraverseDocument(doc, func(s Symbol) {
		if argumentList, ok := s.(*ArgumentList); ok {
			argRanges = append(argRanges, argumentList.ranges...)
		}
	}, nil)
	assert.Equal(t, []protocol.Range{
		{
			Start: protocol.Position{Line: 2, Character: 13},
			End:   protocol.Position{Line: 2, Character: 15},
		},
	}, argRanges)
}

func TestNotDuplicatedExpression(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/argumentsExpression.php")
	assert.NoError(t, err)
	document := NewDocument("test1", data)
	document.Load()

	cupaloy.SnapshotT(t, document.hasTypesSymbols())
}

func TestErrorComma(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
$abc = $DB->get_record('abc',)`))
	doc.Load()

	args, _ := doc.ArgumentListAndFunctionCallAt(protocol.Position{
		Line:      1,
		Character: 29,
	})
	assert.Equal(t, []protocol.Range{
		{Start: protocol.Position{Line: 1, Character: 22}, End: protocol.Position{Line: 1, Character: 28}},
		{Start: protocol.Position{Line: 1, Character: 29}, End: protocol.Position{Line: 1, Character: 30}},
	}, args.ranges)
}

func TestDocumentSignatures(t *testing.T) {
	data, _ := ioutil.ReadFile("../cases/TaskLog.php")
	doc := NewDocument("test1", data)
	doc.Load()

	befores := []protocol.Position{}
	TraverseDocument(doc, func(s Symbol) {
		if argumentList, ok := s.(*ArgumentList); ok {
			for _, r := range argumentList.GetArgumentRanges() {
				befores = append(befores, r.Start)
			}
		}
	}, nil)
	assert.Equal(t, []protocol.Position{
		{Line: 19, Character: 27},
		{Line: 24, Character: 32},
		{Line: 24, Character: 45},
		{Line: 29, Character: 27},
		{Line: 32, Character: 39},
		{Line: 32, Character: 47},
		{Line: 37, Character: 7},
		{Line: 37, Character: 15},
	}, befores)

	hasTypes := doc.HasTypesBeforePos(protocol.Position{Line: 37, Character: 7})
	assert.NotNil(t, hasTypes)
}

func TestDocumentArgsWithComments(t *testing.T) {
	withTestStore("test", "BenchmarkAnalysis", func(store *Store) {
		doc := indexDocumentAndGet(store, "../cases/documentArgs.php", "test1")

		results := []protocol.TextEdit{}
		resolveCtx := NewResolveContext(NewQuery(store), doc)
		TraverseDocument(doc, func(s Symbol) {
			if argumentList, ok := s.(*ArgumentList); ok {
				hasTypes := doc.HasTypesBeforePos(argumentList.GetLocation().Range.Start)
				if resolvable, ok := hasTypes.(HasParamsResolvable); ok {
					hasParams := resolvable.ResolveToHasParams(resolveCtx)
					if len(hasParams) > 0 {
						firstHasParam := hasParams[0]
						ranges := argumentList.GetArgumentRanges()
						for i, param := range firstHasParam.GetParams() {
							if i >= len(ranges) {
								break
							}
							results = append(results, protocol.TextEdit{
								NewText: param.Name,
								Range:   ranges[i],
							})
						}
					}
				}
			}
		}, nil)

		assert.Equal(t, 5, len(results))
		assert.Equal(t, []protocol.TextEdit{
			{
				NewText: "$arg1",
				Range:   protocol.RangeFromString("7:4-7:5"),
			},
			{
				NewText: "$arg2",
				Range:   protocol.RangeFromString("8:4-8:5"),
			},
			{
				NewText: "$arg3",
				Range:   protocol.RangeFromString("9:4-9:5"),
			},
			{
				NewText: "$arg4",
				Range:   protocol.RangeFromString("10:4-10:5"),
			},
			{
				NewText: "$arg5",
				Range:   protocol.RangeFromString("11:4-11:5"),
			},
		}, results)
	})
}
