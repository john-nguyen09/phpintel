package analysis

import (
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

type ForeachCollection struct {
	location protocol.Location
	scope    HasTypes
}

func analyseForeachStatement(document *Document, node *phrase.Phrase) (HasTypes, bool) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	var f *ForeachCollection = nil
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.ForeachCollection:
				f = newForeachCollection(document, p)
			case phrase.ForeachValue:
				analyseForeachValue(document, f, p)
			case phrase.CompoundStatement:
				scanForChildren(document, p)
			}
		}
		child = traverser.Advance()
	}
	return nil, false
}

func newForeachCollection(document *Document, node *phrase.Phrase) *ForeachCollection {
	f := &ForeachCollection{
		location: document.GetNodeLocation(node),
	}
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			expr := scanForExpression(document, p)
			if expr != nil {
				f.setExpression(expr)
			}
		}
		child = traverser.Advance()
	}
	return f
}

func analyseForeachValue(document *Document, f *ForeachCollection, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.SimpleVariable:
				if v, shouldAdd := newVariable(document, p, true); shouldAdd {
					v.setExpression(f)
					document.addSymbol(v)
				}
			}
		}
		child = traverser.Advance()
	}
}

func (s ForeachCollection) GetLocation() protocol.Location {
	return s.location
}

func (s *ForeachCollection) setExpression(expr HasTypes) {
	s.scope = expr
}

func (s ForeachCollection) GetTypes() TypeComposite {
	types := newTypeComposite()
	if s.scope == nil {
		return types
	}
	for _, t := range s.scope.GetTypes().Resolve() {
		ok := false
		if t, ok = t.Dearray(); ok {
			types.add(t)
		}
	}
	return types
}

func (s ForeachCollection) Resolve(ctx ResolveContext) {
	if s.scope == nil {
		return
	}
	s.scope.Resolve(ctx)
}
