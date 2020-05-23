package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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
	cpuprofile string
)

func main() {
	flag.BoolVar(&flgVersion, "version", false, "Show version of the language server")
	flag.StringVar(&memprofile, "memprofile", "", "write mem profile to `file`")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.Parse()

	if flgVersion {
		fmt.Println(version)
		return
	}

	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
	}
	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx = protocol.WithVersion(ctx, version)
	ctx = protocol.WithMemprofile(ctx, memprofile)
	ctx = protocol.WithCpuprofile(ctx, cpuprofile)
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
