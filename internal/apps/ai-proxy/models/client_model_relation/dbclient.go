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

package client_model_relation

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/client_model_relation/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/client"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model"
)

type DBClient struct {
	DB           *gorm.DB
	ClientClient *client.DBClient
}

func (dbClient *DBClient) Allocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if err := TxCheckClientID(tx, req.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}
	// check model ids
	if err := TxCheckModelIDs(tx, req.ModelIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	// do allocate
	for _, modelId := range req.ModelIds {
		c := &ClientModelRelation{
			ClientID: req.ClientId,
			ModelID:  modelId,
		}
		if err := tx.Model(c).Create(c).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to allocate model %s to client %s: %v", modelId, req.ClientId, err)
		}
	}
	tx.Commit()
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) UnAllocate(ctx context.Context, req *pb.AllocateRequest) (*commonpb.VoidResponse, error) {
	tx := dbClient.DB.Begin()
	// check client id
	if err := TxCheckClientID(tx, req.ClientId); err != nil {
		tx.Rollback()
		return nil, err
	}
	// check model ids
	if err := TxCheckModelIDs(tx, req.ModelIds); err != nil {
		tx.Rollback()
		return nil, err
	}
	// do unallocate
	if err := tx.Model(&ClientModelRelation{ClientID: req.ClientId}).
		Where("model_id in (?)", req.ModelIds).
		Delete(nil).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to unallocate models %v from client %s: %v", req.ModelIds, req.ClientId, err)
	}
	tx.Commit()
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) ListClientModels(ctx context.Context, req *pb.ListClientModelsRequest) (*pb.ListAllocatedModelsResponse, error) {
	modelTableName, relationTableName := (&model.Model{}).TableName(), (&ClientModelRelation{}).TableName()
	var relations []ClientModelRelation
	sql := dbClient.DB.Table(relationTableName).
		Joins(fmt.Sprintf(`left join %s on %s.id = %s.model_id`, modelTableName, modelTableName, relationTableName)).
		Where(fmt.Sprintf(`%s.client_id = ?`, relationTableName), req.ClientId)
	if len(req.ModelTypes) > 0 {
		sql = sql.Where(fmt.Sprintf(`%s.type in (?)`, modelTableName), req.ModelTypes)
	}
	sql.Find(&relations)
	if sql.Error != nil {
		return nil, fmt.Errorf("failed to list client models: %v", sql.Error)
	}
	var modelIds []string
	for _, rel := range relations {
		modelIds = append(modelIds, rel.ModelID)
	}
	return &pb.ListAllocatedModelsResponse{
		ClientId: req.ClientId,
		ModelIds: modelIds,
	}, nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.PagingRequest) (*pb.PagingResponse, error) {
	relationTableName := (&ClientModelRelation{}).TableName()
	sql := dbClient.DB.Table(relationTableName)
	if len(req.ClientIds) > 0 {
		sql = sql.Where("client_id in (?)", req.ClientIds)
	}
	if len(req.ModelIds) > 0 {
		sql = sql.Where("model_id in (?)", req.ModelIds)
	}
	// order by
	if len(req.OrderBy) == 0 {
		sql = sql.Order("updated_at desc")
	} else {
		for _, orderBy := range req.OrderBy {
			// get is desc or asc
			parts := strings.Split(orderBy, " ")
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid order by: %s", orderBy)
			}
			sql = sql.Order(clause.OrderByColumn{
				Column: clause.Column{Name: parts[0], Raw: false},
				Desc:   strings.EqualFold(parts[1], "desc"),
			})
		}
	}
	var (
		total int64
		list  ClientModelRelations
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	err := sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, fmt.Errorf("failed to paging client model relations: %v", err)
	}
	return &pb.PagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}

func TxCheckClientID(tx *gorm.DB, clientID string) error {
	if clientID == "" {
		return fmt.Errorf("client id is empty")
	}
	var clientCount int64
	c := &client.Client{BaseModel: common.BaseModelWithID(clientID)}
	if err := tx.Model(c).Where(c).Count(&clientCount).Error; err != nil {
		return fmt.Errorf("failed to check client id: %v", err)
	}
	if clientCount == 0 {
		return fmt.Errorf("client id %s not found", clientID)
	}
	return nil
}

func TxCheckModelIDs(tx *gorm.DB, modelIDs []string) error {
	if len(modelIDs) == 0 {
		return fmt.Errorf("modelIds is empty")
	}
	var existModelIds []string
	if err := tx.Model(&model.Model{}).
		Select("id").
		Where("id in (?)", modelIDs).
		Find(&existModelIds).Error; err != nil {
		return fmt.Errorf("failed to check model ids: %v", err)
	}
	if len(existModelIds) != len(modelIDs) {
		return fmt.Errorf("model ids %v not found", findMissingModelIds(modelIDs, existModelIds))
	}
	return nil
}

func findMissingModelIds(expected, actually []string) []string {
	actuallyMap := make(map[string]bool, len(actually))
	for _, id := range actually {
		actuallyMap[id] = true
	}
	var missings []string
	for _, expected := range expected {
		if _, ok := actuallyMap[expected]; !ok {
			missings = append(missings, expected)
		}
	}
	return missings
}
