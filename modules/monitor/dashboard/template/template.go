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

package template

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/modules/pkg/mysql"
	"github.com/erda-project/erda/pkg/crypto/uuid"
)

func (p *provider) createTemplate(query struct {
	ScopeID string `query:"scopeId"`
}, body Template) interface{} {
	if len(query.ScopeID) > 0 {
		body.ScopeID = query.ScopeID
	}
	if len(body.ID) == 0 {
		body.ID = uuid.UUID()
	}
	if err := p.db.templateDB.Save(&body); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.Internal(fmt.Errorf("aleady existed, err: %s", err))
		}
		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) deleteTemplate(params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	err := p.db.templateDB.authTemplate(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}
	err = p.db.templateDB.Delete(&templateQuery{ID: params.ID})
	if err != nil {

		return api.Errors.Internal(err)
	}
	return api.Success(true)
}

func (p *provider) updateTemplate(params struct {
	ID string `param:"id" validate:"required"`
}, update templateUpdate) interface{} {
	err := p.db.templateDB.authTemplate(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}

	tx := p.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else if err := recover(); err != nil {
			p.L.Errorf("panic: %s", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	block, err := tx.templateDB.Get(&templateQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}

	block = editTemplateFields(block, &update)
	if err := tx.templateDB.Save(block); err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return api.Errors.AlreadyExists("block")
		}
		return api.Errors.Internal(err)
	}

	return api.Success(true)
}

func (p *provider) getTemplate(r *http.Request, params struct {
	ID string `param:"id" validate:"required"`
}) interface{} {
	err := p.db.templateDB.authTemplate(params.ID)
	if err != nil {
		return api.Errors.Internal(err)
	}

	obj, err := p.db.templateDB.Get(&templateQuery{ID: params.ID})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return api.Errors.NotFound("block")
		}
		return api.Errors.Internal(err)
	}
	return api.Success(p.templateFormat(obj))
}

func (p *provider) getListTemplates(params templateSearch) interface{} {
	objs, total, err := p.db.templateDB.List(&templateQuery{
		Scope:   params.Scope,
		ScopeID: params.ScopeID,
		Type:    params.Type,
		Name:    params.Name,
	}, params.PageSize, params.PageNo)

	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return api.Errors.Internal("failed to list templates :", err)
	}

	var templates []*templateOverview
	for _, obj := range objs {
		blockDTO := p.templateOverviewFormat(obj)
		templates = append(templates, blockDTO)
	}

	return api.Success(&templateResp{
		TemplateDTO: templates,
		Total:       total,
	})
}

func (p *provider) templateFormat(obj *Template) *templateDTO {
	if obj.ViewConfig != nil {
		for _, v := range *obj.ViewConfig {
			v.View.StaticData = struct{}{}
		}
	}
	return &templateDTO{
		ID:          obj.ID,
		Name:        obj.Name,
		Description: obj.Description,
		Scope:       obj.Scope,
		ScopeID:     obj.ScopeID,
		ViewConfig:  obj.ViewConfig,
		CreatedAt:   utils.ConvertTimeToMS(obj.CreatedAt),
		UpdatedAt:   utils.ConvertTimeToMS(obj.UpdatedAt),
		Version:     obj.Version,
	}
}

func (p *provider) templateOverviewFormat(obj *Template) *templateOverview {
	return &templateOverview{
		ID:          obj.ID,
		Name:        obj.Name,
		Description: obj.Description,
		Scope:       obj.Scope,
		ScopeID:     obj.ScopeID,
		CreatedAt:   utils.ConvertTimeToMS(obj.CreatedAt),
		UpdatedAt:   utils.ConvertTimeToMS(obj.UpdatedAt),
		Version:     obj.Version,
	}
}

func editTemplateFields(block *Template, update *templateUpdate) *Template {
	if update.Name != nil {
		block.Name = *update.Name
	}
	if update.Description != nil {
		block.Description = *update.Description
	}
	if update.ViewConfig != nil {
		block.ViewConfig = update.ViewConfig
	}
	block.UpdatedAt = time.Now()
	return block
}

// If template type was 2 ,operation is not be allowed
// If template type was 1 ,operator should be admin
func (db *templateDB) authTemplate(ID string) error {
	var result templateType
	err := db.DB.Table(tableTemplate).Select("type").Where("id=?", ID).First(&result).Error
	if err != nil {
		return err
	}
	if result.Type == 2 {
		return fmt.Errorf("system template operation is not be allowed")
	}
	if result.Type == 1 {
		//TODO auth admin
	}
	return nil
}
