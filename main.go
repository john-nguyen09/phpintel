package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/john-nguyen09/phpintel/internal/jsonrpc2"
	"github.com/john-nguyen09/phpintel/internal/lsp"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

var (
	flgVersion bool
	version    string = "Unknown"
	cpuprofile string
)

func main() {
	flag.BoolVar(&flgVersion, "version", false, "Show version of the language server")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to `file`")
	flag.Parse()
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	if flgVersion {
		fmt.Println(version)
		return
	}

	stream := jsonrpc2.NewHeaderStream(os.Stdin, os.Stdout)
	ctx := context.Background()
	ctx = protocol.WithCpuProfile(ctx, cpuprofile != "")
	ctx = protocol.WithVersion(ctx, version)
	ctx, srv := lsp.NewServer(ctx, stream)
	srv.Run(ctx)
}
