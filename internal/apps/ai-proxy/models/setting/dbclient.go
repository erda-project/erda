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

package setting

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/setting/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/sqlutil"
)

type DBClient struct {
	DB *gorm.DB
}

func (dbClient *DBClient) CreateOrUpdate(ctx context.Context, item *Setting) error {
	if item == nil {
		return nil
	}

	var existing Setting
	err := dbClient.DB.WithContext(ctx).
		Where(&Setting{Namespace: item.Namespace, Key: item.Key}).
		First(&existing).Error
	if err == nil {
		existing.Value = item.Value
		return dbClient.DB.WithContext(ctx).Model(&existing).Updates(map[string]any{
			"value": item.Value,
		}).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	return dbClient.DB.WithContext(ctx).Create(item).Error
}

func (dbClient *DBClient) GetByNamespaceKey(ctx context.Context, namespace, key string) (*Setting, error) {
	var item Setting
	err := dbClient.DB.WithContext(ctx).
		Where(&Setting{Namespace: namespace, Key: key}).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (dbClient *DBClient) GetByNamespaceKeys(ctx context.Context, namespace string, keys ...string) (map[string]*Setting, error) {
	if len(keys) == 0 {
		return map[string]*Setting{}, nil
	}

	var items []*Setting
	if err := dbClient.DB.WithContext(ctx).
		Where("namespace = ? AND `key` IN (?)", namespace, keys).
		Find(&items).Error; err != nil {
		return nil, err
	}

	result := make(map[string]*Setting, len(items))
	for _, item := range items {
		result[item.Key] = item
	}
	return result, nil
}

func (dbClient *DBClient) Create(ctx context.Context, req *pb.SettingCreateRequest) (*pb.Setting, error) {
	item := &Setting{
		Namespace: req.Namespace,
		Key:       req.Key,
		Value:     req.Value,
	}
	if err := dbClient.DB.WithContext(ctx).Create(item).Error; err != nil {
		return nil, err
	}
	got, err := dbClient.GetByNamespaceKey(ctx, req.Namespace, req.Key)
	if err != nil {
		return nil, err
	}
	return got.ToProtobuf(), nil
}

func (dbClient *DBClient) Get(ctx context.Context, req *pb.SettingGetRequest) (*pb.Setting, error) {
	item := &Setting{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.WithContext(ctx).Model(item).First(item).Error; err != nil {
		return nil, err
	}
	return item.ToProtobuf(), nil
}

func (dbClient *DBClient) Delete(ctx context.Context, req *pb.SettingDeleteRequest) (*commonpb.VoidResponse, error) {
	item := &Setting{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.WithContext(ctx).Model(item).Delete(item)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

func (dbClient *DBClient) Update(ctx context.Context, req *pb.SettingUpdateRequest) (*pb.Setting, error) {
	item := &Setting{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.WithContext(ctx).Model(item).Updates(Setting{
		Namespace: req.Namespace,
		Key:       req.Key,
		Value:     req.Value,
	})
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return dbClient.Get(ctx, &pb.SettingGetRequest{Id: req.Id})
}

func (dbClient *DBClient) ListAll(ctx context.Context) ([]*Setting, error) {
	var list []*Setting
	if err := dbClient.DB.WithContext(ctx).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (dbClient *DBClient) Paging(ctx context.Context, req *pb.SettingPagingRequest) (*pb.SettingPagingResponse, error) {
	item := &Setting{}
	sql := dbClient.DB.WithContext(ctx).Model(item)

	if req.Namespace != "" {
		sql = sql.Where("namespace = ?", req.Namespace)
	}
	if req.Key != "" {
		sql = sql.Where("`key` = ?", req.Key)
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id IN (?)", req.Ids)
	}
	var err error
	sql, err = sqlutil.HandleOrderBy(sql, req.OrderBys)
	if err != nil {
		return nil, err
	}

	if req.PageNum == 0 {
		req.PageNum = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	var (
		total int64
		list  []*Setting
	)
	offset := (req.PageNum - 1) * req.PageSize
	if err := sql.Count(&total).Limit(int(req.PageSize)).Offset(int(offset)).Find(&list).Error; err != nil {
		return nil, err
	}

	pbList := make([]*pb.Setting, 0, len(list))
	for _, one := range list {
		pbList = append(pbList, one.ToProtobuf())
	}
	return &pb.SettingPagingResponse{
		Total: total,
		List:  pbList,
	}, nil
}
