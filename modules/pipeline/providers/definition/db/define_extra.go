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

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/apistructs"
)

type PipelineDefinitionExtra struct {
	ID          uint64     `json:"id" xorm:"pk autoincr"`
	Extra       Extra      `json:"extra" xorm:"json"`
	TimeCreated *time.Time `json:"timeCreated,omitempty" xorm:"created_at created"`
	TimeUpdated *time.Time `json:"timeUpdated,omitempty" xorm:"updated_at updated"`
}

type Extra struct {
	SnippetConfig *apistructs.SnippetConfigOrder      `json:"snippetConfig" xorm:"json:"`
	CreateRequest *apistructs.PipelineCreateRequestV2 `json:"createRequest" xorm:"json:"`
}

func (PipelineDefinitionExtra) TableName() string {
	return "pipeline_definition_extras"
}

func (client *Client) CreatePipelineDefinitionExtra(pipelineDefinitionExtra *PipelineDefinitionExtra, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err = session.InsertOne(pipelineDefinitionExtra)
	return err
}

func (client *Client) UpdatePipelineDefinitionExtra(id uint64, pipelineDefinitionExtra *PipelineDefinitionExtra, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineDefinitionExtra)
	return err
}

func (client *Client) DeletePipelineDefinitionExtraExtra(id uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).Delete(new(PipelineDefinitionExtra))
	return err
}

func (client *Client) GetPipelineDefinitionExtra(id uint64, ops ...mysqlxorm.SessionOption) (*PipelineDefinitionExtra, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineDefinitionExtra PipelineDefinitionExtra
	var has bool
	var err error
	if has, _, err = session.Where("id = ?", id).GetFirst(&pipelineDefinitionExtra).GetResult(); err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &pipelineDefinitionExtra, nil
}
