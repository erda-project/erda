// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"crypto/md5" // #nosec G501
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/modules/hepa/bundle"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/config"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/kong"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
)

type GatewayGlobalServiceImpl struct {
	azDb   db.GatewayAzInfoService
	kongDb db.GatewayKongInfoService
}

var diceHealth *gw.DiceHealthDto = &gw.DiceHealthDto{Status: gw.DiceHealthOK}

var once sync.Once

func NewGatewayGlobalServiceImpl() (*GatewayGlobalServiceImpl, error) {
	azDb, err := db.NewGatewayAzInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	kongDb, err := db.NewGatewayKongInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	impl := &GatewayGlobalServiceImpl{
		azDb:   azDb,
		kongDb: kongDb,
	}
	once.Do(func() {
		go func() {
			defer func() {
				util.DoRecover()
				diceHealth.Status = gw.DiceHealthFail
				diceHealth.Modules = []gw.DiceHealthModule{
					{
						Name:    "panic-happened",
						Status:  gw.DiceHealthFail,
						Message: "panic happened",
					},
				}
			}()
			for range time.Tick(time.Second * 60) {
				newHealthInfo := impl.checkKongHealth()
				diceHealth = &newHealthInfo
			}
		}()
	})
	return impl, nil
}

func (impl *GatewayGlobalServiceImpl) checkKongHealth() (dto gw.DiceHealthDto) {
	var err error
	dto.Status = gw.DiceHealthOK
	defer func() {
		if err != nil {
			dto.Status = gw.DiceHealthFail
			dto.Modules = append(dto.Modules, gw.DiceHealthModule{
				Name:    "error-happened",
				Status:  gw.DiceHealthFail,
				Message: errors.Cause(err).Error(),
			})
			log.Errorf("error happened:%+v", err)
		}
	}()
	azs, err := impl.azDb.SelectValidAz()
	if err != nil {
		return
	}
	for _, az := range azs {
		data, err := bundle.Bundle.QueryClusterInfo(az.Az)
		if err != nil || data.IsDCOS() {
			continue
		}
		kongInfo, err := impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
			Az: az.Az,
		})
		if err != nil || kongInfo.KongAddr == "" || kongInfo.InnerAddr == "" {
			continue
		}
		adapter := kong.NewKongAdapter(kongInfo.KongAddr)
		_, err = adapter.GetRoutes()
		if err != nil {
			dto.Status = gw.DiceHealthFail
			dto.Modules = append(dto.Modules, gw.DiceHealthModule{
				Name:    fmt.Sprintf("cluster-gateway-control-%s", az.Az),
				Status:  gw.DiceHealthFail,
				Message: errors.Cause(err).Error(),
			})
			log.Errorf("error happened:%+v", err)
		}
		// proxyAddr := kongInfo.InnerAddr
		// if !strings.HasPrefix(proxyAddr, "inet") && strings.HasPrefix(kongInfo.KongAddr, "inet") {
		// 	addrsplit := strings.Split(kongInfo.KongAddr, "/")
		// 	inetPrefix := strings.Join(addrsplit[:len(addrsplit)-1], "/")
		// 	proxyAddr = strings.Replace(proxyAddr, "http:/", inetPrefix, 1)
		// 	if !strings.HasPrefix(proxyAddr, "inet") {
		// 		proxyAddr = fmt.Sprintf("%s/%s", inetPrefix, proxyAddr)
		// 	}
		// }
		// err = checkKongProxyHealth(proxyAddr, &dto, routes)
		// if err != nil {
		// 	dto.Status = gw.DiceHealthFail
		// 	dto.Modules = append(dto.Modules, gw.DiceHealthModule{
		// 		Name:    fmt.Sprintf("cluster-gateway-proxy-%s", az.Az),
		// 		Status:  gw.DiceHealthFail,
		// 		Message: errors.Cause(err).Error(),
		// 	})
		// 	log.Errorf("error happened:%+v", err)
		// }
	}
	return
}

func (impl *GatewayGlobalServiceImpl) GetDiceHealth() gw.DiceHealthDto {
	return *diceHealth
}

func (impl *GatewayGlobalServiceImpl) GetOrgId(projectId string) (string, error) {
	projectIdi, err := strconv.ParseUint(projectId, 10, 64)
	if err != nil {
		return "", err
	}
	projectDto, err := bundle.Bundle.GetProject(projectIdi)
	if err != nil {
		return "", err
	}
	orgId := projectDto.OrgID
	if orgId == 0 {
		return "", errors.Errorf("invalid project: %+v", projectDto)
	}
	return strconv.FormatUint(orgId, 10), nil
}

func (impl *GatewayGlobalServiceImpl) GetClusterUIType(orgId, projectId, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if projectId == "" || env == "" {
		log.Errorf("clusterName is empty")
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	az, err := impl.azDb.GetAzInfo(&orm.GatewayAzInfo{
		Env:       env,
		ProjectId: projectId,
	})
	if err != nil {
		log.Errorf("get az failed, err:%+v", err)
		return res
	}
	uiType := gw.UI_NORMAL
	if az.Type == orm.AT_K8S || az.Type == orm.AT_EDAS {
		uiType = gw.UI_K8S
	}
	if config.ServerConf.ClusterUIType != "" {
		uiType = config.ServerConf.ClusterUIType
	}
	return res.SetSuccessAndData(uiType)
}

func (impl *GatewayGlobalServiceImpl) GenerateEndpoint(info DiceInfo, session ...*db.SessionHelper) (string, string, error) {
	var kongService db.GatewayKongInfoService
	var err error
	if len(session) > 0 {
		kongService, err = impl.kongDb.NewSession(session[0])
		if err != nil {
			return "", "", err
		}
	} else {
		kongService = impl.kongDb
	}
	if info.Az == "" || info.ProjectId == "" || info.Env == "" {
		return "", "", errors.Errorf("invalid diceinfo:%+v", info)
	}
	kongInfo, err := kongService.GetKongInfo(&orm.GatewayKongInfo{
		Az:        info.Az,
		ProjectId: info.ProjectId,
		Env:       info.Env,
	})
	if err != nil {
		return "", "", err
	}
	endpoint := kongInfo.Endpoint
	inner := kong.InnerHost
	if !strings.EqualFold(info.Env, ENV_TYPE_PROD) {
		endpoint = strings.ToLower(info.Env + config.ServerConf.SubDomainSplit + endpoint)
		inner = strings.ToLower(info.Env + "." + inner)
	}
	return endpoint, inner, nil
}

func (impl *GatewayGlobalServiceImpl) GetServiceAddr(env string) string {
	addr := "api-gateway.kube-system.svc.cluster.local"
	if !strings.EqualFold(env, ENV_TYPE_PROD) {
		addr = strings.ToLower(env + "-" + addr)
	}
	return addr
}

func (impl *GatewayGlobalServiceImpl) GenerateDefaultPath(projectId string, session ...*db.SessionHelper) (string, error) {
	if projectId == "" {
		return "", errors.New("empty projectId")
	}
	projectName, err := impl.GetProjectName(DiceInfo{ProjectId: projectId}, session...)
	if err != nil {
		return "", err
	}
	return strings.ToLower(fmt.Sprintf("/%s", projectName)), nil
}

func (impl *GatewayGlobalServiceImpl) GetProjectNameFromCmdb(projectId string) (string, error) {
	idNum, err := strconv.ParseUint(projectId, 10, 64)
	if err != nil {
		return "", nil
	}
	projectInfo, err := bundle.Bundle.GetProject(idNum)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return projectInfo.Name, nil
}

func (impl *GatewayGlobalServiceImpl) GetClustersByOrg(orgId string) ([]string, error) {
	idNum, err := strconv.ParseUint(orgId, 10, 64)
	if err != nil {
		return nil, nil
	}
	clusters, err := bundle.Bundle.ListClusters("", idNum)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var names []string
	for _, cluster := range clusters {
		names = append(names, cluster.Name)
	}
	return names, nil
}

func (impl *GatewayGlobalServiceImpl) GetProjectName(info DiceInfo, session ...*db.SessionHelper) (string, error) {
	var kongService db.GatewayKongInfoService
	var err error
	if len(session) > 0 {
		kongService, err = impl.kongDb.NewSession(session[0])
		if err != nil {
			return "", err
		}
	} else {
		kongService = impl.kongDb
	}
	if info.ProjectId == "" {
		return "", errors.Errorf("invalid diceinfo:%+v", info)
	}
	kongInfo, err := kongService.GetKongInfo(&orm.GatewayKongInfo{
		ProjectId: info.ProjectId,
	})
	if err != nil {
		return "", err
	}
	return kongInfo.ProjectName, nil
}

// md5V md5加密
func md5V(str string) string {
	h := md5.New() // #nosec G401
	_, _ = h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func (impl *GatewayGlobalServiceImpl) GenTenantGroup(projectId, env, clusterName string) string {
	tenantGroup := md5V(projectId + "_" + strings.ToUpper(env) + "_" + clusterName + config.ServerConf.TenantGroupKey)
	tenant, err := bundle.Bundle.CreateMSPTenant(projectId, env, pb.Type_DOP.String(), tenantGroup)
	if err != nil {
		return ""
	}
	return tenant
}

func (impl *GatewayGlobalServiceImpl) GetTenantGroup(projectId, env string) (res *common.StandardResult) {
	var err error
	res = &common.StandardResult{Success: false}
	defer func() {
		if err != nil {
			log.Errorf("error happened: %+v", err)
			res.SetErrorInfo(&common.ErrInfo{
				Msg: errors.Cause(err).Error(),
			})
		}
	}()
	info, err := impl.kongDb.GetByAny(&orm.GatewayKongInfo{
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		return
	}
	if info == nil {
		err = errors.New("tenant not found")
		return
	}
	tenantGroup := info.TenantGroup
	if tenantGroup == "" {
		tenantGroup = impl.GenTenantGroup(projectId, env, info.Az)
	}
	res.SetSuccessAndData(tenantGroup)
	return
}

func (impl *GatewayGlobalServiceImpl) CreateTenant(tenant *gw.TenantDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var session *db.SessionHelper
	var kongSession db.GatewayKongInfoService
	var az *orm.GatewayAzInfo
	exist, err := impl.kongDb.GetByAny(&orm.GatewayKongInfo{
		TenantId: tenant.Id,
	})
	if err != nil {
		log.Errorf("create tenant failed, err:%+v", err)
		return res
	}
	if exist != nil {
		log.Infof("tenant already exists, tenant:%+v", tenant)
		return res.SetSuccessAndData(true)
	}
	session, err = db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	kongSession, err = impl.kongDb.NewSession(session)
	if err != nil {
		goto failed
	}
	exist, err = kongSession.GetByAny(&orm.GatewayKongInfo{
		Az:        tenant.Az,
		Env:       tenant.Env,
		ProjectId: tenant.ProjectId,
	})
	if err != nil {
		goto failed
	}
	if exist != nil {
		err = kongSession.Update(&orm.GatewayKongInfo{
			BaseRow: orm.BaseRow{
				Id: exist.Id,
			},
			TenantId:        tenant.Id,
			TenantGroup:     tenant.TenantGroup,
			Az:              tenant.Az,
			Env:             tenant.Env,
			ProjectId:       tenant.ProjectId,
			ProjectName:     tenant.ProjectName,
			KongAddr:        tenant.AdminAddr,
			Endpoint:        tenant.GatewayEndpoint,
			InnerAddr:       tenant.InnerAddr,
			ServiceName:     tenant.ServiceName,
			AddonInstanceId: tenant.InstanceId,
		})
		if err != nil {
			goto failed
		}
	} else {
		err = kongSession.Insert(&orm.GatewayKongInfo{
			TenantId:        tenant.Id,
			TenantGroup:     tenant.TenantGroup,
			Az:              tenant.Az,
			Env:             tenant.Env,
			ProjectId:       tenant.ProjectId,
			ProjectName:     tenant.ProjectName,
			KongAddr:        tenant.AdminAddr,
			Endpoint:        tenant.GatewayEndpoint,
			InnerAddr:       tenant.InnerAddr,
			ServiceName:     tenant.ServiceName,
			AddonInstanceId: tenant.InstanceId,
		})
		if err != nil {
			goto failed
		}
		az, err = impl.azDb.GetAzInfoByClusterName(tenant.Az)
		if err != nil {
			goto failed
		}
		if az == nil {
			err = errors.Errorf("get az failed, tenant:%+v", tenant)
			goto failed
		}
		if az.Type == orm.AT_K8S || az.Type == orm.AT_EDAS {
			var packageBiz GatewayOpenapiService
			packageBiz, err = NewGatewayOpenapiServiceImpl()
			if err != nil {
				goto failed
			}
			err = packageBiz.CreateTenantPackage(tenant.Id, session)
			if err != nil {
				goto failed
			}
		}
	}
	err = session.Commit()
	if err != nil {
		goto failed
	}
	session.Close()
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	if session != nil {
		_ = session.Rollback()
		session.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})

}

func (impl GatewayGlobalServiceImpl) GetRuntimeServicePrefix(dao *orm.GatewayRuntimeService) (string, error) {
	return "/" + dao.Id, nil
}

func (impl GatewayGlobalServiceImpl) GetGatewayFeatures(clusterName string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	return res.SetSuccessAndData(map[string]string{
		"runtime-register": "on",
	})
}
