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
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
)

type PipelineDefinition struct {
	ID                        string    `json:"id" xorm:"pk"`
	Name                      string    `json:"name"`
	CostTime                  uint64    `json:"costTime"`
	Creator                   string    `json:"creator"`
	Executor                  string    `json:"executor"`
	SoftDeletedAt             uint64    `json:"softDeletedAt"`
	PipelineSourceId          string    `json:"pipelineSourceId"`
	PipelineDefinitionExtraId string    `json:"pipelineDefinitionExtraId"`
	Category                  string    `json:"category"`
	StartedAt                 time.Time `json:"startedAt,omitempty" xorm:"started_at"`
	EndedAt                   time.Time `json:"endedAt,omitempty" xorm:"ended_at"`
	TimeCreated               time.Time `json:"timeCreated,omitempty" xorm:"created_at created"`
	TimeUpdated               time.Time `json:"timeUpdated,omitempty" xorm:"updated_at updated"`
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

type PipelineDefinitionSource struct {
	PipelineDefinition `xorm:"extends"`

	SourceType string `json:"sourceType"`
	Remote     string `json:"remote"`
	Ref        string `json:"ref"`
	Path       string `json:"path"`
	FileName   string `json:"fileName"`
}

func (client *Client) ListPipelineDefinition(req *pb.PipelineDefinitionListRequest, ops ...mysqlxorm.SessionOption) ([]PipelineDefinitionSource, int64, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var (
		pipelineDefinitionSources []PipelineDefinitionSource
		//	allPipelineDefinitionSources []PipelineDefinitionSource
		err error
	)
	engine := session.Table("pipeline_definitions").Alias("d").
		Select("d.*,s.source_type,s.remote,s.ref,s.path,s.name AS file_name").
		Join("LEFT", []string{"pipeline_sources", "s"}, "d.pipeline_source_id = s.id AND s.soft_deleted_at = 0").
		Where("d.soft_deleted_at = 0").
		In("s.remote", req.Remote)
	if req.Name != "" {
		engine = engine.Where("d.name LIKE ?", "%"+req.Name+"%")
	}
	if len(req.IdList) != 0 {
		engine = engine.In("d.id", req.IdList)
	}
	if len(req.Creator) != 0 {
		engine = engine.In("d.creator", req.Creator)
	}
	if len(req.Executor) != 0 {
		engine = engine.In("d.executor", req.Executor)
	}
	if len(req.Category) != 0 {
		engine = engine.In("d.category", req.Category)
	}
	if len(req.Ref) != 0 {
		engine = engine.In("s.ref", req.Ref)
	}
	if len(req.Status) != 0 {
		engine = engine.In("d.status", req.Status)
	}
	if len(req.TimeCreated) == 2 {
		if req.TimeCreated[0] != "" {
			engine = engine.Where("d.created_at >= ?", req.TimeCreated[0])
		}
		if req.TimeCreated[1] != "" {
			engine = engine.Where("d.created_at <= ?", req.TimeCreated[1])
		}
	}
	if len(req.TimeStarted) == 2 {
		if req.TimeStarted[0] != "" {
			engine = engine.Where("d.started_at >= ?", req.TimeStarted[0])
		}
		if req.TimeStarted[1] != "" {
			engine = engine.Where("d.started_at <= ?", req.TimeStarted[1])
		}
	}

	if err = engine.Limit(int(req.PageSize), int((req.PageNo-1)*req.PageSize)).
		Find(&pipelineDefinitionSources); err != nil {
		return nil, 0, err
	}

	total, err := client.CountPipelineDefinition(req, ops...)
	if err != nil {
		return nil, 0, err
	}
	return pipelineDefinitionSources, total, nil
}

func (client *Client) CountPipelineDefinition(req *pb.PipelineDefinitionListRequest, ops ...mysqlxorm.SessionOption) (int64, error) {
	session := client.NewSession(ops...)
	defer session.Close()

	var (
		total int64
		err   error
	)
	engine := session.Table("pipeline_definitions").Alias("d").
		Select("COUNT(*)").
		Join("LEFT", []string{"pipeline_sources", "s"}, "d.pipeline_source_id = s.id AND s.soft_deleted_at = 0").
		Where("d.soft_deleted_at = 0").
		In("s.remote", req.Remote)
	if req.Name != "" {
		engine = engine.Where("d.name LIKE ?", "%"+req.Name+"%")
	}
	if len(req.Creator) != 0 {
		engine = engine.In("d.creator", req.Creator)
	}
	if len(req.Executor) != 0 {
		engine = engine.In("d.executor", req.Executor)
	}
	if len(req.Category) != 0 {
		engine = engine.In("d.category", req.Category)
	}
	if len(req.Ref) != 0 {
		engine = engine.In("s.ref", req.Ref)
	}
	if len(req.Status) != 0 {
		engine = engine.In("d.status", req.Status)
	}
	if len(req.TimeCreated) == 2 {
		if req.TimeCreated[0] != "" {
			engine = engine.Where("d.created_at >= ?", req.TimeCreated[0])
		}
		if req.TimeCreated[1] != "" {
			engine = engine.Where("d.created_at <= ?", req.TimeCreated[1])
		}
	}
	if len(req.TimeStarted) == 2 {
		if req.TimeStarted[0] != "" {
			engine = engine.Where("d.started_at >= ?", req.TimeStarted[0])
		}
		if req.TimeStarted[1] != "" {
			engine = engine.Where("d.started_at <= ?", req.TimeStarted[1])
		}
	}

	total, err = engine.Count(new(PipelineDefinitionSource))
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (p *PipelineDefinitionSource) Convert() *pb.PipelineDefinition {
	timestamp := timestamppb.New(p.StartedAt.In(time.Local))
	fmt.Println(timestamp.AsTime())
	return &pb.PipelineDefinition{
		ID:          p.ID,
		Name:        p.Name,
		Creator:     p.Creator,
		Category:    p.Category,
		CostTime:    p.CostTime,
		Executor:    p.Executor,
		StartedAt:   timestamppb.New(p.StartedAt),
		EndedAt:     timestamppb.New(p.EndedAt),
		TimeCreated: timestamppb.New(p.TimeCreated),
		TimeUpdated: timestamppb.New(p.TimeUpdated),
		SourceType:  p.SourceType,
		Remote:      p.Remote,
		Ref:         p.Ref,
		Path:        p.Path,
		FileName:    p.FileName,
	}
}
