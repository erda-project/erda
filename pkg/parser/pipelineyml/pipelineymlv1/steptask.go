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

package pipelineymlv1

import (
	"time"

	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1/steptasktype"
)

type StepTask interface {
	Type() steptasktype.StepTaskType
	Name() string
	GetEnvs() map[string]string
	Validate() error
	OutputToContext() []string
	// includes required resources and explicit context
	RequiredContext(y PipelineYml) []string

	IsResourceTask() bool
	// if IsResourceTask, GetResourceType will return resource type name.
	GetResourceType() string

	IsDisable() bool
	IsPause() bool

	GetStatus() string
	SetStatus(status string)

	GetTaskParams() Params
	GetTaskConfig() TaskConfig
	GetCustomTaskRunPathArgs() []string
	GetResource() *Resource
	GetVersion() Version

	GetTimeout() (time.Duration, error)
}
