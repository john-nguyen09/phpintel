package lsp

import (
	"context"
	"sort"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func (s *Server) signatureHelp(ctx context.Context, params *protocol.SignatureHelpParams) (*protocol.SignatureHelp, error) {
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
	if par := nodeStack.Parent(); par != nil && par.Type() == "array_element_initializer" {
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
