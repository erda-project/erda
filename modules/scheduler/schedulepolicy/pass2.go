package schedulepolicy

import (
	"github.com/erda-project/erda/apistructs"
)

// Pass2ScheduleInfo request -> Pass1ScheduleInfo(LabelInfo) ------------> Pass2ScheduleInfo(apistructs.ScheduleInfo)
//                                                            filters
type Pass2ScheduleInfo apistructs.ScheduleInfo

func (p *Pass2ScheduleInfo) validate() error {
	return nil
}

type Pass2ScheduleInfo2 apistructs.ScheduleInfo2

func (p *Pass2ScheduleInfo2) validate() error {
	return nil
}
