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

package pipelineyml

type Pipeline struct {
	Version string   `json:"version"`
	Stages  []*Stage `json:"stages,omitempty"`

	Envs map[string]string `json:"envs,omitempty"`

	Resources     []Resource     `json:"resources,omitempty"`
	ResourceTypes []ResourceType `json:"resource_types,omitempty"`
}

type Stage struct {
	Name        string       `json:"name,omitempty"`
	TaskConfigs []TaskConfig `json:"tasks,omitempty" ymal:"tasks" yml:"tasks" ymal.org:"tasks"`
	//PipelineTasks       []StepTask   `json:"-"`
}

type TaskConfig map[string]interface{}

type Resource struct {
	Name string `json:"name"`
	Type string `json:"type"`
	// Optional
	Source Source `json:"source"`
}

type ResourceType struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Source Source `json:"source"`
}
