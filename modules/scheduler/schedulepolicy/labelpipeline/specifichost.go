package labelpipeline

import (
	"strings"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

func SpecificHostLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	v, ok := li.Label[labelconfig.SPECIFIC_HOSTS]
	if !ok {
		return
	}
	result := []string{}
	hosts := strings.Split(v, ",")
	for _, host := range hosts {
		trimmedHost := strings.TrimSpace(host)
		if trimmedHost != "" {
			result = append(result, trimmedHost)
		}
	}
	r.SpecificHost = result
	r2.SpecificHost = result
}
