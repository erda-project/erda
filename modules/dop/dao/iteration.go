// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package dao

import (
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type Iteration struct {
	dbengine.BaseModel

	StartedAt  *time.Time                // 迭代开始时间
	FinishedAt *time.Time                // 迭代结束时间
	ProjectID  uint64                    // 所属项目 ID
	Title      string                    // 标题
	Content    string                    // 内容
	Creator    string                    // 创建者 ID
	State      apistructs.IterationState // 归档状态
}

func (Iteration) TableName() string {
	return "dice_iterations"
}

func (i *Iteration) Convert() apistructs.Iteration {
	return apistructs.Iteration{
		ID:         int64(i.ID),
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
		StartedAt:  i.StartedAt,
		FinishedAt: i.FinishedAt,
		ProjectID:  i.ProjectID,
		Title:      i.Title,
		Content:    i.Content,
		Creator:    i.Creator,
		State:      i.State,
	}
}

// CreateIteration 创建
func (client *DBClient) CreateIteration(Iteration *Iteration) error {
	return client.Create(Iteration).Error
}

// UpdateIteration 更新
func (client *DBClient) UpdateIteration(Iteration *Iteration) error {
	return client.Save(Iteration).Error
}

// GetIteration Iteration 详情
func (client *DBClient) GetIteration(id uint64) (*Iteration, error) {
	var iteration Iteration
	if err := client.Where("id = ?", id).Find(&iteration).Error; err != nil {
		return nil, err
	}
	return &iteration, nil
}

// GetIterationByTitle 根据 title 获取 Iteration 信息
func (client *DBClient) GetIterationByTitle(projectID uint64, title string) (*Iteration, error) {
	var iteration Iteration
	if err := client.Where("project_id = ?", projectID).Where("title = ?", title).First(&iteration).Error; err != nil {
		return nil, err
	}
	return &iteration, nil
}

// DeleteIteration 删除
func (client *DBClient) DeleteIteration(id uint64) error {
	Iteration := Iteration{BaseModel: dbengine.BaseModel{ID: id}}
	return client.Delete(&Iteration).Error
}

// PagingIterations 分页查询
func (client *DBClient) PagingIterations(req apistructs.IterationPagingRequest) ([]Iteration, uint64, error) {
	var (
		total      uint64
		iterations []Iteration
	)
	cond := Iteration{}
	if req.ProjectID > 0 {
		cond.ProjectID = req.ProjectID
	}
	if req.State != "" {
		cond.State = req.State
	}
	sql := client.Where(cond)
	if req.Deadline != "" {
		sql = sql.Where("finished_at > ?", req.Deadline)
	}

	// result
	offset := (req.PageNo - 1) * req.PageSize
	if err := sql.Order("id DESC").
		Offset(offset).Limit(req.PageSize).Find(&iterations).
		// reset offset & limit before count
		Offset(0).Limit(-1).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return iterations, total, nil
}

func (client *DBClient) FindIterations(projectID uint64) (iterations []Iteration, err error) {
	if err := client.Where("project_id = ?", projectID).Find(&iterations).Error; err != nil {
		return nil, err
	}
	return iterations, nil
}
