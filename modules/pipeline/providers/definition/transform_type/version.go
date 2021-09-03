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

package transform_type

import (
	"fmt"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
)

type GetPipelineDefinitionVersion struct {
	PipelineSource  apistructs.PipelineSource
	PipelineYmlName string
}

type GetPipelineDefinitionVersionResult struct {
	VersionLock uint64
}

func (req *GetPipelineDefinitionVersion) Validate() error {
	if req == nil {
		return fmt.Errorf("request was empty")
	}

	if !req.PipelineSource.Valid() {
		return fmt.Errorf("invalid pipelineSource: %s", req.PipelineSource)
	}

	if req.PipelineYmlName == "" {
		return fmt.Errorf("invalid pipelineYmlName: %s", req.PipelineYmlName)
	}

	return nil
}

func (bo *GetPipelineDefinitionVersionResult) TransformToResp() (*pb.PipelineDefinitionProcessVersionResponse, error) {
	if bo == nil {
		return nil, nil
	}

	var resp = pb.PipelineDefinitionProcessVersionResponse{
		VersionLock: bo.VersionLock,
	}

	return &resp, nil
}

func (bo *GetPipelineDefinitionVersion) ReqTransform(req *pb.PipelineDefinitionProcessVersionRequest) {
	if req == nil {
		return
	}
	if bo == nil {
		return
	}

	bo.PipelineSource = apistructs.PipelineSource(req.PipelineSource)
	bo.PipelineYmlName = req.PipelineYmlName
}
