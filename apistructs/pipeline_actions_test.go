package apistructs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataField_GetLevel(t *testing.T) {
	field := MetadataField{Name: "Error.web"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: "ERROR.web"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: " ERROR.web"}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())

	field = MetadataField{Name: "ERROR"}
	assert.Equal(t, MetadataLevelError, field.GetLevel())

	field = MetadataField{Name: "WARN.."}
	assert.Equal(t, MetadataLevelWarn, field.GetLevel())

	field = MetadataField{Name: "warn"}
	assert.Equal(t, MetadataLevelWarn, field.GetLevel())

	field = MetadataField{Name: "INFO.x"}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())

	field = MetadataField{Name: "Error.x", Level: MetadataLevelInfo}
	assert.Equal(t, MetadataLevelInfo, field.GetLevel())
}
