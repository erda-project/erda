package reconciler

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestParsePipelineOutputRef(t *testing.T) {
	reffedTask, reffedKey, err := parsePipelineOutputRef("${Dice 文档:OUTPUT:status}")
	spew.Dump(reffedTask, reffedKey)
	assert.NoError(t, err)
	assert.Equal(t, "Dice 文档", reffedTask)
	assert.Equal(t, "status", reffedKey)
}
