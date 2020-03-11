package analysis

import (
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestPhpDocComment(t *testing.T) {
	phpDocStr := `<?php /**
	* Run the validation routine against the given validator.
	*
	* @param  \\Illuminate\\Contracts\\Validation\\Validator|array  $validator
	* @param  \\Illuminate\\Http\\Request|null  $request
	* @return array
	*
	* @throws \\Illuminate\\Validation\\ValidationException
	*/`
	doc := NewDocument("test", []byte(phpDocStr))
	doc.Load()
	phpDoc := newPhpDocFromNode(doc, doc.GetRootNode().Child(1))
	cupaloy.SnapshotT(t, phpDoc)

	phpDocStrs := []string{
		`<?php /**
		* Information about a course that is cached in the course table 'modinfo' field (and then in
		* memory) in order to reduce the need for other database queries.
		*
		* This includes information about the course-modules and the sections on the course. It can also
		* include dynamic data that has been updated for the current user.
		*
		* Use {@link get_fast_modinfo()} to retrieve the instance of the object for particular course
		* and particular user.
		*
		* @property-read int $courseid Course ID
		* @property-read int $userid User ID
		* @property-read array $sections Array from section number (e.g. 0) to array of course-module IDs in that
		*     section; this only includes sections that contain at least one course-module
		* @property-read cm_info[] $cms Array from course-module instance to cm_info object within this course, in
		*     order of appearance
		* @property-read cm_info[][] $instances Array from string (modname) => int (instance id) => cm_info object
		* @property-read array $groups Groups that the current user belongs to. Calculated on the first request.
		*     Is an array of grouping id => array of group id => group id. Includes grouping id 0 for 'all groups'
		*/`,
	}
	for i, phpDocStr := range phpDocStrs {
		doc := NewDocument("test", []byte(phpDocStr))
		doc.Load()
		phpDoc := newPhpDocFromNode(doc, doc.GetRootNode().Child(1))
		t.Run("phpDocStr_"+strconv.Itoa(i), func(t *testing.T) {
			cupaloy.SnapshotT(t, phpDoc)
		})
	}
}

func TestPhpDocCommentSymbols(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/phpDocComment.php")
	assert.NoError(t, err)
	doc := NewDocument("test1", data)
	doc.Load()
	cupaloy.SnapshotT(t, doc.hasTypesSymbols)
}
