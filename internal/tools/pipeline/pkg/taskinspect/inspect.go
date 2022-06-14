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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskerror"
)

const (
	PipelineTaskMaxErrorPerHour = 180
)

type PipelineTaskInspect struct {
	Inspect     string                               `json:"inspect,omitempty"`
	Events      string                               `json:"events,omitempty"`
	MachineStat *PipelineTaskMachineStat             `json:"machineStat,omitempty"`
	Errors      []*taskerror.PipelineTaskErrResponse `json:"errors,omitempty"`
}

func (t *PipelineTaskInspect) ConvertErrors() {
	for _, response := range t.Errors {
		if response.Ctx.Count > 1 {
			response.Msg = fmt.Sprintf("%s\nstartTime: %s\nendTime: %s\ncount: %d",
				response.Msg, response.Ctx.StartTime.Format("2006-01-02 15:04:05"),
				response.Ctx.EndTime.Format("2006-01-02 15:04:05"), response.Ctx.Count)
		}
	}
}

func (t *PipelineTaskInspect) IsErrorsExceed() (bool, *taskerror.PipelineTaskErrResponse) {
	for _, g := range t.Errors {
		if g.Ctx.CalculateFrequencyPerHour() > PipelineTaskMaxErrorPerHour {
			return true, g
		}
	}
	return false, nil
}

func (t *PipelineTaskInspect) AppendError(newResponses ...*taskerror.PipelineTaskErrResponse) []*taskerror.PipelineTaskErrResponse {
	if len(newResponses) == 0 {
		return t.Errors
	}
	var ordered taskerror.OrderedResponses
	for _, g := range t.Errors {
		ordered = append(ordered, g)
	}

	var newResponseOrder taskerror.OrderedResponses
	now := time.Now()
	for index, g := range newResponses {
		if g.Ctx.StartTime.IsZero() {
			g.Ctx.StartTime = now.Add(time.Duration(index) * time.Millisecond)
		}
		if g.Ctx.EndTime.IsZero() {
			g.Ctx.EndTime = now.Add(time.Duration(index) * time.Millisecond)
		}
		if g.Ctx.Count == 0 {
			g.Ctx.Count = 1
		}
		newResponseOrder = append(newResponseOrder, g)
	}
	sort.Sort(newResponseOrder)

	var lastResponse *taskerror.PipelineTaskErrResponse
	if len(ordered) != 0 {
		lastResponse = ordered[len(ordered)-1]
	}

	for _, g := range newResponseOrder {
		if lastResponse == nil {
			ordered = append(ordered, g)
			lastResponse = g
			continue
		}

		if strings.EqualFold(lastResponse.Msg, g.Msg) {
			if !g.Ctx.StartTime.IsZero() && g.Ctx.StartTime.Before(lastResponse.Ctx.StartTime) {
				lastResponse.Ctx.StartTime = g.Ctx.StartTime
			}
			if g.Ctx.EndTime.After(lastResponse.Ctx.EndTime) {
				lastResponse.Ctx.EndTime = g.Ctx.EndTime
			}
			lastResponse.Ctx.Count++
			continue
		} else {
			ordered = append(ordered, g)
			lastResponse = g
		}
	}
	return ordered
}
