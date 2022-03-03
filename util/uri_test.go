package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinuxPathToURI(t *testing.T) {
	ins := []string{
		"/home/john/example1/example1.php",
	}
	expectedOuts := []string{
		"file:///home/john/example1/example1.php",
	}

	for i, in := range ins {
		assert.Equal(t, expectedOuts[i], PathToURI(in))
	}
}
