package lsp

import "github.com/pkg/errors"

import "github.com/john-nguyen09/phpintel/internal/lsp/protocol"

func StoreNotFound(uri string) error {
	return errors.Errorf("store not found for %s", uri)
}

func DocumentNotFound(uri string) error {
	return errors.Errorf("document %s not found", uri)
}

func ArgumentListNotFound(uri string, position protocol.Position) error {
	return errors.Errorf("ArgumentList not found at %s:%v", uri, position)
}
