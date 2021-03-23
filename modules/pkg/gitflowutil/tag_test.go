package gitflowutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsReleaseTag(t *testing.T) {
	validTags := []string{
		"3.4.1",
		"v3.4.1",
		"3.5.0",
		"v3.5.0",
		"3.4.1-fix-your-bug",
		"v3.5.0-fix-123-bug-456",
	}
	for _, tag := range validTags {
		assert.True(t, IsReleaseTag(tag), tag)
	}

	invalidTags := []string{
		"3",
		"v3",
		"3.4",
		"v3.4",
		"3.4.1.1",
		"v3.4.1.1",
		"v3.4.1@",
	}
	for _, tag := range invalidTags {
		assert.False(t, IsReleaseTag(tag), tag)
	}
}
