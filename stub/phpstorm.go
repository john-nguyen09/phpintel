package stub

import (
	"os"
	"strings"

	rice "github.com/GeertJohan/go.rice"
)

type phpStormStubber struct {
	box *rice.Box
}

var _ Stubber = (*phpStormStubber)(nil)

func newPHPStormStub() (Stubber, error) {
	box, err := rice.FindBox("phpstorm-stubs")
	if err != nil {
		return nil, err
	}
	return &phpStormStubber{
		box: box,
	}, nil
}

func (s *phpStormStubber) Name() string {
	return "phpstorm-stubs"
}

func (s *phpStormStubber) Walk(fn WalkFunc) {
	if s.box == nil {
		return
	}
	s.box.Walk("", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".php") {
			return nil
		}
		bytes, err := s.box.Bytes(path)
		if err != nil {
			return nil
		}
		return fn(path, bytes)
	})
}

func (s *phpStormStubber) GetURI(path string) string {
	return s.Name() + "://" + strings.ReplaceAll(path, "\\", "/")
}
