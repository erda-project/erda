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

package aksk

import (
	"context"
	"sync"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/pkg/secret"
)

type Service struct {
	db    *dao.DBClient
	cache *simpleCache
}

// Option 定义 Member 对象配置选项
type Option func(*Service)

// New 新建 Audit 实例
func New(options ...Option) (*Service, error) {
	s := &Service{
		cache: &simpleCache{
			store: make(map[AccessKeyID]model.AkSk),
		},
	}
	for _, op := range options {
		op(s)
	}

	if err := s.loadAllInternalAkSk(); err != nil {
		return nil, err
	}
	return s, nil
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(a *Service) {
		a.db = db
	}
}

// this is small and enough for internal key.
// TODO for user's api key
type simpleCache struct {
	store map[AccessKeyID]model.AkSk
	mu    sync.RWMutex
}

type AccessKeyID string

func (sc *simpleCache) Get(key AccessKeyID) (model.AkSk, bool) {
	sc.mu.RLock()
	obj, ok := sc.store[key]
	sc.mu.RUnlock()
	return obj, ok
}

func (sc *simpleCache) Set(key AccessKeyID, val model.AkSk) {
	sc.mu.Lock()
	sc.store[key] = val
	sc.mu.Unlock()
}

func (sc *simpleCache) Delete(key AccessKeyID) {
	sc.mu.Lock()
	delete(sc.store, key)
	sc.mu.Unlock()
}

func (s *Service) CreateAkSk(ctx context.Context, req apistructs.AkSkCreateRequest) (model.AkSk, error) {
	obj, err := s.db.CreateAkSk(toModel(req))
	if err != nil {
		return model.AkSk{}, err
	}
	if obj.IsSystem {
		s.cache.Set(AccessKeyID(obj.Ak), obj)
	}
	return obj, nil
}

func (s *Service) GetAkSkByAk(ctx context.Context, ak string) (model.AkSk, error) {
	if val, ok := s.cache.Get(AccessKeyID(ak)); ok {
		return val, nil
	}

	obj, err := s.db.GetAkSkByAk(ak)
	if err != nil {
		return model.AkSk{}, err
	}
	if obj.IsSystem {
		s.cache.Set(AccessKeyID(obj.Ak), obj)
	}
	return obj, nil
}

func (s *Service) DeleteAkSkByAk(ctx context.Context, ak string) error {
	err := s.db.DeleteAkSkByAk(ak)
	if err != nil {
		return err
	}
	s.cache.Delete(AccessKeyID(ak))
	return nil
}

func (s *Service) loadAllInternalAkSk() error {
	objs, err := s.db.ListAkSk(true)
	if err != nil {
		return err
	}
	for _, item := range objs {
		s.cache.Set(AccessKeyID(item.Ak), item)
	}
	return nil
}

func toModel(req apistructs.AkSkCreateRequest) model.AkSk {
	pair := secret.CreateAkSkPair()
	return model.AkSk{
		Ak:          pair.AccessKeyID,
		Sk:          pair.SecretKey,
		IsSystem:    req.Internal,
		SubjectID:   req.Owner,
		SubjectType: req.Scope,
		Description: req.Description,
	}
}
