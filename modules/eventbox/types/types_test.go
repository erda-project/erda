package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelKeyEqual(t *testing.T) {
	assert.True(t, LabelKey("aaa-xxx").Equal("aaa-xxx"))
	assert.True(t, LabelKey("/aaa-xxx").Equal("aaa-xxx"))
	assert.True(t, LabelKey("aaa-xxx").Equal("/aaa-xxx"))
	assert.True(t, LabelKey("/aaa-xxx").Equal("/aaa-xxx"))
}
