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

package accesskey

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
	"github.com/erda-project/erda/pkg/secret"
)

type Service struct {
	db *dao.DBClient
}

// Option 定义 Member 对象配置选项
type Option func(*Service)

// New 新建 Audit 实例
func New(options ...Option) (*Service, error) {
	s := &Service{}
	for _, op := range options {
		op(s)
	}

	return s, nil
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(a *Service) {
		a.db = db
	}
}

func (s *Service) CreateAccessKey(ctx context.Context, req apistructs.AccessKeyCreateRequest) (model.AccessKey, error) {
	obj, err := s.db.CreateAccessKey(toModel(req))
	if err != nil {
		return model.AccessKey{}, err
	}
	return obj, nil
}

func (s *Service) UpdateAccessKey(ctx context.Context, ak string, req apistructs.AccessKeyUpdateRequest) (model.AccessKey, error) {
	_, err := s.db.UpdateAccessKey(ak, req)
	if err != nil {
		return model.AccessKey{}, err
	}
	obj, err := s.db.GetByAccessKeyID(ak)
	if err != nil {
		return model.AccessKey{}, err
	}
	return obj, nil
}

func (s *Service) GetByAccessKeyID(ctx context.Context, ak string) (model.AccessKey, error) {
	obj, err := s.db.GetByAccessKeyID(ak)
	if err != nil {
		return model.AccessKey{}, err
	}
	return obj, nil
}

func (s *Service) ListAccessKey(ctx context.Context, req apistructs.AccessKeyListQueryRequest) ([]model.AccessKey, error) {
	obj, err := s.db.ListAccessKey(req)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *Service) DeleteByAccessKeyID(ctx context.Context, ak string) error {
	err := s.db.DeleteByAccessKeyID(ak)
	if err != nil {
		return err
	}
	return nil
}

func toModel(req apistructs.AccessKeyCreateRequest) model.AccessKey {
	// todo verify SubjectType
	pair := secret.CreateAkSkPair()
	return model.AccessKey{
		AccessKeyID: pair.AccessKeyID,
		SecretKey:   pair.SecretKey,
		IsSystem:    req.IsSystem,
		Status:      apistructs.AccessKeyStatusActive,
		SubjectType: req.SubjectType,
		Subject:     req.Subject,
		Description: req.Description,
	}
}
