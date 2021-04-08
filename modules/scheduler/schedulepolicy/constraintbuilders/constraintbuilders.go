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

// Package constraintbuilders Generate various executor specific constraints (constraints) according to scheduleInfo
package constraintbuilders

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/constraints"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/k8s"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/marathon"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/constraintbuilders/metronome"
)

type constraintBuilder interface {
	// Build parameters:
	// podlabels: Used to set podAntiAffinity, for marathon, this parameter has no meaning
	Build(scheduleInfo *apistructs.ScheduleInfo2, service *apistructs.Service, podlabels []constraints.PodLabelsForAffinity, hostnameUtil constraints.HostnameUtil) constraints.Constraints
}

var (
	k8sBuilder       constraintBuilder = &k8s.Builder{}
	marathonBuilder  constraintBuilder = &marathon.Builder{}
	metronomeBuilder constraintBuilder = &metronome.Builder{}
)

// K8S build k8s constraints
func K8S(s *apistructs.ScheduleInfo2, service *apistructs.Service, podlabels []constraints.PodLabelsForAffinity, hostnameUtil constraints.HostnameUtil) *k8s.Constraints {
	return k8sBuilder.Build(s, service, podlabels, hostnameUtil).(*k8s.Constraints)
}

// Marathon build marathon constraints
func Marathon(s *apistructs.ScheduleInfo2, service *apistructs.Service) *marathon.Constraints {
	return marathonBuilder.Build(s, service, nil, nil).(*marathon.Constraints)
}

// Metronome build metronome constraints
func Metronome(s *apistructs.ScheduleInfo2, service *apistructs.Service) *metronome.Constraints {
	return metronomeBuilder.Build(s, service, nil, nil).(*metronome.Constraints)
}
