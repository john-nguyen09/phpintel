package analysis

import (
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestInsertUse(t *testing.T) {
	data := []byte(`<?php
namespace Namespace1;

App`)
	document := NewDocument("test1", data)
	document.Load()

	ctx := GetInsertUseContext(document)
	pos, ok := ctx.GetInsertPosition()
	assert.True(t, ok)
	assert.Equal(t, protocol.Position{
		Line:      1,
		Character: 21,
	}, pos)
	assert.Equal(t, "", getIndentation(document, ctx.GetInsertAfterNode()))
	assert.Equal(t, "\n", getNewLine(document, ctx.GetInsertAfterNode()))

	document = NewDocument("test2", []byte(`<?php
	namespace Namespace1;`))
	document.Load()
	ctx = GetInsertUseContext(document)
	assert.Equal(t, "\t", getIndentation(document, ctx.GetInsertAfterNode()))
}
