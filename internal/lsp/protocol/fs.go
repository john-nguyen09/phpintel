package protocol

import (
	"context"
	"io/ioutil"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/util"
	"github.com/karrick/godirwalk"
)

// FS is an interface for reading files
type FS interface {
	ConvertToURI(uri string) string
	ReadFile(ctx context.Context, uri string) ([]byte, error)
	ListFiles(ctx context.Context, base string) ([]TextDocumentIdentifier, error)
}

// URIFS is a LSP file reader which relies on xcontentProvider and xfilesProvider
// client capabilities
type URIFS struct {
	conn *jsonrpc2.Conn
}

var _ FS = (*URIFS)(nil)

// NewLSPFS creates an URIFS instance from the connection
func NewLSPFS(conn *jsonrpc2.Conn) *URIFS {
	return &URIFS{
		conn: conn,
	}
}

// ConvertToURI does nothing and return the path because it's already URI
func (f *URIFS) ConvertToURI(path string) string {
	return path
}

// ReadFile uses textDocument/xcontent request to read content of the uri
func (f *URIFS) ReadFile(ctx context.Context, uri string) ([]byte, error) {
	var textDocument TextDocumentItem
	err := f.conn.Call(ctx, "textDocument/xcontent", &ContentParams{
		TextDocument: TextDocumentIdentifier{
			URI: uri,
		},
	}, &textDocument)
	if err != nil {
		return nil, err
	}
	return []byte(textDocument.Text), nil
}

// ListFiles uses workspace/xfiles request to list files under a base
func (f *URIFS) ListFiles(ctx context.Context, base string) (docs []TextDocumentIdentifier, err error) {
	err = f.conn.Call(ctx, "workspace/xfiles", FilesParams{Base: base}, &docs)
	return
}

// FileFS uses file path
type FileFS struct{}

// NewFileFS creates a FileFS instance
func NewFileFS() *FileFS {
	return &FileFS{}
}

var _ FS = (*FileFS)(nil)

// ConvertToURI converts path to URI
func (f *FileFS) ConvertToURI(path string) string {
	return util.PathToURI(path)
}

// ReadFile reads the file from uri, given that the uri is the file path
// rather than actually the URI
func (f *FileFS) ReadFile(ctx context.Context, uri string) ([]byte, error) {
	path, err := util.URIToPath(uri)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(path)
}

// ListFiles lists files from the base
func (f *FileFS) ListFiles(ctx context.Context, base string) ([]TextDocumentIdentifier, error) {
	var results []TextDocumentIdentifier
	godirwalk.Walk(base, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				results = append(results, TextDocumentIdentifier{
					URI: path,
				})
			}
			return nil
		},
		Unsorted: true,
	})
	return results, nil
}
