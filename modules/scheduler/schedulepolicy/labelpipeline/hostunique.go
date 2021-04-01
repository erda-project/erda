package labelpipeline

import (
	"encoding/json"

	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"

	"github.com/sirupsen/logrus"
)

// HostUniqueLabelFilter 处理 Pass1ScheduleInfo.label 中的 HOST_UNIQUE
func HostUniqueLabelFilter(
	r *labelconfig.RawLabelRuleResult, r2 *labelconfig.RawLabelRuleResult2, li *labelconfig.LabelInfo) {
	hostUniqueStr, ok := li.Label[labelconfig.HOST_UNIQUE]
	if !ok {
		return
	}
	var hostUniqueGroup [][]string
	if err := json.Unmarshal([]byte(hostUniqueStr), &hostUniqueGroup); err != nil {
		logrus.Errorf("bad input label: %v, err: %v", labelconfig.HOST_UNIQUE, err)
		return
	}
	r.HostUnique = true
	r.HostUniqueInfo = hostUniqueGroup

	r2.HasHostUnique = true
	r2.HostUnique = hostUniqueGroup
}
