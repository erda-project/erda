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

package db

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

type PipelineDefinition struct {
	ID              uint64                    `json:"id" xorm:"pk autoincr"`
	PipelineSource  apistructs.PipelineSource `json:"pipelineSource"`
	PipelineYmlName string                    `json:"pipelineYmlName"`
	PipelineYml     string                    `json:"pipelineYml"`
	Extra           PipelineDefinitionExtra   `json:"extra" xorm:"json"`
	VersionLock     uint64                    `json:"versionLock" xorm:"version_lock version"`

	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created_at created"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated_at updated"`
}

type PipelineDefinitionExtra struct {
	SnippetConfig *apistructs.SnippetConfigOrder      `json:"snippetConfig" xorm:"json:"`
	CreateRequest *apistructs.PipelineCreateRequestV2 `json:"createRequest" xorm:"json:"`
}

func (PipelineDefinition) TableName() string {
	return "pipeline_definitions"
}
