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

package activity

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core-services/dao"
	"github.com/erda-project/erda/modules/core-services/model"
)

// Activity 活动操作封装
type Activity struct {
	db  *dao.DBClient
	bdl *bundle.Bundle
}

// Option 定义 Activity 配置选项
type Option func(*Activity)

// New 新建 Activity 实例
func New(options ...Option) *Activity {
	t := &Activity{}
	for _, op := range options {
		op(t)
	}
	return t
}

// WithDBClient 配置 Activity 数据库选项
func WithDBClient(db *dao.DBClient) Option {
	return func(t *Activity) {
		t.db = db
	}
}

// WithBundle 配置 Ticket bundle选项
func WithBundle(bdl *bundle.Bundle) Option {
	return func(t *Activity) {
		t.bdl = bdl
	}
}

// ListByRuntime 根据runtimeID获取活动列表
func (a *Activity) ListByRuntime(runtimeID int64, pageNo, pageSize int) (int, []model.Activity, error) {
	return a.db.GetActivitiesByRuntime(runtimeID, pageNo, pageSize)
}
