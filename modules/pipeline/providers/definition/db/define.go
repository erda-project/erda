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
)

type PipelineDefinition struct {
	ID                        string     `json:"id" xorm:"pk autoincr"`
	Name                      string     `json:"name"`
	CostTime                  uint64     `json:"costTime"`
	Creator                   string     `json:"creator"`
	Executor                  string     `json:"executor"`
	SoftDeletedAt             uint64     `json:"softDeletedAt"`
	PipelineSourceId          string     `json:"pipelineSourceId"`
	PipelineDefinitionExtraId string     `json:"pipelineDefinitionExtraId"`
	Category                  string     `json:"category"`
	StartedAt                 *time.Time `json:"startedAt,omitempty" xorm:"started_at"`
	EndedAt                   *time.Time `json:"endedAt,omitempty" xorm:"ended_at"`
	TimeCreated               *time.Time `json:"timeCreated,omitempty" xorm:"created_at created"`
	TimeUpdated               *time.Time `json:"timeUpdated,omitempty" xorm:"updated_at updated"`
}

func (PipelineDefinition) TableName() string {
	return "pipeline_definitions"
}

func (client *Client) CreatePipelineDefinition(pipelineDefinition *PipelineDefinition, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err = session.InsertOne(pipelineDefinition)
	return err
}

func (client *Client) UpdatePipelineDefinition(id string, pipelineDefinition *PipelineDefinition, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineDefinition)
	return err
}

func (client *Client) DeletePipelineDefinition(id string, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.Table(new(PipelineDefinition)).ID(id).Update(map[string]interface{}{"soft_deleted_at": time.Now().UnixNano() / 1e6})
	return err
}

func (client *Client) GetPipelineDefinition(id string, ops ...mysqlxorm.SessionOption) (*PipelineDefinition, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineDefinition PipelineDefinition
	var has bool
	var err error
	if has, _, err = session.Where("id = ? and soft_deleted_at = 0", id).GetFirst(&pipelineDefinition).GetResult(); err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &pipelineDefinition, nil
}
