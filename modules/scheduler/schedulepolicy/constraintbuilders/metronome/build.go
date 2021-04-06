// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
