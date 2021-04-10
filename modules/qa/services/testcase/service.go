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

package testcase

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/qa/dao"
	"github.com/erda-project/erda/pkg/httpclient"
)

// Service testCase 实例对象封装
type Service struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
	hc  *httpclient.HTTPClient

	CreateTestSetFn func(apistructs.TestSetCreateRequest) (*apistructs.TestSet, error)
}

// New 新建 testcase service
func New(options ...Option) *Service {
	var svc Service
	for _, op := range options {
		op(&svc)
	}
	return &svc
}

// Option testCase 实例对象配置选项
type Option func(*Service)

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(svc *Service) {
		svc.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *Service) {
		svc.bdl = bdl
	}
}
