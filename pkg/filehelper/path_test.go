package filehelper

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestAbs2Rel(t *testing.T) {
	path1 := "/"
	path1 = Abs2Rel(path1)
	assert.Equal(t, ".", path1)

	path2 := "//testdata/"
	path2 = Abs2Rel(path2)
	assert.Equal(t, "testdata", path2)
}
