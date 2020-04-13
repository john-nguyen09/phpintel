package stub

import (
	"log"
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
