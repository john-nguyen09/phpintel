package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"

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

	f, err := os.Create("C:\\Users\\Thuan\\.phpintel\\cpu.prof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx = protocol.WithVersion(ctx, version)
	ctx = protocol.WithMemprofile(ctx, memprofile)
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
