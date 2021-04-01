package labelpipeline

import (
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

// LocationLabelFilter LabelInfo.Selectors
func LocationLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	if r.Location == nil {
		r.Location = make(map[string]interface{})
	}
	if r2.Location == nil {
		r2.Location = make(map[string]interface{})
	}
	for service, selectors := range li.Selectors {
		selector, ok := selectors["location"]
		if !ok {
			continue
		}
		r.Location[service] = selector
		r2.Location[service] = selector
	}
}
