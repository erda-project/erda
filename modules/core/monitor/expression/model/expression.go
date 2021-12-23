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
	Id         string          `json:"id"`
	Expression jsonmap.JSONMap `json:"expression"`
	Attributes jsonmap.JSONMap `json:"attributes"`
}

type Template struct {
	Name       string                 `json:"name"`
	AlertType  string                 `json:"alert_type"`
	AlertIndex string                 `json:"alert_index"`
	Target     string                 `json:"target"`
	Trigger    string                 `json:"trigger"`
	Title      string                 `json:"title"`
	Template   string                 `json:"template"`
	Formats    map[string]interface{} `json:"formats"`
	Version    string                 `json:"version"`
}

type ExpressionConfig struct {
	Id         string                 `json:"id" yaml:"id"`
	Name       string                 `json:"name" yaml:"name"`
	AlertScope string                 `json:"alert_scope" yaml:"alert_scope"`
	Attributes map[string]interface{} `json:"attributes" yaml:"attributes"`
}

type Attribute struct {
	Level            string `json:"level" yaml:"level"`
	Recover          bool   `json:"recover" yaml:"recover"`
	AlertGroup       string `json:"alert_group" yaml:"alert_group"`
	DisplayUrlId     string `json:"display_url_id" yaml:"display_url_id"`
	TicketsMetricKey string `json:"tickets_metric_key" yaml:"tickets_metric_key"`
}
