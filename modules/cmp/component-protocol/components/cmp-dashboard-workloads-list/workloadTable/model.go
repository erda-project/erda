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

package workloadTable

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/cmp/cmp_interface"

	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentWorkloadTable struct {
	base.DefaultProvider
	sdk    *cptype.SDK
	ctx    context.Context
	server cmp_interface.SteveServer

	Type       string                 `json:"type,omitempty"`
	State      State                  `json:"state,omitempty"`
	Data       Data                   `json:"data,omitempty"`
	Props      Props                  `json:"props,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type State struct {
	ClusterName           string      `json:"clusterName,omitempty"`
	CountValues           CountValues `json:"countValues,omitempty"`
	PageNo                uint64      `json:"pageNo"`
	PageSize              uint64      `json:"pageSize"`
	Sorter                Sorter      `json:"sorterData,omitempty"`
	Total                 uint64      `json:"total"`
	Values                Values      `json:"values,omitempty"`
	WorkloadTableURLQuery string      `json:"workloadTable__urlQuery,omitempty"`
}

type CountValues struct {
	DeploymentsCount Count `json:"deploymentsCount,omitempty"`
	DaemonSetCount   Count `json:"daemonSetCount,omitempty"`
	StatefulSetCount Count `json:"statefulSetCount,omitempty"`
	JobCount         Count `json:"jobCount,omitempty"`
	CronJobCount     Count `json:"cronJobCount,omitempty"`
}

type Count struct {
	Active    int `json:"active"`
	Error     int `json:"error"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type Sorter struct {
	Field string `json:"field,omitempty"`
	Order string `json:"order,omitempty"`
}

type Values struct {
	Namespace []string `json:"namespace,omitempty"`
	Kind      []string `json:"kind,omitempty"`
	Status    []string `json:"status,omitempty"`
	Search    string   `json:"search,omitempty"`
}

type Data struct {
	List []Item `json:"list,omitempty"`
}

type Item struct {
	ID           string `json:"id,omitempty"`
	Status       Status `json:"status,omitempty"`
	Name         Link   `json:"name,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Age          string `json:"age,omitempty"`
	Ready        string `json:"ready,omitempty"`
	UpToDate     string `json:"upToDate,omitempty"`
	Available    string `json:"available,omitempty"`
	Desired      string `json:"desired,omitempty"`
	Current      string `json:"current,omitempty"`
	Completions  string `json:"completions,omitempty"`
	Duration     string `json:"duration,omitempty"`
	Schedule     string `json:"schedule,omitempty"`
	LastSchedule string `json:"lastSchedule,omitempty"`
}

type Status struct {
	RenderType  string      `json:"renderType,omitempty"`
	Value       string      `json:"value,omitempty"`
	StyleConfig StyleConfig `json:"styleConfig,omitempty"`
}

type StyleConfig struct {
	Color string `json:"color,omitempty"`
}

type Link struct {
	RenderType string                 `json:"renderType,omitempty"`
	Value      string                 `json:"value,omitempty"`
	Operations map[string]interface{} `json:"operations,omitempty"`
}

type LinkOperation struct {
	Reload bool   `json:"reload"`
	Key    string `json:"key,omitempty"`
}

type CommandState struct {
	Params map[string]string `json:"params,omitempty"`
	Query  map[string]string `json:"query,omitempty"`
}

type Props struct {
	RequestIgnore   []string `json:"requestIgnore,omitempty"`
	PageSizeOptions []string `json:"pageSizeOptions,omitempty"`
	Columns         []Column `json:"columns,omitempty"`
	RowKey          string   `json:"rowKey,omitempty"`
	SortDirections  []string `json:"sortDirections,omitempty"`
}

type Column struct {
	DataIndex string `json:"dataIndex,omitempty"`
	Title     string `json:"title,omitempty"`
	Width     int    `json:"width"`
	Sorter    bool   `json:"sorter,omitempty"`
}

type Operation struct {
	Key    string `json:"key,omitempty"`
	Reload bool   `json:"reload"`
}
