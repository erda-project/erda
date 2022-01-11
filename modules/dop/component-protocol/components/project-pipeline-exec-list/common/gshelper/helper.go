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

package gshelper

import (
	"time"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

type GSHelper struct {
	gs *cptype.GlobalStateData
}

func NewGSHelper(gs *cptype.GlobalStateData) *GSHelper {
	return &GSHelper{gs: gs}
}

func (h *GSHelper) SetStatuesFilter(statues []string) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalStatuesFilter"] = statues
}

func (h *GSHelper) GetStatuesFilter() []string {
	if h.gs == nil {
		return nil
	}
	v, ok := (*h.gs)["GlobalStatuesFilter"]
	if !ok {
		return nil
	}
	if _, ok = v.([]string); ok {
		return v.([]string)
	}
	return nil
}

func (h *GSHelper) SetAppsFilter(apps []uint64) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalAppsFilter"] = apps
}

func (h *GSHelper) GetAppsFilter() []uint64 {
	if h.gs == nil {
		return nil
	}
	v, ok := (*h.gs)["GlobalAppsFilter"]
	if !ok {
		return nil
	}
	if _, ok = v.([]uint64); ok {
		return v.([]uint64)
	}
	return nil
}

func (h *GSHelper) SetExecutorsFilter(executors []string) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalExecutorsFilter"] = executors
}

func (h *GSHelper) GetExecutorsFilter() []string {
	if h.gs == nil {
		return nil
	}
	v, ok := (*h.gs)["GlobalExecutorsFilter"]
	if !ok {
		return nil
	}
	if _, ok = v.([]string); ok {
		return v.([]string)
	}
	return nil
}

func (h *GSHelper) SetPipelineNameFilter(name string) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalPipelineNameFilter"] = name
}

func (h *GSHelper) GetPipelineNameFilter() string {
	if h.gs == nil {
		return ""
	}
	v, ok := (*h.gs)["GlobalPipelineNameFilter"]
	if !ok {
		return ""
	}
	if _, ok = v.(string); ok {
		return v.(string)
	}
	return ""
}

func (h *GSHelper) SetBeginTimeStartFilter(time *time.Time) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalBeginTimeStartFilter"] = time
}

func (h *GSHelper) GetBeginTimeStartFilter() *time.Time {
	if h.gs == nil {
		return nil
	}
	v, ok := (*h.gs)["GlobalBeginTimeStartFilter"]
	if !ok {
		return nil
	}
	if _, ok = v.(*time.Time); ok {
		return v.(*time.Time)
	}
	return nil
}

func (h *GSHelper) SetBeginTimeEndFilter(time *time.Time) {
	if h.gs == nil {
		return
	}
	(*h.gs)["GlobalBeginTimeEndFilter"] = time
}

func (h *GSHelper) GetBeginTimeEndFilter() *time.Time {
	if h.gs == nil {
		return nil
	}
	v, ok := (*h.gs)["GlobalBeginTimeEndFilter"]
	if !ok {
		return nil
	}
	if _, ok = v.(*time.Time); ok {
		return v.(*time.Time)
	}
	return nil
}
