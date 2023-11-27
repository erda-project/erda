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

package model_provider

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_provider_type"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ModelProviderCreateRequest) (*pb.ModelProvider, error) {
	c := &ModelProvider{
		Name:     req.Name,
		Desc:     req.Desc,
		Type:     model_provider_type.GetModelProviderTypeFromProtobuf(req.Type),
		APIKey:   req.ApiKey,
		Metadata: metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ModelProviderGetRequest) (*pb.ModelProvider, error) {
	c := &ModelProvider{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).First(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ModelProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	c := &ModelProvider{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.Model(c).Delete(c)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ModelProviderUpdateRequest) (*pb.ModelProvider, error) {
	c := &ModelProvider{
		BaseModel: common.BaseModelWithID(req.Id),
		Name:      req.Name,
		Desc:      req.Desc,
		Type:      model_provider_type.GetModelProviderTypeFromProtobuf(req.Type),
		APIKey:    req.ApiKey,
		Metadata:  metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.Model(c).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.ModelProviderGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ModelProviderPagingRequest) (*pb.ModelProviderPagingResponse, error) {
	c := &ModelProvider{}
	sql := dbClient.DB.Model(c)
	if req.Name != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Type != pb.ModelProviderType_MODEL_PROVIDER_TYPE_UNSPECIFIED {
		c.Type = model_provider_type.GetModelProviderTypeFromProtobuf(req.Type)
		sql = sql.Where("type = ?", c.Type)
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id in (?)", req.Ids)
	}
	var (
		total int64
		list  ModelProviders
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
		return nil, err
	}
	return &pb.ModelProviderPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
