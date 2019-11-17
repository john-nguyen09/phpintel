package main

import (
	"context"
	"os"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp"
)

func main() {
	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
