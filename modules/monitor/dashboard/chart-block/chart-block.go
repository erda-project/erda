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

package block

import (
	"net/http"
	"time"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/modules/pkg/mysql"
	"github.com/jinzhu/gorm"
)

func (p *provider) getSystemBlock(params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	obj, err := p.db.systemBlock.Get(&DashboardBlockQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}
	if obj.ViewConfig != nil && obj.DataConfig != nil {
		for _, v := range *obj.ViewConfig {
			v.View.StaticData = struct{}{}
			for _, d := range *obj.DataConfig {
				if v.I == d.I {
					v.View.StaticData = d.StaticData
				}
			}
		}
	}
	return api.Success(&DashboardBlockDTO{
		ID:         obj.ID,
		Name:       obj.Name,
		Desc:       obj.Desc,
		Scope:      obj.Scope,
		ScopeID:    obj.ScopeID,
		ViewConfig: obj.ViewConfig,
		CreatedAt:  utils.ConvertTimeToMS(obj.CreatedAt),
		UpdatedAt:  utils.ConvertTimeToMS(obj.UpdatedAt),
		Version:    obj.Version,
	})
}

func (p *provider) listBlockSystem(params struct {
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int64  `query:"pageNo" validate:"gte=0"`
	PageSize int64  `query:"pageSize" validate:"gte=0"`
}) interface{} {

	objs, total, err := p.db.systemBlock.List(&DashboardBlockQuery{
		Scope:   params.Scope,
		ScopeID: params.ScopeID,
	}, params.PageSize, params.PageNo)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return api.Errors.Internal("failed to list system dashboard :", err)
	}
	var blocks []*DashboardBlockDTO
	for _, obj := range objs {
		if obj.ViewConfig != nil && obj.DataConfig != nil {
			for _, v := range *obj.ViewConfig {
				for _, d := range *obj.DataConfig {
					if v.I == d.I {
						v.View.StaticData = d.StaticData
					}
				}
			}
		}
		block := &DashboardBlockDTO{
			ID:         obj.ID,
			Name:       obj.Name,
			Desc:       obj.Desc,
			Scope:      obj.Scope,
			ScopeID:    obj.ScopeID,
			ViewConfig: obj.ViewConfig,
			CreatedAt:  utils.ConvertTimeToMS(obj.CreatedAt),
			UpdatedAt:  utils.ConvertTimeToMS(obj.UpdatedAt),
			Version:    obj.Version,
		}
		blocks = append(blocks, block)
	}
	return api.Success(&dashboardBlockResp{
		DashboardBlocks: blocks,
		Total:           total,
	})
}

func (p *provider) delSystemBlock(params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	err := p.db.systemBlock.Delete(&DashboardBlockQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) delUserBlock(params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	err := p.db.userBlock.Delete(&DashboardBlockQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) createBlockSystem(body SystemBlock) interface{} {
	err := p.db.systemBlock.Save(&body)
	if err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("block")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) createUserBlock(query struct {
	ScopeID string `query:"scopeId"`
}, body UserBlock) interface{} {
	if len(query.ScopeID) > 0 {
		body.ScopeID = query.ScopeID
	}
	dash, err := p.CreateDashboard(&body)
	if err != nil {
		return api.Errors.Internal(err)
	}
	return api.Success(dash)
}

func (p *provider) listUserBlock(params struct {
	Scope    string `query:"scope" validate:"required"`
	ScopeID  string `query:"scopeId" validate:"required"`
	PageNo   int64  `query:"pageNo" validate:"gte=0"`
	PageSize int64  `query:"pageSize" validate:"gte=0"`
}) interface{} {

	objs, total, err := p.db.userBlock.List(&DashboardBlockQuery{
		Scope:         params.Scope,
		ScopeID:       params.ScopeID,
		CreatedAtDesc: true,
	}, params.PageSize, params.PageNo)

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return api.Errors.Internal("failed to list user dashboard :", err)
	}

	var blocks []*DashboardBlockDTO
	for _, obj := range objs {
		blockDTO := p.getDashboardDTOWithUserBlock(obj)
		blocks = append(blocks, blockDTO)
	}

	return api.Success(&dashboardBlockResp{
		DashboardBlocks: blocks,
		Total:           total,
	})
}

func (p *provider) getDashboardDTOWithUserBlock(obj *UserBlock) *DashboardBlockDTO {
	if obj.ViewConfig != nil {
		for _, v := range *obj.ViewConfig {
			v.View.StaticData = struct{}{}
			if obj.DataConfig != nil {
				for _, d := range *obj.DataConfig {
					if v.I == d.I {
						v.View.StaticData = d.StaticData
					}
				}
			}
		}
	}
	return &DashboardBlockDTO{
		ID:         obj.ID,
		Name:       obj.Name,
		Desc:       obj.Desc,
		Scope:      obj.Scope,
		ScopeID:    obj.ScopeID,
		ViewConfig: obj.ViewConfig,
		DataConfig: obj.DataConfig,
		CreatedAt:  utils.ConvertTimeToMS(obj.CreatedAt),
		UpdatedAt:  utils.ConvertTimeToMS(obj.UpdatedAt),
		Version:    obj.Version,
	}
}

func (p *provider) getUserBlock(r *http.Request, params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	obj, err := p.db.userBlock.Get(&DashboardBlockQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(p.getDashboardDTOWithUserBlock(obj))
}

func (p *provider) updateUserBlock(params struct {
	ID string `param:"id" validate:"required"`
}, update UserBlockUpdate) interface{} {
	tx := p.db.Begin()
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.Log.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	block, err := tx.userBlock.Get(&DashboardBlockQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}

	block = editUserBlockFields(block, &update)
	if err := tx.userBlock.Save(block); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("block")
		}
		return api.Errors.Internal(err)
	}

	return api.Success(true)
}

func editUserBlockFields(block *UserBlock, update *UserBlockUpdate) *UserBlock {
	if update.Name != nil {
		block.Name = *update.Name
	}
	if update.Desc != nil {
		block.Desc = *update.Desc
	}
	if update.ViewConfig != nil {
		block.ViewConfig = update.ViewConfig
	}
	if update.DataConfig != nil {
		block.DataConfig = update.DataConfig
	}
	block.UpdatedAt = time.Now()
	return block
}
