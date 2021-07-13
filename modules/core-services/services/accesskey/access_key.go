// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
