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
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
)

type PipelineDefinitionProcess struct {
	PipelineSource        apistructs.PipelineSource           `json:"pipelineSource"`
	PipelineYmlName       string                              `json:"pipelineYmlName"`
	PipelineYml           string                              `json:"pipelineYml"`
	SnippetConfig         *apistructs.SnippetConfig           `json:"snippetConfig"`
	VersionLock           uint64                              `json:"versionLock"`
	IsDelete              bool                                `json:"isDelete"`
	PipelineCreateRequest *apistructs.PipelineCreateRequestV2 `json:"pipelineCreateRequest"`
}

type PipelineDefinitionProcessResult struct {
	ID              uint64
	PipelineSource  apistructs.PipelineSource
	PipelineYmlName string
	PipelineYml     string
	VersionLock     uint64
}

func (req *PipelineDefinitionProcess) Validate() error {
	if req == nil {
		return fmt.Errorf("request was empty")
	}

	if !req.PipelineSource.Valid() {
		return fmt.Errorf("invalid pipelineSource: %s", req.PipelineSource)
	}

	if req.PipelineYmlName == "" {
		return fmt.Errorf("invalid pipelineYmlName: %s", req.PipelineYmlName)
	}

	if req.PipelineYml == "" && req.IsDelete == false {
		return fmt.Errorf("invalid pipelineYml: %s", req.PipelineYml)
	}

	if req.VersionLock <= 0 {
		return fmt.Errorf("invalid versionLock: %v, get versionLock before save definition", req.VersionLock)
	}

	return nil
}

func (bo *PipelineDefinitionProcessResult) TransformToResp() (*pb.PipelineDefinitionProcessResponse, error) {
	if bo == nil {
		return nil, nil
	}

	var resp = pb.PipelineDefinitionProcessResponse{
		Id:              bo.ID,
		PipelineYml:     bo.PipelineYml,
		PipelineYmlName: bo.PipelineYmlName,
		PipelineSource:  bo.PipelineSource.String(),
		VersionLock:     bo.VersionLock,
	}

	return &resp, nil
}

func (bo *PipelineDefinitionProcess) ReqTransform(req *pb.PipelineDefinitionProcessRequest) error {
	if req == nil {
		return nil
	}
	if bo == nil {
		return nil
	}

	bo.PipelineSource = apistructs.PipelineSource(req.PipelineSource)
	bo.PipelineYmlName = req.PipelineYmlName
	bo.PipelineYml = req.PipelineYml
	bo.VersionLock = req.VersionLock
	bo.IsDelete = req.IsDelete

	if req.SnippetConfig != nil {
		data, err := req.SnippetConfig.MarshalJSON()
		if err != nil {
			return err
		}
		var snippetConfig apistructs.SnippetConfig
		err = json.Unmarshal(data, &snippetConfig)
		if err != nil {
			return err
		}
		bo.SnippetConfig = &snippetConfig
	}

	if req.PipelineCreateRequest != nil {
		data, err := req.PipelineCreateRequest.MarshalJSON()
		if err != nil {
			return err
		}
		var pipelineCreateRequest apistructs.PipelineCreateRequestV2
		err = json.Unmarshal(data, &pipelineCreateRequest)
		if err != nil {
			return err
		}
		bo.PipelineCreateRequest = &pipelineCreateRequest
	}

	return nil
}
