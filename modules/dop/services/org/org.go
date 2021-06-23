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

package org

import (
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/nexussvc"
	"github.com/erda-project/erda/modules/dop/services/publisher"
	"github.com/erda-project/erda/pkg/ucauth"
)

type Org struct {
	db        *dao.DBClient
	uc        *ucauth.UCClient
	bdl       *bundle.Bundle
	publisher *publisher.Publisher
	nexusSvc  *nexussvc.NexusSvc
	redisCli  *redis.Client
}

// Option 定义 Org 对象的配置选项
type Option func(*Org)

// New 新建 Org 实例，通过 Org 实例操作企业资源
func New(options ...Option) *Org {
	o := &Org{}
	for _, op := range options {
		op(o)
	}
	return o
}

// WithDBClient 配置 db client
func WithDBClient(db *dao.DBClient) Option {
	return func(o *Org) {
		o.db = db
	}
}

// WithUCClient 配置 uc client
func WithUCClient(uc *ucauth.UCClient) Option {
	return func(o *Org) {
		o.uc = uc
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(o *Org) {
		o.bdl = bdl
	}
}

// WithPublisher 配置 publisher
func WithPublisher(publisher *publisher.Publisher) Option {
	return func(o *Org) {
		o.publisher = publisher
	}
}

// WithNexusSvc 配置 nexus service
func WithNexusSvc(svc *nexussvc.NexusSvc) Option {
	return func(o *Org) {
		o.nexusSvc = svc
	}
}

// WithRedisClient 配置 redis client
func WithRedisClient(cli *redis.Client) Option {
	return func(o *Org) {
		o.redisCli = cli
	}
}

func (o *Org) GetPublisherID(orgID int64) int64 {
	pub, err := o.db.GetPublisherByOrgID(orgID)
	if err != nil && err != dao.ErrNotFoundPublisher {
		logrus.Warning(err)
		return 0
	}
	if pub == nil {
		return 0
	}
	return int64(pub.ID)
}
