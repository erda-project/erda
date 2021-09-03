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

package definition

import (
	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/transform_type"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *definitionService) GetPipelineDefinitionVersionLock(bo *transform_type.GetPipelineDefinitionVersion) (*pb.PipelineDefinitionProcessVersionResponse, error) {

	// validate
	if err := bo.Validate(); err != nil {
		return nil, apierrors.ErrrProcessPipelineDefinition.InvalidParameter(err)
	}

	dbPipelineDefinition, err := s.dbClient.GetPipelineDefinitionByNameAndSource(bo.PipelineSource, bo.PipelineYmlName)
	if err != nil {
		return nil, err
	}

	// mean first save version should be 1
	if dbPipelineDefinition == nil {
		var resp = pb.PipelineDefinitionProcessVersionResponse{
			VersionLock: 1,
		}
		return &resp, nil
	}

	var resp = pb.PipelineDefinitionProcessVersionResponse{
		VersionLock: dbPipelineDefinition.VersionLock,
	}
	return &resp, nil
}
