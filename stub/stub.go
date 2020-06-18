package stub

import (
	"log"
	"strings"
)

// WalkFunc is a callback function that is executed on every files in a stub
type WalkFunc func(path string, content []byte) error

// Stubber contains stubs
type Stubber interface {
	// Name returns the unique name of the stubber
	Name() string
	// Walk walks all the stubs in the stubber
	Walk(WalkFunc)
	// GetURI returns the URI for the path
	GetURI(path string) string
}

var stubbers []Stubber = nil
var stubberPrefixes []string = nil

// GetStubbers initialises stubbers and returns them
func GetStubbers() []Stubber {
	if stubbers == nil {
		phpStorm, err := newPHPStormStub()
		if err == nil {
			stubbers = append(stubbers, phpStorm)
		} else {
			log.Println(err)
		}
	}
	return stubbers
}

// GetStubberPrefixes gets stubbers and returns their prefixes
func GetStubberPrefixes() []string {
	if stubberPrefixes == nil {
		for _, stubber := range GetStubbers() {
			stubberPrefixes = append(stubberPrefixes, stubber.Name()+"://")
		}
	}
	return stubberPrefixes
}

// IsStub checks if the given URI is from stub
func IsStub(uri string) bool {
	for _, prefix := range GetStubberPrefixes() {
		if strings.HasPrefix(uri, prefix) {
			return true
		}
	}
	return false
}
