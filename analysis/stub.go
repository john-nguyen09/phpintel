package analysis

import (
	"log"
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type PhpStub struct {
	name string
	box  *rice.Box
}

func NewPhpStormStub() *PhpStub {
	box, err := rice.FindBox("phpstorm-stubs")
	if err != nil {
		log.Println(err)
		box = nil
	}
	return &PhpStub{
		name: box.Name(),
		box:  box,
	}
}

type PhpWalkFn func(string, []byte) error

func (s *PhpStub) Walk(walkFn PhpWalkFn) {
	if s.box == nil {
		return
	}
	s.box.Walk(".", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}
		bytes, err := s.box.Bytes(path)
		if err != nil {
			return nil
		}
		return walkFn(path, bytes)
	})
}

func (s *PhpStub) GetUri(path string) protocol.DocumentURI {
	return s.name + "://" + strings.ReplaceAll(path, "\\", "/")
}
