package approval

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/dao"
)

type Approval struct {
	db *dao.DBClient
}

// Option
type Option func(*Approval)

// New 新建 approval service
func New(options ...Option) *Approval {
	a := &Approval{}
	for _, op := range options {
		op(a)
	}
	return a
}

// WithDBClient 设置 dbclient
func WithDBClient(db *dao.DBClient) Option {
	return func(a *Approval) {
		a.db = db
	}
}

// WatchApprovalStatusChanged 监听审批流状态变更，同步审批流状态至依赖方
func (a *Approval) WatchApprovalStatusChanged(event *apistructs.ApprovalStatusChangedEvent) error {
	// 更新审批流状态至库引用
	if err := a.db.UpdateApprovalStatusByApprovalID(event.Content.ApprovalID, event.Content.ApprovalStatus); err != nil {
		return err
	}
	return nil
}
