package pipelinesvc

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestValidateCreateRequest(t *testing.T) {
	req := apistructs.PipelineCreateRequestV2{
		PipelineYml:    "1.yml",
		ClusterName:    "local",
		PipelineSource: apistructs.PipelineSourceQA,
		IdentityInfo:   apistructs.IdentityInfo{InternalClient: "local"},
		Labels: map[string]string{
			"1": "01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789",
			"01234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789": "1",
			"key": "v",
		},
		NormalLabels: map[string]string{
			"1": "value",
		},
	}
	err := validateCreateRequest(&req)
	assert.NoError(t, err)
	spew.Dump(req)
}
