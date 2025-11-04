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

func (dbClient *DBClient) Create(ctx context.Context, req *pb.Model) (*pb.Model, error) {
	c := &Model{
		Name:           req.Name,
		Desc:           req.Desc,
		Type:           model_type.GetModelTypeFromProtobuf(req.Type),
		Publisher:      req.Publisher,
		ProviderID:     req.ProviderId,
		APIKey:         req.ApiKey,
		ClientID:       req.ClientId,
		TemplateID:     req.TemplateId,
		TemplateParams: req.TemplateParams,
		IsEnabled:      req.IsEnabled,
		Metadata:       metadata.FromProtobuf(req.Metadata),
	}
	if err := dbClient.DB.WithContext(ctx).Model(c).Create(c).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ModelGetRequest) (*pb.Model, error) {
	c := &Model{BaseModel: common.BaseModelWithID(req.Id)}
	whereC := &Model{ClientID: req.ClientId}
	if err := dbClient.DB.WithContext(ctx).Model(c).First(c, whereC).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.Model) (*pb.Model, error) {
	if req.Id == "" {
		return nil, gorm.ErrPrimaryKeyRequired
	}
	c := &Model{
		BaseModel:      common.BaseModelWithID(req.Id),
		Name:           req.Name,
		Desc:           req.Desc,
		APIKey:         req.ApiKey,
		TemplateParams: req.TemplateParams,
		IsEnabled:      req.IsEnabled,
		Metadata:       metadata.FromProtobuf(req.Metadata),
	}
	whereC := &Model{
		BaseModel: common.BaseModelWithID(req.Id),
		ClientID:  req.ClientId,
	}
	if err := dbClient.DB.WithContext(ctx).Model(c).Where(whereC).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.ModelGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ModelDeleteRequest) (*commonpb.VoidResponse, error) {
	if req.Id == "" {
		return nil, gorm.ErrPrimaryKeyRequired
	}
	c := &Model{BaseModel: common.BaseModelWithID(req.Id)}
	whereC := &Model{ClientID: req.ClientId}
	sql := dbClient.DB.Model(c).Delete(c, whereC)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ModelPagingRequest) (*pb.ModelPagingResponse, error) {
	c := &Model{
		ProviderID: req.ProviderId,
		ClientID:   req.ClientId,
		TemplateID: req.TemplateId,
		IsEnabled:  req.IsEnabled,
	}
	sql := dbClient.DB.Model(c).Where(c)
	if req.Name != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.NameFull != "" {
		c.Name = req.NameFull
	}
	if req.Type != pb.ModelType_MODEL_TYPE_UNSPECIFIED {
		c.Type = model_type.GetModelTypeFromProtobuf(req.Type)
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id in (?)", req.Ids)
	}
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

func (dbClient *DBClient) UpdateModelAbilitiesInfo(ctx context.Context, req *pb.ModelAbilitiesInfoUpdateRequest) (*commonpb.VoidResponse, error) {
	m, err := dbClient.Get(ctx, &pb.ModelGetRequest{Id: req.Id, ClientId: req.ClientId})
	if err != nil {
		return nil, err
	}
	meta := metadata.FromProtobuf(m.Metadata)
	publicMeta := meta.Public
	if publicMeta == nil {
		publicMeta = make(map[string]any)
	}
	publicMeta["abilities"] = req.Abilities
	publicMeta["context"] = req.Context
	publicMeta["pricing"] = req.Pricing

	meta.Public = publicMeta

	c := &Model{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(c).UpdateColumn("metadata", meta).Error; err != nil {
		return nil, err
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) GetClientModelIDs(ctx context.Context, clientId string) ([]string, error) {
	sql := `
SELECT
    t.model_id
FROM (
    -- 客户端专属模型
    SELECT
        m.id          AS model_id
    FROM ai_proxy_model AS m
    WHERE m.client_id = ? and (m.deleted_at IS NULL or m.deleted_at <= '1970-01-01 08:00:00')

    UNION

    -- 关联关系里的模型
    SELECT
        r.model_id    AS model_id
    FROM ai_proxy_client_model_relation AS r
    WHERE r.client_id = ? and (r.deleted_at IS NULL or r.deleted_at <= '1970-01-01 08:00:00')
) AS t
`
	var ids []string
	if err := dbClient.DB.WithContext(ctx).Raw(sql,
		clientId,
		clientId,
	).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
