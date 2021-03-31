package metronome

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/marathon"
)

// Constraints  metronome constraints
type Constraints = marathon.Constraints

type Builder struct {
	marathon marathon.Builder
}

func (b *Builder) Build(s *apistructs.ScheduleInfo2, service *apistructs.Service, _ []constraints.PodLabelsForAffinity, _ constraints.HostnameUtil) constraints.Constraints {
	return b.Build(s, service, nil, nil)
}
