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

package service

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/modules/hepa/common"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GatewayUpstreamApiServiceImpl struct {
	engine *orm.OrmEngine
}

func NewGatewayUpstreamApiServiceImpl() (*GatewayUpstreamApiServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayUpstreamApiServiceImpl failed")
	}
	return &GatewayUpstreamApiServiceImpl{engine}, nil
}

func (impl *GatewayUpstreamApiServiceImpl) Insert(session *xorm.Session, item *orm.GatewayUpstreamApi) (string, error) {
	if session == nil || item == nil || len(item.UpstreamId) == 0 || item.RegisterId == "" {
		return "", errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Insert(session, item)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return item.Id, nil
}

func (impl *GatewayUpstreamApiServiceImpl) updateFields(update *orm.GatewayUpstreamApi, fields ...string) error {
	if update == nil || len(update.Id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Update(impl.engine, update, fields...)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (impl *GatewayUpstreamApiServiceImpl) GetLastApiId(cond *orm.GatewayUpstreamApi) string {
	var descList []orm.GatewayUpstreamApi
	err := orm.Desc(impl.engine, "register_id").And("upstream_id = ? and api_name = ? and api_id != ''",
		cond.UpstreamId, cond.ApiName).Find(&descList)
	if err != nil {
		log.Errorf("error happend:%+v", errors.WithStack(err))
		return ""
	}
	if len(descList) == 0 {
		log.Errorf("can't find upstream_id:%s api_name:%s has api_id", cond.UpstreamId, cond.ApiName)
		return ""
	}
	return descList[0].ApiId
}

func (impl *GatewayUpstreamApiServiceImpl) UpdateApiId(update *orm.GatewayUpstreamApi) error {
	return impl.updateFields(update, "api_id")
}

func (impl *GatewayUpstreamApiServiceImpl) countInIds(ids []string) (int64, error) {
	return orm.In(impl.engine, "id", ids).Count(&orm.GatewayUpstreamApi{})
}

func (impl *GatewayUpstreamApiServiceImpl) GetPage(ids []string, page *common.Page) (*common.PageQuery, error) {
	if page == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	total, err := impl.countInIds(ids)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	page.SetTotalNum(total)
	if total == 0 {
		return &common.PageQuery{Result: []orm.GatewayUpstreamApi{}, Page: page}, nil
	}
	var result []orm.GatewayUpstreamApi
	err = orm.SelectPageNoCond(impl.engine.In("id", ids).Desc("create_time"), &result, page)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &common.PageQuery{Result: result, Page: page}, nil
}

func (impl *GatewayUpstreamApiServiceImpl) SelectInIdsAndDeleted(ids []string) ([]orm.GatewayUpstreamApi, error) {
	var result []orm.GatewayUpstreamApi
	err := orm.SelectNoCondMissing(impl.engine.Desc("create_time").In("id", ids), &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (impl *GatewayUpstreamApiServiceImpl) SelectInIds(ids []string) ([]orm.GatewayUpstreamApi, error) {
	var result []orm.GatewayUpstreamApi
	err := orm.SelectNoCond(impl.engine.Desc("create_time").In("id", ids), &result)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (impl *GatewayUpstreamApiServiceImpl) Recover(id string) error {
	_, err := orm.Recover(impl.engine, &orm.GatewayUpstreamApi{}, id)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (impl *GatewayUpstreamApiServiceImpl) GetById(id string) (*orm.GatewayUpstreamApi, error) {
	if len(id) == 0 {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayUpstreamApi{}
	succ, err := orm.Get(impl.engine, dao, "id = ?", id)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl *GatewayUpstreamApiServiceImpl) DeleteById(id string) error {
	if len(id) == 0 {
		return errors.New(ERR_INVALID_ARG)
	}
	_, err := orm.Delete(impl.engine, &orm.GatewayUpstreamApi{}, "id = ?", id)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}
