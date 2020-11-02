package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestCatchClause(t *testing.T) {
	data := []byte(`<?php
namespace App;

use Exception;

try {
} catch (\NotFoundException $ex) {
} catch (\HttpException | Exception $ex) {
} catch (\Throwable $ex) {
}`)
	doc := NewDocument("test1", data)
	doc.Load()
	results := []*ClassAccess{}
	tra := newTraverser()
	tra.traverseDocument(doc, func(_ *traverser, symbol Symbol, _ []Symbol) {
		if classAccess, ok := symbol.(*ClassAccess); ok {
			results = append(results, classAccess)
		}
	})
	cupaloy.SnapshotT(t, results)
}
