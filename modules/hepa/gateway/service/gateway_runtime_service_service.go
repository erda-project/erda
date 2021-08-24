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
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/bundle"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/k8s"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type GatewayRuntimeServiceServiceImpl struct {
	runtimeDb  db.GatewayRuntimeServiceService
	azDb       db.GatewayAzInfoService
	kongDb     db.GatewayKongInfoService
	packageBiz GatewayOpenapiService
	domainBiz  GatewayDomainService
	apiBiz     GatewayApiService
}

func NewGatewayRuntimeServiceServiceImpl() (*GatewayRuntimeServiceServiceImpl, error) {
	runtimeDb, err := db.NewGatewayRuntimeServiceServiceImpl()
	if err != nil {
		return nil, err
	}
	azDb, err := db.NewGatewayAzInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	packageBiz, err := NewGatewayOpenapiServiceImpl()
	if err != nil {
		return nil, err
	}
	domainBiz, err := NewGatewayDomainServiceImpl()
	if err != nil {
		return nil, err
	}
	apiBiz, err := NewGatewayApiServiceImpl()
	if err != nil {
		return nil, err
	}
	kongDb, err := db.NewGatewayKongInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	return &GatewayRuntimeServiceServiceImpl{
		runtimeDb:  runtimeDb,
		azDb:       azDb,
		kongDb:     kongDb,
		packageBiz: packageBiz,
		domainBiz:  domainBiz,
		apiBiz:     apiBiz,
	}, nil

}

type runtimeService struct {
	dto gw.ServiceDetailDto
	dao *orm.GatewayRuntimeService
}

func diffServices(reqServices []gw.ServiceDetailDto, existServices []orm.GatewayRuntimeService) (addOrUpdates []runtimeService, dels []orm.GatewayRuntimeService) {
	for _, service := range reqServices {
		exist := false
		for i := len(existServices) - 1; i >= 0; i-- {
			serviceObj := existServices[i]
			if service.ServiceName == serviceObj.ServiceName {
				exist = true
				existServices = append(existServices[:i], existServices[i+1:]...)
				addOrUpdates = append(addOrUpdates, runtimeService{service, &serviceObj})
				break
			}

		}
		if !exist {
			addOrUpdates = append(addOrUpdates, runtimeService{service, nil})
		}
	}
	dels = existServices
	return
}

func (impl GatewayRuntimeServiceServiceImpl) TouchRuntimeComplete(c *gin.Context, reqDto *gw.RuntimeServiceReqDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var err error
	var ok bool
	var sessionI, endpointsI interface{}
	var session *db.SessionHelper
	var package2Endpoint map[string]*orm.GatewayRuntimeService
	sessionI, ok = c.Get("dbsession")
	if !ok {
		err = errors.New("can't find dbsession from context")
		goto failed
	}
	session, ok = sessionI.(*db.SessionHelper)
	if !ok {
		err = errors.New("acquire sesion failed")
		goto failed
	}
	endpointsI, ok = c.Get("endpoints")
	if !ok {
		err = errors.New("can't find endpoints from context")
		goto failed
	}
	package2Endpoint, ok = endpointsI.(map[string]*orm.GatewayRuntimeService)
	if !ok {
		err = errors.New("acquire endpoints failed")
		goto failed
	}
	// refresh endpoint services' relation package (ingress & endpoint api)
	for packageId, endpoint := range package2Endpoint {
		_, err = impl.domainBiz.GiveRuntimeDomainToPackage(endpoint, session)
		if err != nil {
			goto failed
		}
		err = impl.packageBiz.RefreshRuntimePackage(packageId, endpoint, session)
		if err != nil {
			goto failed
		}
	}
	err = session.Commit()
	if err != nil {
		goto failed
	}
	session.Close()

	go func() {
		defer util.DoRecover()
		runtimeEndpointsI, ok := c.Get("runtime_endpoints")
		if !ok {
			log.Errorf("can't find dice endpoints from context")
			return
		}
		runtimeEndpoints, ok := runtimeEndpointsI.([]RuntimeEndpointInfo)
		if !ok {
			log.Errorf("acquire dice endpoints failed")
			return
		}
		for _, runtimeEndpoint := range runtimeEndpoints {
			err = impl.packageBiz.SetRuntimeEndpoint(runtimeEndpoint)
			if err != nil {
				log.Errorf("set runtime endpoint failed, err:%+v, runtimeEndpoint:%+v", err, runtimeEndpoint)
				humanLog := fmt.Sprintf("服务:%s 的 endpoint 绑定失败，点击查看详情", runtimeEndpoint.RuntimeService.ServiceName)
				detailLog := fmt.Sprintf("endpoint and service info: %+v, error:%s", runtimeEndpoint, errors.Cause(err).Error())
				go common.AsyncRuntimeError(runtimeEndpoint.RuntimeService.RuntimeId, humanLog, detailLog)
			}
		}
	}()

	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	if session != nil {
		_ = session.Rollback()
		session.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayRuntimeServiceServiceImpl) TouchRuntime(c *gin.Context, reqDto *gw.RuntimeServiceReqDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var session *db.SessionHelper
	var runtimeSession db.GatewayRuntimeServiceService
	var endpoints []*orm.GatewayRuntimeService
	package2Endpoint := map[string]*orm.GatewayRuntimeService{}
	var addOrUpdates []runtimeService
	var dels []orm.GatewayRuntimeService
	var daos []orm.GatewayRuntimeService
	var cond *orm.GatewayRuntimeService
	var kongInfo *orm.GatewayKongInfo
	var azInfo *orm.GatewayAzInfo
	var k8sAdapter k8s.K8SAdapter
	var isK8S bool
	var runtimeEndpoints []RuntimeEndpointInfo
	var diceYaml *diceyml.DiceYaml
	var diceObj *diceyml.Object
	apiGatewayError := false
	err := reqDto.CheckValid()
	if err != nil {
		goto failed
	}
	diceYaml, err = bundle.Bundle.GetDiceYAML(reqDto.ReleaseId, reqDto.Env)
	if err != nil {
		goto failed
	}
	if diceYaml == nil {
		err = errors.Errorf("get dice yaml failed, release:%s", reqDto.ReleaseId)
		goto failed
	}
	diceObj = diceYaml.Obj()
	azInfo, err = impl.azDb.GetAzInfoByClusterName(reqDto.ClusterName)
	if err != nil {
		goto failed
	}
	isK8S = azInfo.Type == orm.AT_K8S || azInfo.Type == orm.AT_EDAS
	kongInfo, err = impl.kongDb.GetByAny(&orm.GatewayKongInfo{
		Az: reqDto.ClusterName,
	})
	if err != nil {
		goto failed
	}
	if isK8S {
		k8sAdapter, err = k8s.NewAdapter(reqDto.ClusterName)
		if err != nil {
			goto failed
		}
	}
	session, err = db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	_, err = session.Session().Exec("set innodb_lock_wait_timeout=600")
	if err != nil {
		goto failed
	}
	runtimeSession, err = impl.runtimeDb.NewSession(session)
	if err != nil {
		goto failed
	}
	// get runtime all service, diff req: adds, updates, dels
	cond = &orm.GatewayRuntimeService{
		ProjectId:   reqDto.ProjectId,
		Workspace:   reqDto.Env,
		ClusterName: reqDto.ClusterName,
		AppId:       reqDto.AppId,
		RuntimeName: reqDto.RuntimeName,
	}
	daos, err = runtimeSession.SelectByAny(cond)
	if err != nil {
		goto failed
	}
	addOrUpdates, dels = diffServices(reqDto.Services, daos)

	// create or update runtime service
	for _, service := range addOrUpdates {
		obj := *cond
		obj.RuntimeId = reqDto.RuntimeId
		obj.AppName = reqDto.AppName
		obj.ReleaseId = reqDto.ReleaseId
		obj.GroupNamespace = reqDto.ServiceGroupNamespace
		obj.GroupName = reqDto.ServiceGroupName
		obj.ProjectNamespace = reqDto.ProjectNamespace
		obj.ServiceName = service.dto.ServiceName
		obj.InnerAddress = service.dto.InnerAddress
		obj.UseApigw = 0
		if *reqDto.UseApigw {
			obj.UseApigw = 1
		}
		obj.IsEndpoint = 0
		newCreated := true
		serviceRelease, exist := diceObj.Services[obj.ServiceName]
		if !exist {
			err = errors.Errorf("%s not found in dice.yml",
				obj.ServiceName)
			goto failed
		}
		for _, port := range serviceRelease.Ports {
			if port.Expose {
				obj.IsEndpoint = 1
				obj.BackendProtocol = port.Protocol
				break
			}
		}
		if serviceRelease.TrafficSecurity.Mode != "" {
			obj.IsSecurity = 1
		}
		// create or update
		if service.dao == nil {
			err = runtimeSession.Insert(&obj)
			if err != nil {
				goto failed
			}
		} else {
			obj.Id = service.dao.Id
			if service.dao.InnerAddress != "" {
				newCreated = false
			}
			err = runtimeSession.Update(&obj)
			if err != nil {
				goto failed
			}
			// need clear domains
			if !newCreated && service.dao.IsEndpoint == 1 && obj.IsEndpoint == 0 {
				err = impl.clearDomain(&obj, session)
				if err != nil {
					goto failed
				}
			}
			if obj.IsEndpoint == 1 {
				err = impl.domainBiz.UpdateRuntimeServicePort(&obj, diceObj)
				if err != nil {
					goto failed
				}
				err = impl.domainBiz.RefreshRuntimeDomain(&obj, session)
				if err != nil {
					goto failed
				}
			}
		}
		_ = session.Commit()
		if len(serviceRelease.Endpoints) > 0 {
			runtimeEndpoints = append(runtimeEndpoints, RuntimeEndpointInfo{
				RuntimeService: &obj,
				Endpoints:      serviceRelease.Endpoints,
			})
		}
		_ = session.Begin()
		if kongInfo != nil {
			var namespace string
			namespace, _, err = impl.kongDb.GenK8SInfo(kongInfo)
			if err != nil {
				goto failed
			}
			if isK8S && serviceRelease.MeshEnable != nil && *serviceRelease.MeshEnable {
				err = k8sAdapter.MakeGatewaySupportMesh(namespace)
				if err != nil {
					goto failed
				}
			}
			if isK8S && serviceRelease.TrafficSecurity.Mode == "https" {
				err = k8sAdapter.MakeGatewaySupportHttps(namespace)
				if err != nil {
					goto failed
				}
			}
			err = impl.apiBiz.TouchRuntimeApi(&obj, session, newCreated)
			if err != nil {
				_ = session.Rollback()
				_ = session.Begin()
				log.Errorf("touch runtime api failed, err:%+v", err)
				apiGatewayError = true
			}
		}
		if obj.IsEndpoint == 1 {
			endpoints = append(endpoints, &obj)
		}
	}
	// touch endpoint services' relation package meta
	if !apiGatewayError {
		for _, endpoint := range endpoints {
			var packageId string
			if *reqDto.UseApigw {
				packageId, _, err = impl.packageBiz.TouchRuntimePackageMeta(endpoint, session)
				if err != nil {
					_ = session.Rollback()
					_ = session.Begin()
					log.Errorf("touch runtime package meta failed, err:%+v", err)
					apiGatewayError = true
					break
				}
				package2Endpoint[packageId] = endpoint
			} else {
				// 尝试清除之前依赖API网关自动创建的流量入口
				err = impl.packageBiz.TryClearRuntimePackage(endpoint, session)
				if err != nil {
					_ = session.Rollback()
					_ = session.Begin()
					log.Errorf("try clear runtime package failed, err:%+v", err)
					apiGatewayError = true
					break
				}
			}
		}
	}
	for _, dao := range dels {
		err = impl.clearService(&dao, session)
		if err != nil {
			goto failed
		}
	}
	if *reqDto.UseApigw && !apiGatewayError {
		c.Set("dbsession", session)
		c.Set("endpoints", package2Endpoint)
		c.Set("runtime_endpoints", runtimeEndpoints)
		c.Set("do_async", true)
	} else {
		err = session.Commit()
		if err != nil {
			goto failed
		}
		session.Close()
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	if session != nil {
		_ = session.Rollback()
		session.Close()
	}
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayRuntimeServiceServiceImpl) clearDomain(dao *orm.GatewayRuntimeService, session *db.SessionHelper) error {
	material, err := MakeEndpointMaterial(dao)
	if err != nil {
		return err
	}
	if material.ServiceGroupNamespace == "" || material.ServiceGroupName == "" {
		log.Errorf("invalid material:%+v maybe old, ignored", material)
		return nil
	}
	_, err = impl.domainBiz.TouchRuntimeDomain(dao, material, nil, nil, session)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayRuntimeServiceServiceImpl) clearService(dao *orm.GatewayRuntimeService, session *db.SessionHelper) error {
	runtimeSession, err := impl.runtimeDb.NewSession(session)
	if err != nil {
		return err
	}
	dao.InnerAddress = ""
	err = runtimeSession.Update(dao)
	if err != nil {
		return err
	}
	err = impl.clearDomain(dao, session)
	if err != nil {
		return err
	}
	err = impl.packageBiz.ClearRuntimeRoute(dao.Id)
	if err != nil {
		return err
	}
	err = impl.packageBiz.TryClearRuntimePackage(dao, session)
	if err != nil {
		return err
	}
	err = impl.apiBiz.ClearRuntimeApi(dao)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayRuntimeServiceServiceImpl) DeleteRuntime(runtimeId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if runtimeId == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var daos []orm.GatewayRuntimeService
	// get runtime service id
	daos, err := impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
		RuntimeId: runtimeId,
	})
	if err != nil {
		goto failed
	}
	// clear all inner_addr and runtime domain
	for _, dao := range daos {
		err = impl.clearService(&dao, nil)
		if err != nil {
			goto failed
		}
	}
	return res.SetSuccessAndData(true)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayRuntimeServiceServiceImpl) GetRegisterAppInfo(projectId, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if projectId == "" || env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var resDto gw.RegisterAppsDto
	appMap := map[string]map[string]bool{}
	var runtimeServices []orm.GatewayRuntimeService
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		goto failed
	}
	if az == "" {
		err = errors.New("find cluster failed")
		goto failed
	}
	runtimeServices, err = impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
		ProjectId:   projectId,
		Workspace:   env,
		ClusterName: az,
	})
	if err != nil {
		goto failed
	}
	for _, serviceDao := range runtimeServices {
		if serviceDao.InnerAddress == "" || serviceDao.AppName == "" ||
			serviceDao.ServiceName == "" {
			continue
		}
		appName := serviceDao.AppName
		svcName := serviceDao.ServiceName
		if appName == "" || svcName == "" {
			continue
		}
		serviceMap, exist := appMap[appName]
		if !exist {
			appMap[appName] = map[string]bool{}
			appMap[appName][svcName] = true
			continue
		}
		serviceMap[svcName] = true
	}
	resDto = gw.RegisterAppsDto{
		Apps: []gw.RegisterApp{},
	}
	for appName, services := range appMap {
		app := gw.RegisterApp{
			Name: appName,
		}
		for svcName := range services {
			app.Services = append(app.Services, svcName)
		}
		resDto.Apps = append(resDto.Apps, app)
	}
	return res.SetSuccessAndData(resDto)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayRuntimeServiceServiceImpl) GetServiceRuntimes(projectId, env, app, service string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if projectId == "" || env == "" || app == "" || service == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	var daos []orm.GatewayRuntimeService
	results := []orm.GatewayRuntimeService{}
	az, err := impl.azDb.GetAz(&orm.GatewayAzInfo{
		ProjectId: projectId,
		Env:       env,
	})
	if err != nil {
		goto failed
	}
	if az == "" {
		err = errors.New("find cluster failed")
		goto failed
	}
	daos, err = impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
		ProjectId:   projectId,
		Workspace:   env,
		AppName:     app,
		ServiceName: service,
		ClusterName: az,
	})
	if err != nil {
		goto failed
	}
	for _, dao := range daos {
		if dao.InnerAddress != "" {
			results = append(results, dao)
		}
	}
	return res.SetSuccessAndData(results)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

// 获取指定服务的API前缀
func (impl GatewayRuntimeServiceServiceImpl) GetServiceApiPrefix(req *gw.ApiPrefixReqDto) (res *common.StandardResult) {
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
	dao, err := impl.runtimeDb.GetByAny(&orm.GatewayRuntimeService{
		ProjectId:   req.ProjectId,
		Workspace:   req.Env,
		ServiceName: req.Service,
		AppName:     req.App,
		RuntimeId:   req.RuntimeId,
	})
	if err != nil {
		return
	}
	if dao == nil {
		log.Errorf("runtime service not found, req:%+v", req)
		res.SetSuccessAndData([]string{})
		return
	}
	apis, err := impl.apiBiz.GetRuntimeApis(dao.Id)
	if err != nil {
		return
	}
	pathPrefixMap := map[string]bool{}
	if len(apis) > 0 {
		pathPrefixMap["/"] = true
	}
	for _, api := range apis {
		path := api.Path
		if !strings.Contains(path, "{") {
			pathPrefixMap[path] = true
		}
		splitSlice := strings.SplitN(path, "/", 4)
		switch len(splitSlice) {
		case 0, 1, 2:
		default:
			path = strings.Join(splitSlice[:2], "/")
			if !strings.Contains(path, "{") {
				pathPrefixMap[path] = true
			}
			path = strings.Join(splitSlice[:3], "/")
			if !strings.Contains(path, "{") {
				pathPrefixMap[path] = true
			}
		}
	}
	var prefixs []string
	for pathPrefix := range pathPrefixMap {
		prefixs = append(prefixs, pathPrefix)
	}
	sort.Strings(prefixs)
	res.SetSuccessAndData(prefixs)
	return
}
