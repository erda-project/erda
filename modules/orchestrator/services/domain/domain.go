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

package domain

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/modules/orchestrator/events"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/strutil"
)

// Domain 域名封装
type Domain struct {
	db    *dbclient.DBClient
	evMgr *events.EventManager
	bdl   *bundle.Bundle
}

// Option 域名对象配置选项
type Option func(*Domain)

// New 新建域名对象实例
func New(options ...Option) *Domain {
	d := &Domain{}
	for _, op := range options {
		op(d)
	}
	return d
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(d *Domain) {
		d.db = db
	}
}

// WithEventManager 配置 EventManager
func WithEventManager(evMgr *events.EventManager) Option {
	return func(d *Domain) {
		d.evMgr = evMgr
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(d *Domain) {
		d.bdl = bdl
	}
}

// List 查询域名列表
func (d *Domain) List(userID user.ID, orgID uint64, runtimeID uint64) (*apistructs.DomainGroup, error) {
	runtime, err := d.db.GetRuntime(runtimeID)
	if err != nil {
		return nil, apierrors.ErrListDomain.InternalError(err)
	}
	perm, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return nil, apierrors.ErrListDomain.InternalError(err)
	}
	if !perm.Access {
		return nil, apierrors.ErrListDomain.AccessDenied()
	}
	dc := newCtx(d.db, d.bdl)
	if err := dc.load(runtimeID); err != nil {
		return nil, err
	}
	return dc.GroupDomains(), nil
}

// Update 更新域名
func (d *Domain) Update(userID user.ID, orgID uint64, runtimeID uint64, group *apistructs.DomainGroup) error {
	runtime, err := d.db.GetRuntime(runtimeID)
	if err != nil {
		return apierrors.ErrUpdateDomain.InternalError(err)
	}
	perm, err := d.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  runtime.ApplicationID,
		Resource: "runtime-" + strutil.ToLower(runtime.Workspace),
		Action:   apistructs.OperateAction,
	})
	if err != nil {
		return apierrors.ErrUpdateDomain.InternalError(err)
	}
	if !perm.Access {
		return apierrors.ErrUpdateDomain.AccessDenied()
	}
	dc := newCtx(d.db, d.bdl)
	if err := dc.load(runtimeID); err != nil {
		return err
	}
	return dc.UpdateDomains(group)
}
