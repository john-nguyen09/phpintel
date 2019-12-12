package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp"
)

var (
	flgVersion bool
	version    string
)

func main() {
	flag.BoolVar(&flgVersion, "version", false, "Show version of the language server")
	flag.Parse()

	if flgVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
