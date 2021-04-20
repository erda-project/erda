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

package spec

type PipelineConfig struct {
	ID    uint64             `json:"id" xorm:"pk autoincr"`
	Type  PipelineConfigType `json:"type"`
	Value interface{}        `json:"value" xorm:"json"`
}

func (PipelineConfig) TableName() string {
	return "pipeline_configs"
}

type ActionExecutorConfig struct {
	Kind    string            `json:"kind,omitempty"`
	Name    string            `json:"name,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

type PipelineConfigType string

var (
	PipelineConfigTypeActionExecutor PipelineConfigType = "action_executor"
)
