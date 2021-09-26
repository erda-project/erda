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

package PodDistribution

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodDistribution struct {
	base.DefaultProvider

	Props Props  `json:"props"`
	Data  Data   `json:"data"`
	Type  string `json:"type"`
}

type Props struct {
	IsLoadMore bool `json:"isLoadMore,omitempty"`
}

type Data struct {
	Total int    `json:"total"`
	Lists []List `json:"list"`
}

type List struct {
	Color string `json:"color"`
	Tip   string `json:"tip"`
	Value int    `json:"value"`
	Label string `json:"label"`
}
