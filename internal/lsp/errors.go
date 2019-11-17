package lsp

import "github.com/pkg/errors"

func StoreNotFound(uri string) error {
	return errors.Errorf("store not found for %s", uri)
}

func DocumentNotFound(uri string) error {
	return errors.Errorf("document %s not found", uri)
}
