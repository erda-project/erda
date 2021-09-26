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

package infoMapTable

import (
	"context"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type InfoMapTable struct {
	*cptype.SDK
	Ctx context.Context
	base.DefaultProvider
	Type  string            `json:"type"`
	Props Props             `json:"props"`
	Data  map[string][]Pair `json:"data"`
}

type Pair struct {
	Id    string `json:"id"`
	Label Label  `json:"label"`
	Value string `json:"value"`
}
type Label struct {
	Value       string      `json:"value"`
	RenderType  string      `json:"renderType"`
	StyleConfig StyleConfig `json:"styleConfig"`
}

type StyleConfig struct {
	FontWeight string `json:"fontWeight"`
}

type Props struct {
	IsLoadMore bool     `json:"isLoadMore,omitempty"`
	RowKey     string   `json:"rowKey"`
	Bordered   bool     `json:"bordered"`
	ShowHeader bool     `json:"showHeader"`
	Pagination bool     `json:"pagination"`
	Columns    []Column `json:"columns"`
}

type Column struct {
	DataIndex string `json:"dataIndex"`
	Title     string `json:"title"`
	Width     int    `json:"width"`
}
