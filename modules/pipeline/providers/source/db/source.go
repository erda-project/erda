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
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"time"
)

type PipelineSource struct {
	ID          string `json:"id" xorm:"pk autoincr"`
	SourceType  string `json:"sourceType"`
	Remote      string `json:"remote"`
	Ref         string `json:"ref"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	PipelineYml string `json:"pipelineYml"`

	VersionLock   uint64     `json:"versionLock" xorm:"version_lock version"`
	SoftDeletedAt uint64     `json:"softDeletedAt"`
	TimeCreated   *time.Time `json:"timeCreated,omitempty" xorm:"created_at created"`
	TimeUpdated   *time.Time `json:"timeUpdated,omitempty" xorm:"updated_at updated"`
}

func (PipelineSource) TableName() string {
	return "pipeline_sources"
}

func (client *Client) CreatePipelineSource(pipelineSource *PipelineSource, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err = session.InsertOne(pipelineSource)
	return err
}

func (client *Client) UpdatePipelineSource(id uint64, pipelineSource *PipelineSource, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineSource)
	return err
}

func (client *Client) DeletePipelineSourceExtra(id uint64, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).Delete(new(PipelineSource))
	return err
}

func (client *Client) GetPipelineSource(id uint64, ops ...mysqlxorm.SessionOption) (*PipelineSource, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var pipelineSource PipelineSource
	var has bool
	var err error
	if has, _, err = session.Where("id = ?", id).GetFirst(&pipelineSource).GetResult(); err != nil {
		return nil, err
	}

	if !has {
		return nil, nil
	}

	return &pipelineSource, nil
}

