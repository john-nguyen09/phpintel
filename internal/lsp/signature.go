package lsp

import (
	"context"
	"sort"
	"strings"

	"github.com/john-nguyen09/go-phpparser/phrase"
	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func (s *Server) signatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
	// ) always means hiding signature popup
	if params.Context != nil && params.Context.TriggerCharacter == ")" {
		return nil, nil
	}
	signatureHelp := &protocol.SignatureHelp{
		Signatures:      []protocol.SignatureInformation{},
		ActiveSignature: 0,
		ActiveParameter: 0,
	}
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	pos := params.TextDocumentPositionParams.Position
	nodeStack := document.NodeSpineAt(document.OffsetAtPosition(pos))
	if nodeStack.Parent().Type == phrase.ArrayInitialiserList {
		return nil, nil
	}
	argumentList, hasParamsResolvable := document.ArgumentListAndFunctionCallAt(pos)
	if argumentList == nil || hasParamsResolvable == nil {
		return nil, nil
	}
	hasParams := hasParamsResolvable.ResolveToHasParams(analysis.NewResolveContext(store, document))
	for _, hasParam := range hasParams {
		signatureHelp.Signatures = append(signatureHelp.Signatures, hasParamToSignatureInformation(hasParam))
	}

	ranges := argumentList.GetRanges()
	signatureHelp.ActiveParameter = sort.Search(len(ranges), func(i int) bool {
		return util.IsInRange(pos, ranges[i]) <= 0
	})

	return signatureHelp, nil
}

type signatureTuple struct {
	argumentList        *analysis.ArgumentList
	hasParamsResolvable analysis.HasParamsResolvable
}

func (s *Server) documentSignatures(ctx context.Context, params *protocol.TextDocumentIdentifier) ([]protocol.TextEdit, error) {
	uri := params.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	document.Lock()
	defer document.Unlock()
	document.Load()
	results := []protocol.TextEdit{}
	analysis.TraverseDocument(document, func(s analysis.Symbol) {
		if argumentList, ok := s.(*analysis.ArgumentList); ok {
			hasTypes := document.HasTypesBeforePos(argumentList.GetLocation().Range.Start)
			if resolvable, ok := hasTypes.(analysis.HasParamsResolvable); ok {
				hasParams := resolvable.ResolveToHasParams(analysis.NewResolveContext(store, document))
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
	return results, nil
}

func hasParamToSignatureInformation(hasParam analysis.HasParams) protocol.SignatureInformation {
	paramLabels := []string{}
	parameters := []protocol.ParameterInformation{}

	for _, param := range hasParam.GetParams() {
		label := ""
		if !param.Type.IsEmpty() {
			label += param.Type.ToString() + " "
		}
		label += param.Name
		if param.Value != "" {
			label += " = " + param.Value
		}
		paramLabels = append(paramLabels, label)
		parameters = append(parameters, protocol.ParameterInformation{
			Label:         label,
			Documentation: param.GetDescription(),
		})
	}

	signature := protocol.SignatureInformation{
		Label:         hasParam.GetNameLabel() + "(" + strings.Join(paramLabels, ", ") + ")",
		Documentation: hasParam.GetDescription(),
		Parameters:    parameters,
	}
	return signature
}
