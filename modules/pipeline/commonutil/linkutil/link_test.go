package linkutil

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPipelineLink(t *testing.T) {
	fakeBdl := bundle.New()

	type invalidCase struct {
		p    spec.Pipeline
		desc string
	}

	invalidCases := []invalidCase{
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID: "a",
				},
			},
			desc: "invalid orgID",
		},
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID:     "1",
					apistructs.LabelProjectID: "a",
				},
			},
			desc: "invalid projectID",
		},
		{
			p: spec.Pipeline{
				Labels: map[string]string{
					apistructs.LabelOrgID:     "1",
					apistructs.LabelProjectID: "1",
					apistructs.LabelAppID:     "a",
				},
			},
			desc: "invalid appID",
		},
	}

	for _, c := range invalidCases {
		valid, link := GetPipelineLink(fakeBdl, c.p)
		assert.False(t, valid)
		assert.Empty(t, link)
	}
}
