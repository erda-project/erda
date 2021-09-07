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

package deftype

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

type ClientDefinitionProcessRequest struct {
	PipelineSource        apistructs.PipelineSource           `json:"pipelineSource"`
	PipelineYmlName       string                              `json:"pipelineYmlName"`
	PipelineYml           string                              `json:"pipelineYml"`
	SnippetConfig         *apistructs.SnippetConfig           `json:"snippetConfig"`
	VersionLock           uint64                              `json:"versionLock"`
	IsDelete              bool                                `json:"isDelete"`
	PipelineCreateRequest *apistructs.PipelineCreateRequestV2 `json:"pipelineCreateRequest"`
}

type ClientDefinitionProcessResponse struct {
	ID              uint64                    `json:"id"`
	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`
	PipelineYml     string                    `json:"pipelineYml"`
	VersionLock     uint64                    `json:"versionLock"`
	TimeCreated     *time.Time                `json:"timeCreated,omitempty"`
	TimeUpdated     *time.Time                `json:"timeUpdated,omitempty"`
}
