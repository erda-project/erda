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
	"fmt"
	"os"
	"strings"

	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/repository/orm"

	"github.com/pkg/errors"
	"github.com/xormplus/xorm"
)

type GatewayKongInfoServiceImpl struct {
	engine *orm.OrmEngine
	*SessionHelper
	executor xorm.Interface
}

func NewGatewayKongInfoServiceImpl() (*GatewayKongInfoServiceImpl, error) {
	engine, error := orm.GetSingleton()
	if error != nil {
		return nil, errors.Wrap(error, "new GatewayKongInfoServiceImpl failed")
	}
	return &GatewayKongInfoServiceImpl{
		engine:   engine,
		executor: engine,
	}, nil
}

func (impl *GatewayKongInfoServiceImpl) NewSession(helper ...*SessionHelper) (GatewayKongInfoService, error) {
	var session *SessionHelper
	var err error
	if len(helper) == 0 {
		session, err = NewSessionHelper()
		if err != nil {
			return nil, err
		}
	} else if helper[0] == nil {
		return &GatewayKongInfoServiceImpl{
			engine:   impl.engine,
			executor: impl.engine,
		}, nil
	} else {
		session = helper[0]
	}
	return &GatewayKongInfoServiceImpl{
		engine:        impl.engine,
		executor:      session.session,
		SessionHelper: session,
	}, nil
}

func (impl *GatewayKongInfoServiceImpl) GetBelongTuples(instanceId string) ([]KongBelongTuple, error) {
	var result []orm.GatewayKongInfo
	err := orm.Select(impl.executor.Distinct("env", "project_id", "az"), &result, "addon_instance_id = ?", instanceId)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	var res []KongBelongTuple
	for _, info := range result {
		res = append(res, KongBelongTuple{
			Env:       info.Env,
			ProjectId: info.ProjectId,
			Az:        info.Az,
		})
	}
	return res, nil
}

func (impl *GatewayKongInfoServiceImpl) GenK8SInfo(kongInfo *orm.GatewayKongInfo) (string, string, error) {
	return fmt.Sprintf("addon-%s--%s", kongInfo.ServiceName, kongInfo.AddonInstanceId), kongInfo.ServiceName, nil
}

func (impl *GatewayKongInfoServiceImpl) GetK8SInfo(cond *orm.GatewayKongInfo) (string, string, error) {
	kongInfo, err := impl.GetKongInfo(cond)
	if err != nil {
		return "", "", err
	}
	return impl.GenK8SInfo(kongInfo)
}

func (impl *GatewayKongInfoServiceImpl) GetKongInfo(cond *orm.GatewayKongInfo) (*orm.GatewayKongInfo, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	kongInfo, err := impl.GetByAny(cond)
	if err != nil {
		return nil, err
	}
	if kongInfo == nil {
		kongInfo, err = impl.GetByAny(&orm.GatewayKongInfo{
			Az: cond.Az,
		})
		if err != nil {
			return nil, err
		}
	}
	if kongInfo == nil {
		return nil, errors.Errorf("get kong info faild, %+v", cond)
	}
	return kongInfo, nil
}

func (impl *GatewayKongInfoServiceImpl) GetByAny(cond *orm.GatewayKongInfo) (*orm.GatewayKongInfo, error) {
	if cond == nil {
		return nil, errors.New(ERR_INVALID_ARG)
	}
	dao := &orm.GatewayKongInfo{}
	bCond, err := orm.BuildConds(impl.engine, cond, cond.GetMustCondCols())
	if err != nil {
		return nil, errors.Wrap(err, "buildConds failed")
	}
	succ, err := orm.GetByAnyI(impl.executor, bCond, dao)
	if err != nil {
		return nil, errors.Wrap(err, ERR_SQL_FAIL)
	}
	if !succ {
		return nil, nil
	}
	return dao, nil
}

func (impl *GatewayKongInfoServiceImpl) GetTenantId(projectId, env, az string) (string, error) {
	kongInfo, err := impl.GetByAny(&orm.GatewayKongInfo{
		ProjectId: projectId,
		Env:       env,
		Az:        az,
	})
	if err != nil {
		return "", err
	}
	if kongInfo == nil {
		return "", errors.Errorf("get kong info failed, projectId:%s, env:%s, az:%s",
			projectId, env, az)
	}
	return kongInfo.TenantId, nil
}

func (impl *GatewayKongInfoServiceImpl) adjustKonfInfo(info *orm.GatewayKongInfo) error {
	selfAz := os.Getenv("DICE_CLUSTER_NAME")
	// 同集群不走netportal
	if strings.HasPrefix(info.KongAddr, "inet://") {
		if selfAz == info.Az {
			pathSlice := strings.SplitN(strings.TrimPrefix(info.KongAddr, "inet://"), "/", 2)
			if len(pathSlice) != 2 {
				return errors.Errorf("invalid addr:%s", info.KongAddr)
			}
			info.KongAddr = "http://" + pathSlice[1]
		}
	} else if !strings.HasPrefix(info.KongAddr, "http://") || !strings.HasPrefix(info.KongAddr, "https://") {
		info.KongAddr = "https://" + info.KongAddr
	}
	// 兼容海油
	if config.ServerConf.UseAdminEndpoint && selfAz != info.Az {
		info.KongAddr = "http://" + strings.Replace(info.Endpoint, "gateway", "gateway-admin", 1)
	}
	return nil
}

func (impl *GatewayKongInfoServiceImpl) Update(info *orm.GatewayKongInfo) error {
	if info == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	err := impl.adjustKonfInfo(info)
	if err != nil {
		return err
	}
	_, err = orm.Update(impl.executor, info)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}

func (impl *GatewayKongInfoServiceImpl) Insert(info *orm.GatewayKongInfo) error {
	if info == nil {
		return errors.New(ERR_INVALID_ARG)
	}
	err := impl.adjustKonfInfo(info)
	if err != nil {
		return err
	}
	_, err = orm.Insert(impl.executor, info)
	if err != nil {
		return errors.Wrap(err, ERR_SQL_FAIL)
	}
	return nil
}
