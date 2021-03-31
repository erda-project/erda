package labelpipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

func TestLocationLabelFilter(t *testing.T) {
	r := labelconfig.RawLabelRuleResult{}
	r2 := labelconfig.RawLabelRuleResult2{}
	LocationLabelFilter(&r, &r2, &labelconfig.LabelInfo{
		Selectors: map[string]diceyml.Selectors{
			"servicename": {"location": diceyml.Selector{Values: []string{"xxx", "yyy"}}},
		},
	})
	assert.Equal(t, map[string]interface{}{"servicename": diceyml.Selector{Values: []string{"xxx", "yyy"}}}, r.Location)
}
