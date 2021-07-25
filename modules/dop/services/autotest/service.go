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

package autotest

import (
	cmspb "github.com/erda-project/erda-proto-go/core/pipeline/cms/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
)

// Service autotest 实例对象封装
type Service struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
	cms cmspb.CmsServiceServer
}

// New 新建 autotest service
func New(options ...Option) *Service {
	var svc Service
	for _, op := range options {
		op(&svc)
	}
	return &svc
}

// Option autotest 实例对象配置选项
type Option func(*Service)

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(svc *Service) {
		svc.db = db
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *Service) {
		svc.bdl = bdl
	}
}

func WithPipelineCms(cms cmspb.CmsServiceServer) Option {
	return func(svc *Service) {
		svc.cms = cms
	}
}
