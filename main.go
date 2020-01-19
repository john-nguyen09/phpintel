package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

var (
	flgVersion bool
	version    string = "Unknown"
	memprofile string
)

func main() {
	flag.BoolVar(&flgVersion, "version", false, "Show version of the language server")
	flag.StringVar(&memprofile, "memprofile", "", "write mem profile to `file`")
	flag.Parse()

	if flgVersion {
		fmt.Println(version)
		return
	}

	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx = protocol.WithVersion(ctx, version)
	ctx = protocol.WithMemprofile(ctx, memprofile)
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
