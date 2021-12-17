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

package model

import "github.com/erda-project/erda/pkg/encoding/jsonmap"

var WindowKeys = []int64{1, 3, 5, 10, 15, 30}

type Expression struct {
	Name       string          `json:"name" yaml:"name"`
	AlertScope string          `json:"alert_scope" yaml:"alert_scope"`
	AlertType  string          `json:"alert_type" yaml:"alert_type"`
	AlertIndex string          `json:"alert_index"`
	Template   jsonmap.JSONMap `json:"template" yaml:"template"`
	Attributes jsonmap.JSONMap `json:"attributes" yaml:"attributes"`
}
