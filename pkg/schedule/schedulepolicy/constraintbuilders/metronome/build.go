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
	constraints2 "github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/constraints"
	marathon2 "github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/marathon"
)

// Constraints  metronome constraints
type Constraints = marathon2.Constraints

type Builder struct {
	marathon marathon2.Builder
}

func (b *Builder) Build(s *apistructs.ScheduleInfo2, service *apistructs.Service, _ []constraints2.PodLabelsForAffinity, _ constraints2.HostnameUtil) constraints2.Constraints {
	return b.Build(s, service, nil, nil)
}
