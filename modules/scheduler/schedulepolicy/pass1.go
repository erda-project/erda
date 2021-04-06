package schedulepolicy

import (
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelpipeline"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// request
// -----> Pass1ScheduleInfo(aka LabelInfo)
// -----> Pass2ScheduleInfo(apistructs.ScheduleInfo)
type Pass1ScheduleInfo labelconfig.LabelInfo

func NewPass1ScheduleInfo(executorName string, executorKind string, labels map[string]string,
	configs *executortypes.ExecutorWholeConfigs, objName string,
	selectors map[string]diceyml.Selectors) Pass1ScheduleInfo {
	return Pass1ScheduleInfo{
		Label:          labels,
		ExecutorName:   executorName,
		ExecutorKind:   executorKind,
		ExecutorConfig: configs,
		OptionsPlus:    configs.PlusConfigs,
		ObjName:        objName,
		Selectors:      selectors,
	}
}

func (p *Pass1ScheduleInfo) validate() error {
	return nil
}

func (p *Pass1ScheduleInfo) toNextPass() (Pass2ScheduleInfo, Pass2ScheduleInfo2) {
	result := labelconfig.RawLabelRuleResult{}
	result2 := labelconfig.RawLabelRuleResult2{}
	for _, f := range []labelconfig.LabelPipelineFunc{
		labelpipeline.OrgLabelFilter,
		labelpipeline.WorkspaceLabelFilter,
		labelpipeline.IdentityFilter,
		labelpipeline.HostUniqueLabelFilter,
		labelpipeline.SpecificHostLabelFilter,
		labelpipeline.LocationLabelFilter,
	} {
		labelinfo := labelconfig.LabelInfo(*p)
		f(&result, &result2, &labelinfo)
	}
	return Pass2ScheduleInfo(result), Pass2ScheduleInfo2(result2)
}
