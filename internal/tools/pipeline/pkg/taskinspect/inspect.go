// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package taskinspect

import (
	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskerror"
	"github.com/erda-project/erda/pkg/metadata"
)

const (
	PipelineTaskMaxErrorPerHour = 180
)

type Inspect struct {
	Inspect     string                   `json:"inspect,omitempty"`
	Events      string                   `json:"events,omitempty"`
	MachineStat *PipelineTaskMachineStat `json:"machineStat,omitempty"`

	// Errors stores from pipeline internal, not callback(like action-agent).
	// For external errors, use taskresult.Result.Errors.
	Errors taskerror.OrderedErrors `json:"errors,omitempty"`

	// Metadata for internal use
	Metadata metadata.Metadata `json:"metadata,omitempty"`
}

func (t *Inspect) GetPBMachineStat() *basepb.PipelineTaskMachineStat {
	if t.MachineStat == nil {
		return nil
	}
	res := basepb.PipelineTaskMachineStat{}
	res.Mem = &basepb.PipelineTaskMachineMemStat{
		Total:       t.MachineStat.Mem.Total,
		Available:   t.MachineStat.Mem.Available,
		Used:        t.MachineStat.Mem.Used,
		Free:        t.MachineStat.Mem.Free,
		UsedPercent: t.MachineStat.Mem.UsedPercent,
		Buffers:     t.MachineStat.Mem.Buffers,
		Cached:      t.MachineStat.Mem.Cached,
	}
	res.Swap = &basepb.PipelineTaskMachineSwapStat{
		Total:       t.MachineStat.Swap.Total,
		Used:        t.MachineStat.Swap.Used,
		Free:        t.MachineStat.Swap.Free,
		UsedPercent: t.MachineStat.Swap.UsedPercent,
	}
	res.Pod = &basepb.PipelineTaskMachinePodStat{
		PodIP: t.MachineStat.Pod.PodIP,
	}
	res.Host = &basepb.PipelineTaskMachineHostStat{
		HostIP:          t.MachineStat.Host.HostIP,
		Hostname:        t.MachineStat.Host.Hostname,
		UptimeSec:       t.MachineStat.Host.UptimeSec,
		BootTimeSec:     t.MachineStat.Host.BootTimeSec,
		Os:              t.MachineStat.Host.OS,
		Platform:        t.MachineStat.Host.Platform,
		PlatformVersion: t.MachineStat.Host.PlatformVersion,
		KernelArch:      t.MachineStat.Host.KernelArch,
		KernelVersion:   t.MachineStat.Host.KernelVersion,
	}
	res.Load = &basepb.PipelineTaskMachineLoadStat{
		Load1:  t.MachineStat.Load.Load1,
		Load5:  t.MachineStat.Load.Load5,
		Load15: t.MachineStat.Load.Load15,
	}
	return &res
}

func (t *Inspect) IsErrorsExceed() (bool, *taskerror.Error) {
	for _, g := range t.Errors {
		if g.Ctx.CalculateFrequencyPerHour() > PipelineTaskMaxErrorPerHour {
			return true, g
		}
	}
	return false, nil
}
