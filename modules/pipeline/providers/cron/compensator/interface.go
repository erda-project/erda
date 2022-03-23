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

package compensator

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

type PipelineFunc struct {
	RunPipeline    RunPipelineFunc
	CreatePipeline CreatePipelineFunc
}

type RunPipelineFunc func(req *apistructs.PipelineRunRequest) (*spec.Pipeline, error)
type CreatePipelineFunc func(req *apistructs.PipelineCreateRequestV2) (*spec.Pipeline, error)

type Interface interface {
	CronNotExecuteCompensateById(ID uint64) error

	// todo Can be removed after all objects are provider
	WithPipelineFunc(pipelineFunc PipelineFunc)
}
