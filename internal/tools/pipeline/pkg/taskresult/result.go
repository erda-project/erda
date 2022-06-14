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

package taskresult

import (
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskerror"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskinspect"
	"github.com/erda-project/erda/pkg/metadata"
)

// PipelineTaskResult spec.pipeline task only use metadata, task dto has all fields
type PipelineTaskResult struct {
	Metadata    metadata.Metadata                    `json:"metadata,omitempty"`
	Errors      []*taskerror.PipelineTaskErrResponse `json:"errors,omitempty"`
	MachineStat *taskinspect.PipelineTaskMachineStat `json:"machineStat,omitempty"`
	Inspect     string                               `json:"inspect,omitempty"`
	Events      string                               `json:"events,omitempty"`
}
