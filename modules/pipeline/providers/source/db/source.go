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

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
)

type PipelineSource struct {
	ID          string `json:"id"`
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

type PipelineSourceUnique struct {
	SourceType string `json:"sourceType"`
	Remote     string `json:"remote"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	Name       string `json:"name"`
}

func (PipelineSource) TableName() string {
	return "pipeline_sources"
}

func (client *Client) CreatePipelineSource(pipelineSource *PipelineSource, ops ...mysqlxorm.SessionOption) (err error) {
	session := client.NewSession(ops...)
	defer session.Close()

	pipelineSource.ID = uuid.New().String()
	_, err = session.InsertOne(pipelineSource)
	return err
}

func (client *Client) UpdatePipelineSource(id string, pipelineSource *PipelineSource, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	_, err := session.ID(id).AllCols().Update(pipelineSource)
	return err
}

func (client *Client) DeletePipelineSource(id string, ops ...mysqlxorm.SessionOption) error {
	session := client.NewSession(ops...)
	defer session.Close()

	source := new(PipelineSource)
	source.SoftDeletedAt = uint64(time.Now().UnixNano() / 1e6)
	_, err := session.ID(id).Cols("soft_deleted_at").Delete(source)
	return err
}

func (client *Client) GetPipelineSource(id string, ops ...mysqlxorm.SessionOption) (*PipelineSource, error) {
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

func (client *Client) GetPipelineSourceByUnique(unique *PipelineSourceUnique, ops ...mysqlxorm.SessionOption) ([]PipelineSource, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var (
		pipelineSources []PipelineSource
		err             error
	)
	if err = session.
		Where("source_type = ?", unique.SourceType).
		Where("remote = ?", unique.Remote).
		Where("ref = ?", unique.Ref).
		Where("path = ?", unique.Path).
		Where("name = ?", unique.Name).
		Where("soft_deleted_at = 0").
		Find(&pipelineSources); err != nil {
		return nil, err
	}
	return pipelineSources, nil
}

func (p *PipelineSource) Convert() *pb.PipelineSource {
	return &pb.PipelineSource{
		ID:          p.ID,
		SourceType:  p.SourceType,
		Remote:      p.Remote,
		Ref:         p.Ref,
		Path:        p.Path,
		Name:        p.Name,
		PipelineYml: p.PipelineYml,
		VersionLock: p.VersionLock,
		TimeCreated: timestamppb.New(*p.TimeCreated),
		TimeUpdated: timestamppb.New(*p.TimeUpdated),
	}
}
