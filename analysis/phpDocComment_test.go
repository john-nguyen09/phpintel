package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestPhpDocComment(t *testing.T) {
	phpDocStr := `/**
	* Run the validation routine against the given validator.
	*
	* @param  \\Illuminate\\Contracts\\Validation\\Validator|array  $validator
	* @param  \\Illuminate\\Http\\Request|null  $request
	* @return array
	*
	* @throws \\Illuminate\\Validation\\ValidationException
	*/`
	phpDoc, err := parse(phpDocStr)
	if err != nil {
		panic(err)
	}
	cupaloy.SnapshotT(t, phpDoc)
}
