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

package service_provider

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/sqlutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.ServiceProvider) (*pb.ServiceProvider, error) {
	c := &ServiceProvider{
		Name:           req.Name,
		Desc:           req.Desc,
		Type:           req.Type,
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

func (dbClient *DBClient) Get(ctx context.Context, req *pb.ServiceProviderGetRequest) (*pb.ServiceProvider, error) {
	c := &ServiceProvider{BaseModel: common.BaseModelWithID(req.Id), ClientID: req.ClientId}
	whereC := &ServiceProvider{ClientID: req.ClientId}
	if err := dbClient.DB.WithContext(ctx).Model(c).First(c, whereC).Error; err != nil {
		return nil, err
	}
	return c.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.ServiceProviderDeleteRequest) (*commonpb.VoidResponse, error) {
	if req.Id == "" {
		return nil, gorm.ErrPrimaryKeyRequired
	}
	c := &ServiceProvider{BaseModel: common.BaseModelWithID(req.Id)}
	whereC := &ServiceProvider{ClientID: req.ClientId}
	sql := dbClient.DB.WithContext(ctx).Model(c).Delete(c, whereC)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.ServiceProvider) (*pb.ServiceProvider, error) {
	if req.Id == "" {
		return nil, gorm.ErrPrimaryKeyRequired
	}
	c := &ServiceProvider{
		BaseModel:      common.BaseModelWithID(req.Id),
		Name:           req.Name,
		Desc:           req.Desc,
		APIKey:         req.ApiKey,
		TemplateParams: req.TemplateParams,
		IsEnabled:      req.IsEnabled,
		Metadata:       metadata.FromProtobuf(req.Metadata),
	}
	whereC := &ServiceProvider{
		BaseModel: common.BaseModelWithID(req.Id),
		ClientID:  req.ClientId,
	}
	if err := dbClient.DB.WithContext(ctx).Model(c).Where(whereC).Updates(c).Error; err != nil {
		return nil, err
	}
	return dbClient.Get(ctx, &pb.ServiceProviderGetRequest{Id: req.Id})
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.ServiceProviderPagingRequest) (*pb.ServiceProviderPagingResponse, error) {
	c := &ServiceProvider{
		Type:       req.Type,
		ClientID:   req.ClientId,
		TemplateID: req.TemplateId,
		IsEnabled:  req.IsEnabled,
	}
	sql := dbClient.DB.WithContext(ctx).Model(c).Where(c)
	if req.Name != "" {
		sql = sql.Where("name LIKE ?", "%"+req.Name+"%")
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
		list  ServiceProviders
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
	return &pb.ServiceProviderPagingResponse{
		Total: total,
		List:  list.ToProtobuf(),
	}, nil
}
