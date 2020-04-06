package analysis

import (
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestStaticClassTypeDesignator(t *testing.T) {
	doc1 := NewDocument("test1", []byte(`<?php
class Class1
{
    public function test1()
	{
        $var1 = new static();
        $var1->test2();
	}

    public function test2() { }
}`))
	doc1.Load()

	hasTypes := doc1.HasTypesAtPos(protocol.Position{
		Line:      6,
		Character: 11,
	})
	assert.Equal(t, "\\Class1", hasTypes.GetTypes().ToString())
}
