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
	"github.com/erda-project/erda/internal/apps/ai-proxy/models"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/model_type"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ModelCreateRequest) (*pb.Model, error) {
	c := &models.AIProxyModel{
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       string(model_type.GetModelTypeFromProtobuf(req.Type)),
		ProviderID: req.ProviderId,
		APIKey:     req.ApiKey,
		Metadata:   models.MetadataFromProtobuf(req.Metadata),
	}
	if err := c.Creator(dbClient.DB).Create(); err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	var c models.AIProxyModel
	ok, err := (&c).Retriever(dbClient.DB).Where(models.FieldID.Equal(req.GetId())).Get()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ModelUpdateRequest) (*pb.Model, error) {
	var c = new(models.AIProxyModel)
	if _, err := c.Updater(dbClient.DB).
		Where(models.FieldID.Equal(req.GetId())).
		Updates(
			c.FieldName().Set(req.GetName()),
			c.FieldDesc().Set(req.GetDesc()),
			c.FieldType().Set(model_type.GetModelTypeFromProtobuf(req.GetType())),
			c.FieldProviderID().Set(req.GetProviderId()),
			c.FieldAPIKey().Set(req.GetApiKey()),
			c.FieldMetadata().Set(models.MetadataFromProtobuf(req.GetMetadata())),
		); err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	affects, err := (new(models.AIProxyModel)).Deleter(dbClient.DB).Where(models.FieldID.Equal(req.GetId())).Delete()
	if err != nil {
		return nil, err
	}
	if affects < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	var list models.AIProxyModelList
	var wheres = []models.Where{
		list.FieldProviderID().Equal(req.GetProviderId()),
	}
	if req.GetName() != "" {
		wheres = append(wheres, list.FieldName().Like("%"+req.Name+"%"))
	}
	if req.GetType() != pb.ModelType_TYPE_UNSPECIFIED {
		wheres = append(wheres, list.FieldType().Equal(model_type.GetModelTypeFromProtobuf(req.GetType())))
	}
	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	total, err := (&list).Pager(dbClient.DB).Where(wheres...).Paging(int(req.GetPageSize()), int(req.GetPageNum()))
	if err != nil {
		return nil, err
	}
	return &pb.ModelPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
