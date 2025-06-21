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

package i18n

import (
	"context"

	"gorm.io/gorm"

	"github.com/erda-project/erda-proto-go/apps/aiproxy/i18n/pb"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/common"
)

type DBClient struct {
	DB *gorm.DB
}

// GetConfigInternal get configuration value by category, item key, field name and locale
func (dbClient *DBClient) GetConfigInternal(ctx context.Context, category string, itemKey string, fieldName string, locale string) (*Config, error) {
	config := &Config{}
	err := dbClient.DB.Where("category = ? AND item_key = ? AND field_name = ? AND locale = ?",
		category, itemKey, fieldName, locale).First(config).Error
	if err != nil {
		return nil, err
	}
	return config, nil
}

// GetConfigsByCategory get all configurations by category
func (dbClient *DBClient) GetConfigsByCategory(ctx context.Context, category string) ([]*Config, error) {
	var configs []*Config
	err := dbClient.DB.Where("category = ?", category).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetConfigsByItemKey get all configurations by item key
func (dbClient *DBClient) GetConfigsByItemKey(ctx context.Context, category string, itemKey string) ([]*Config, error) {
	var configs []*Config
	err := dbClient.DB.Where("category = ? AND item_key = ?", category, itemKey).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetConfigsByLocale get all configurations by locale
func (dbClient *DBClient) GetConfigsByLocale(ctx context.Context, locale string) ([]*Config, error) {
	var configs []*Config
	err := dbClient.DB.Where("locale = ?", locale).Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetAllConfigs get all i18n configurations - used for cache preloading
func (dbClient *DBClient) GetAllConfigs() ([]*Config, error) {
	var configs []*Config
	err := dbClient.DB.Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// CreateOrUpdate create or update configuration
func (dbClient *DBClient) CreateOrUpdate(ctx context.Context, config *Config) error {
	return dbClient.DB.Save(config).Error
}

// DeleteInternal delete configuration
func (dbClient *DBClient) DeleteInternal(ctx context.Context, category string, itemKey string, fieldName string, locale string) error {
	return dbClient.DB.Where("category = ? AND item_key = ? AND field_name = ? AND locale = ?",
		category, itemKey, fieldName, locale).Delete(&Config{}).Error
}

// BatchGetConfigs batch get configurations - used for performance optimization
func (dbClient *DBClient) BatchGetConfigs(ctx context.Context, keys []ConfigKey) (map[string]*Config, error) {
	if len(keys) == 0 {
		return make(map[string]*Config), nil
	}

	var configs []*Config
	query := dbClient.DB.Where("1=0") // initialize with false condition

	for _, key := range keys {
		query = query.Or("(category = ? AND item_key = ? AND field_name = ? AND locale = ?)",
			key.Category, key.ItemKey, key.FieldName, key.Locale)
	}

	err := query.Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// build result mapping
	result := make(map[string]*Config)
	for _, config := range configs {
		key := BuildConfigKey(config.Category, config.ItemKey, config.FieldName, config.Locale)
		result[key] = config
	}

	return result, nil
}

// ConfigKey configuration key structure
type ConfigKey struct {
	Category  string
	ItemKey   string
	FieldName string
	Locale    string
}

// BuildConfigKey build configuration key
func BuildConfigKey(category, itemKey, fieldName, locale string) string {
	return category + ":" + itemKey + ":" + fieldName + ":" + locale
}

// Protobuf API methods

// Create create i18n configuration
func (dbClient *DBClient) Create(ctx context.Context, req *pb.I18NCreateRequest) (*pb.I18NConfig, error) {
	config := &Config{
		Category:  req.Category,
		ItemKey:   req.ItemKey,
		FieldName: req.FieldName,
		Locale:    req.Locale,
		Value:     req.Value,
	}
	if err := dbClient.DB.Model(config).Create(config).Error; err != nil {
		return nil, err
	}
	return config.ToProtobuf(), nil
}

// Get get i18n configuration by ID
func (dbClient *DBClient) Get(ctx context.Context, req *pb.I18NGetRequest) (*pb.I18NConfig, error) {
	config := &Config{BaseModel: common.BaseModelWithID(req.Id)}
	if err := dbClient.DB.Model(config).First(config).Error; err != nil {
		return nil, err
	}
	return config.ToProtobuf(), nil
}

// Delete delete i18n configuration
func (dbClient *DBClient) Delete(ctx context.Context, req *pb.I18NDeleteRequest) (*commonpb.VoidResponse, error) {
	config := &Config{BaseModel: common.BaseModelWithID(req.Id)}
	sql := dbClient.DB.Model(config).Delete(config)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected < 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return &commonpb.VoidResponse{}, nil
}

// Update update i18n configuration
func (dbClient *DBClient) Update(ctx context.Context, req *pb.I18NUpdateRequest) (*pb.I18NConfig, error) {
	config := &Config{
		BaseModel: common.BaseModelWithID(req.Id),
		Category:  req.Category,
		ItemKey:   req.ItemKey,
		FieldName: req.FieldName,
		Locale:    req.Locale,
		Value:     req.Value,
	}
	sql := dbClient.DB.Model(config).Updates(config)
	if sql.Error != nil {
		return nil, sql.Error
	}
	if sql.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return dbClient.Get(ctx, &pb.I18NGetRequest{Id: req.Id})
}

// Paging paginated query for i18n configurations
func (dbClient *DBClient) Paging(ctx context.Context, req *pb.I18NPagingRequest) (*pb.I18NPagingResponse, error) {
	config := &Config{}
	sql := dbClient.DB.Model(config)

	if req.Category != "" {
		sql = sql.Where("category = ?", req.Category)
	}
	if req.ItemKey != "" {
		sql = sql.Where("item_key = ?", req.ItemKey)
	}
	if req.FieldName != "" {
		sql = sql.Where("field_name = ?", req.FieldName)
	}
	if req.Locale != "" {
		sql = sql.Where("locale = ?", req.Locale)
	}
	if len(req.Ids) > 0 {
		sql = sql.Where("id IN (?)", req.Ids)
	}

	var (
		total int64
		list  []*Config
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

	var pbList []*pb.I18NConfig
	for _, item := range list {
		pbList = append(pbList, item.ToProtobuf())
	}

	return &pb.I18NPagingResponse{
		Total: total,
		List:  pbList,
	}, nil
}

// BatchCreate batch create i18n configurations
func (dbClient *DBClient) BatchCreate(ctx context.Context, req *pb.I18NBatchCreateRequest) (*pb.I18NBatchCreateResponse, error) {
	var configs []*Config
	for _, createReq := range req.Configs {
		config := &Config{
			Category:  createReq.Category,
			ItemKey:   createReq.ItemKey,
			FieldName: createReq.FieldName,
			Locale:    createReq.Locale,
			Value:     createReq.Value,
		}
		configs = append(configs, config)
	}

	if err := dbClient.DB.CreateInBatches(configs, len(configs)).Error; err != nil {
		return nil, err
	}

	var pbConfigs []*pb.I18NConfig
	for _, config := range configs {
		pbConfigs = append(pbConfigs, config.ToProtobuf())
	}

	return &pb.I18NBatchCreateResponse{
		Configs: pbConfigs,
	}, nil
}

// GetByConfig get i18n configuration by configuration key
func (dbClient *DBClient) GetByConfig(ctx context.Context, req *pb.I18NGetByConfigRequest) (*pb.I18NConfig, error) {
	config := &Config{}
	err := dbClient.DB.Where("category = ? AND item_key = ? AND field_name = ? AND locale = ?",
		req.Category, req.ItemKey, req.FieldName, req.Locale).First(config).Error
	if err != nil {
		return nil, err
	}
	return config.ToProtobuf(), nil
}
