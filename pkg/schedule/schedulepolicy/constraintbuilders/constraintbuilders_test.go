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

package constraintbuilders

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	constraints2 "github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/constraints"
	k8s2 "github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/k8s"
	marathon2 "github.com/erda-project/erda/pkg/schedule/schedulepolicy/constraintbuilders/marathon"
)

func TestBuildConstraints(t *testing.T) {
	scheduleinfo := apistructs.ScheduleInfo2{
		IsPlatform: true,
		IsUnLocked: true,
		Location: map[string]interface{}{
			"servicename1": diceyml.Selector{Values: []string{"es", "yyy"}},
		},
	}
	service := apistructs.Service{Name: "servicename1"}
	marathoncons := marathon2.Builder{}.Build(&scheduleinfo, &service, nil, nil).(*marathon2.Constraints)
	k8scons := k8s2.Builder{}.Build(&scheduleinfo, &service, []constraints2.PodLabelsForAffinity{
		{
			PodLabels: map[string]string{"app": "app1"},
			Required:  true,
		},
		{
			PodLabels: map[string]string{"appp": "appp1"},
			Required:  true,
		},
		{
			PodLabels: map[string]string{"apppp": "apppp1"},
		},
	}, nil).(*k8s2.Constraints)
	// marathon:
	// [
	//     [dice_tags LIKE .*\b(platform\b.*)]
	//     [dice_tags UNLIKE .*\b(locked\b.*)]
	//     [dice_tags LIKE .*\b(location-es\b.*|location-yyy\b.*)]
	//     [dice_tags UNLIKE .*\b(org-[^,]+\b.*)]
	//     [dice_tags UNLIKE .*\b(workspace-[^,]+\b.*)]
	//     [dice_tags UNLIKE .*\b(project-[^,]+\b.*)]
	// ]

	// k8s:
	// [
	// {[{dice/locked DoesNotExist []} {dice/location-es Exists []} {dice/platform Exists []}] []}
	// {[{dice/locked DoesNotExist []} {dice/location-yyy Exists []} {dice/platform Exists []}] []}
	// ]
	// required-anti-pod-affinity:
	// [
	//   {app In [app1]}
	//   {appp In [appp1]}
	// ]
	//
	// preferred-anti-pod-affinity:
	// [
	//  {apppp In [apppp1]}
	// ]
	assert.True(t, 6 == len(marathoncons.Cons), "%+v", marathoncons)
	assert.True(t, 2 == len(k8scons.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms), "%+v", k8scons)
	assert.False(t, 1 == len(k8scons.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution), "%+v", k8scons)
	assert.False(t, 2 == len(k8scons.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution), "%+v", k8scons)
}

func TestBuildconstraints(t *testing.T) {
	scheduleinfo := apistructs.ScheduleInfo2{
		HasOrg:       true,
		Org:          "111",
		HasWorkSpace: true,
		WorkSpaces:   []string{"dev", "test"},
		Job:          true,
	}
	marathoncons := marathon2.Builder{}.Build(&scheduleinfo, nil, nil, nil).(*marathon2.Constraints)
	k8scons := k8s2.Builder{}.Build(&scheduleinfo, nil, nil, nil).(*k8s2.Constraints)
	// marathon
	// [
	//     [dice_tags UNLIKE .*\b(platform\b.*)]
	//     [dice_tags LIKE .*\b(locked\b.*)]
	//     [dice_tags UNLIKE .*\b(location-[^,]+\b.*)]
	//     [dice_tags LIKE .*\b(org-111\b.*)]
	//     [dice_tags LIKE .*\b(workspace-dev\b.*|workspace-test\b.*)]
	//     [dice_tags LIKE .*\b(job\b.*)]
	//     [dice_tags UNLIKE .*\b(project-[^,]+\b.*)]
	// ]

	// k8s
	// [
	// {[{dice/locked Exists []} {dice/location DoesNotExist []} {dice/org-111 Exists []} {dice/job Exists []}] []}
	// ]

	assert.True(t, 7 == len(marathoncons.Cons), "%+v", marathoncons)
	assert.True(t, 1 == len(k8scons.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms), "%+v", k8scons)
	assert.True(t, 1 == len(k8scons.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution))
}

func TestBuildConstraints3(t *testing.T) {
	scheduleinfo := apistructs.ScheduleInfo2{
		Job:       true,
		PreferJob: true,
		Pack:      true,
		Stateful:  true,
		Stateless: true,
	}

	marathoncons := marathon2.Builder{}.Build(&scheduleinfo, nil, nil, nil).(*marathon2.Constraints)
	k8scons := k8s2.Builder{}.Build(&scheduleinfo, nil, nil, nil).(*k8s2.Constraints)
	// marathon:
	// [
	//     [dice_tags UNLIKE .*\b(platform\b.*)]
	//     [dice_tags LIKE .*\b(locked\b.*)]
	//     [dice_tags UNLIKE .*\b(location-[^,]+\b.*)]
	//     [dice_tags UNLIKE .*\b(org-[^,]+\b.*)]
	//     [dice_tags UNLIKE .*\b(workspace-[^,]+\b.*)]
	//     [dice_tags LIKE .*\b(job\b.*|any\b.*)]
	//     // [dice_tags LIKE .*\b(pack\b.*)]
	//     [dice_tags LIKE .*\b(stateful\b.*)]
	//     [dice_tags LIKE .*\b(stateless\b.*)]
	//     [dice_tags UNLIKE .*\b(project-[^,]+\b.*)]
	// ]

	// k8s:
	// [
	// {[{dice/locked Exists []} {dice/location DoesNotExist []} {dice/org DoesNotExist []} {dice/job Exists []}] []}
	// {[{dice/locked Exists []} {dice/location DoesNotExist []} {dice/org DoesNotExist []} {dice/pack-job Exists []}] []}
	// ]

	assert.True(t, 9 == len(marathoncons.Cons), "%+v", marathoncons)
	assert.True(t, 2 == len(k8scons.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms), "%+v", k8scons)
	assert.True(t, 1 == len(k8scons.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution), "%+v", k8scons)
}
