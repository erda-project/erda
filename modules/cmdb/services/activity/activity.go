package activity

import (
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/dao"
	"github.com/erda-project/erda/modules/cmdb/model"
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
