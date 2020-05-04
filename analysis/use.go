package analysis

import (
	"github.com/john-nguyen09/go-phpparser/lexer"
	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/util"
)

type UseType int

const (
	UseClass    UseType = iota
	UseFunction         = iota
	UseConst            = iota
)

func processNamespaceUseDeclaration(document *Document, node *phrase.Phrase) Symbol {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	useType := UseClass
	prefix := ""
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.NamespaceUseClauseList:
				processNamespaceUseClauseList(document, useType, p)
			case phrase.NamespaceUseGroupClauseList:
				processNamespaceUseGroupClauseList(document, prefix, useType, p)
			case phrase.NamespaceName:
				prefix = document.getPhraseText(p)
			}
		} else if t, ok := child.(*lexer.Token); ok {
			switch t.Type {
			case lexer.Function:
				useType = UseFunction
			case lexer.Const:
				useType = UseConst
			}
		}
		child = traverser.Advance()
	}
	return nil
}

func processNamespaceUseClauseList(document *Document, useType UseType, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok && p.Type == phrase.NamespaceUseClause {
			var err error = nil
			traverser, err = traverser.Descend()
			if err != nil {
				panic(err) // Should never happen
			}

			name := ""
			alias := ""
			child = traverser.Advance()
			for child != nil {
				if p, ok := child.(*phrase.Phrase); ok {
					switch p.Type {
					case phrase.NamespaceName:
						name = document.getPhraseText(p)
					case phrase.NamespaceAliasingClause:
						alias = getAliasFromNode(document, p)
					}
				}
				child = traverser.Advance()
			}
			addUseToImportTable(document, useType, alias, name)

			traverser, err = traverser.Ascend()
			if err != nil {
				panic(err) // Should never happen
			}
		}
		traverser.Advance()
		child = traverser.Peek()
	}
}

func processNamespaceUseGroupClauseList(document *Document, prefix string, useType UseType, node *phrase.Phrase) {
	traverser := util.NewTraverser(node)
	child := traverser.Peek()
	for child != nil {
		if p, ok := child.(*phrase.Phrase); ok {
			switch p.Type {
			case phrase.NamespaceUseGroupClause:
				var err error = nil
				traverser, err = traverser.Descend()
				if err != nil {
					panic(err) // Should never happen
				}

				name := ""
				alias := ""
				child = traverser.Advance()
				for child != nil {
					if p, ok = child.(*phrase.Phrase); ok {
						switch p.Type {
						case phrase.NamespaceName:
							name = document.getPhraseText(p)
						case phrase.NamespaceAliasingClause:
							alias = getAliasFromNode(document, p)
						}
					}
					child = traverser.Advance()
				}
				name = prefix + "\\" + name
				addUseToImportTable(document, useType, alias, name)

				traverser, err = traverser.Ascend()
				if err != nil {
					panic(err) // Should never happen
				}
			}
		}
		traverser.Advance()
		child = traverser.Peek()
	}
}

func getAliasFromNode(document *Document, node *phrase.Phrase) string {
	traverser := util.NewTraverser(node)
	child := traverser.Advance()
	for child != nil {
		if t, ok := child.(*lexer.Token); ok && t.Type == lexer.Name {
			return document.getTokenText(t)
		}
		child = traverser.Advance()
	}
	return ""
}

func addUseToImportTable(document *Document, useType UseType, alias string, name string) {
	switch useType {
	case UseClass:
		document.currImportTable().addClassName(alias, name)
	case UseFunction:
		document.currImportTable().addFunctionName(alias, name)
	case UseConst:
		document.currImportTable().addConstName(alias, name)
	}
}
