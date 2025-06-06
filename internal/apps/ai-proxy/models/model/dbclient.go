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

package model

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_type"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/sqlutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	c := &Model{
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       model_type.GetModelTypeFromProtobuf(req.Type),
		ProviderID: req.ProviderId,
		APIKey:     req.ApiKey,
		Metadata:   metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	c := &Model{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ModelUpdateRequest) (*pb.Model, error) {
	c := &Model{
		BaseModel:  common.BaseModelWithID(req.Id),
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       model_type.GetModelTypeFromProtobuf(req.Type),
		ProviderID: req.ProviderId,
		APIKey:     req.ApiKey,
		Metadata:   metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.ModelGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	c := &Model{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	c := &Model{}
	sql := dbClient.DB.Model(c)
	if req.Name != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Type != pb.ModelType_MODEL_TYPE_UNSPECIFIED {
		c.Type = model_type.GetModelTypeFromProtobuf(req.Type)
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id in (?)", req.Ids)
	}
	c.ProviderID = req.ProviderId
	sql = sql.Where(c)
	// order by
	sql, err := sqlutil.HandleOrderBy(sql, req.OrderBys)
	if err != nil {
		return nil, err
	}
	var (
		total int64
		list  Models
	)
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	offset := (req.PageNum - 1) * req.PageSize
	err = sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return &pb.ModelPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
