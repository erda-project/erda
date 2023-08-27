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
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	c := &Model{
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       GetModelTypeFromProtobuf(req.Type),
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
		Type:       ModelType(req.Type),
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
