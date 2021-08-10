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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	aliyun_errors "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cloudapi"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/hepa/bundle"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	. "github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/config"
	"github.com/erda-project/erda/modules/hepa/endpoint"
	"github.com/erda-project/erda/modules/hepa/endpoint/factories"
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/k8s"
	"github.com/erda-project/erda/modules/hepa/repository/orm"
	db "github.com/erda-project/erda/modules/hepa/repository/service"
	"github.com/erda-project/erda/pkg/crypto/uuid"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

type GatewayDomainServiceImpl struct {
	domainDb  db.GatewayDomainService
	packageDb db.GatewayPackageService
	runtimeDb db.GatewayRuntimeServiceService
	azDb      db.GatewayAzInfoService
	kongDb    db.GatewayKongInfoService
	globalBiz GatewayGlobalService
	ReqCtx    *gin.Context
}

func NewGatewayDomainServiceImpl() (*GatewayDomainServiceImpl, error) {
	domainDb, err := db.NewGatewayDomainServiceImpl()
	if err != nil {
		return nil, err
	}
	packageDb, err := db.NewGatewayPackageServiceImpl()
	if err != nil {
		return nil, err
	}
	runtimeDb, err := db.NewGatewayRuntimeServiceServiceImpl()
	if err != nil {
		return nil, err
	}
	azDb, err := db.NewGatewayAzInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	kongDb, err := db.NewGatewayKongInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	globalBiz, err := NewGatewayGlobalServiceImpl()
	if err != nil {
		return nil, err
	}
	return &GatewayDomainServiceImpl{
		domainDb:  domainDb,
		packageDb: packageDb,
		runtimeDb: runtimeDb,
		azDb:      azDb,
		kongDb:    kongDb,
		globalBiz: globalBiz,
	}, nil
}

func diffDomains(reqDomains []gw.EndpointDomainDto, existDomains []orm.GatewayDomain) (adds []gw.EndpointDomainDto, dels []orm.GatewayDomain, updates []orm.GatewayDomain) {
	for _, domain := range reqDomains {
		exist := false
		for i := len(existDomains) - 1; i >= 0; i-- {
			domainObj := existDomains[i]
			if domain.Domain == domainObj.Domain {
				exist = true
				existDomains = append(existDomains[:i], existDomains[i+1:]...)
				if domain.Type != domainObj.Type {
					domainObj.Type = domain.Type
					updates = append(updates, domainObj)
				}
				break
			}
		}
		if !exist {
			adds = append(adds, domain)
		}
	}
	dels = existDomains
	return
}

func uniqDomains(domains []gw.EndpointDomainDto) []gw.EndpointDomainDto {
	var res []gw.EndpointDomainDto
	markMap := map[string]bool{}
	for _, domainDto := range domains {
		if _, exist := markMap[domainDto.Domain]; !exist {
			markMap[domainDto.Domain] = true
			res = append(res, domainDto)
		}
	}
	return res
}

func (impl GatewayDomainServiceImpl) FindDomains(domain, projectId, workspace string, matchType orm.OptionType, domainType ...string) ([]orm.GatewayDomain, error) {
	var options []orm.SelectOption
	options = append(options, orm.SelectOption{
		Type:   matchType,
		Column: "domain",
		Value:  domain,
	})
	if projectId != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "project_id",
			Value:  projectId,
		})
	}
	if workspace != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "workspace",
			Value:  workspace,
		})
	}
	if len(domainType) > 0 && domainType[0] != "" {
		options = append(options, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "type",
			Value:  domainType[0],
		})
	}
	return impl.domainDb.SelectByOptions(options)
}

func (impl GatewayDomainServiceImpl) acquireComponenetEndpointFactory(clusterName string) (endpoint.EndpointFactory, error) {
	clusterInfo, err := bundle.Bundle.QueryClusterInfo(clusterName)
	if err != nil {
		return nil, err
	}
	clusterType := clusterInfo.Get("DICE_CLUSTER_TYPE")
	switch clusterType {
	case "kubernetes":
		factory, err := factories.NewK8SFactory(clusterName)
		if err != nil {
			return nil, err
		}
		return factory, nil
	case "edas":
		factory, err := factories.NewEdasFactory(clusterName)
		if err != nil {
			return nil, err
		}
		return factory, nil
	default:
		return nil, errors.Errorf("not support cluster type:%s", clusterType)
	}
}

func (impl GatewayDomainServiceImpl) acquireEndpointFactory(runtimeService *orm.GatewayRuntimeService) (bool, endpoint.EndpointFactory, error) {
	azInfo, err := impl.azDb.GetAzInfoByClusterName(runtimeService.ClusterName)
	if azInfo == nil {
		return false, nil, errors.New("az not found")
	}
	if err != nil {
		return true, nil, err
	}
	switch azInfo.Type {
	case orm.AT_K8S:
		factory, err := factories.NewK8SFactory(runtimeService.ClusterName)
		if err != nil {
			return true, nil, err
		}
		return true, factory, nil
	case orm.AT_EDAS:
		factory, err := factories.NewEdasFactory(runtimeService.ClusterName)
		if err != nil {
			return true, nil, err
		}
		return true, factory, nil
	case orm.AT_DCOS:
		fallthrough
	default:
		return false, nil, errors.Errorf("not support cluster type:%s", azInfo.Type)
	}
}

type RuntimeData struct {
	ReleaseId             string `json:"releaseId"`
	ServiceGroupNamespace string `json:"serviceGroupNamespace"`
	ServiceGroupName      string `json:"serviceGroupName"`
}

func (data RuntimeData) checkValid() error {
	if data.ReleaseId == "" || data.ServiceGroupName == "" || data.ServiceGroupNamespace == "" {
		return errors.Errorf("invalid runtimeData:%+v", data)
	}
	return nil
}

type RuntimeResponse struct {
	Success bool        `json:"success"`
	Data    RuntimeData `json:"data"`
}

func (impl GatewayDomainServiceImpl) UpdateRuntimeServicePort(runtimeService *orm.GatewayRuntimeService, releaseInfo *diceyml.Object) error {
	service, exist := releaseInfo.Services[runtimeService.ServiceName]
	if !exist {
		return errors.Errorf("%s not found in dice.yml", runtimeService.ServiceName)
	}
	if len(service.Ports) == 0 {
		return errors.New("ports field not found in dice.yml")
	}
	var exposePort int
	for _, port := range service.Ports {
		if port.Expose {
			exposePort = port.Port
			break
		}
	}
	if exposePort == 0 {
		return errors.New("expose port not found in dice.yml")

	}
	runtimeService.ServicePort = exposePort
	return nil
}

func (impl GatewayDomainServiceImpl) fillRuntimeService(runtimeService *orm.GatewayRuntimeService, orgId, releaseId string, session *db.SessionHelper) error {
	orcheAddr := discover.Orchestrator()
	userId := config.ServerConf.UserId
	if runtimeService.GroupNamespace == "" || runtimeService.GroupName == "" ||
		runtimeService.ReleaseId == "" {
		code, body, err := util.CommonRequest("GET", orcheAddr+"/api/runtimes/"+runtimeService.RuntimeId,
			nil, map[string]string{"org-id": orgId, "user-id": userId})
		if err != nil {
			return err
		}
		if code >= 300 {
			return errors.Errorf("invalid response, code:%d, msg:%s", code, body)
		}
		resp := &RuntimeResponse{}
		err = json.Unmarshal(body, resp)
		if err != nil {
			return err
		}
		if !resp.Success {
			return errors.Errorf("invalid response, body:%s", body)
		}
		runtimeData := resp.Data
		err = runtimeData.checkValid()
		if err != nil {
			return err
		}
		runtimeService.ReleaseId = runtimeData.ReleaseId
		runtimeService.GroupName = runtimeData.ServiceGroupName
		runtimeService.GroupNamespace = runtimeData.ServiceGroupNamespace
	}
	diceYaml, err := bundle.Bundle.GetDiceYAML(releaseId, runtimeService.Workspace)
	if err != nil {
		return err
	}
	if diceYaml == nil {
		return errors.Errorf("get release failed, release:%s", releaseId)
	}
	diceObj := diceYaml.Obj()
	err = impl.UpdateRuntimeServicePort(runtimeService, diceObj)
	if err != nil {
		return err
	}
	runtimeSession, err := impl.runtimeDb.NewSession(session)
	if err != nil {
		return err
	}
	err = runtimeSession.Update(runtimeService)
	if err != nil {
		return err
	}
	return nil
}

func MakeEndpointMaterial(runtimeService *orm.GatewayRuntimeService) (endpoint.EndpointMaterial, error) {
	material := endpoint.EndpointMaterial{}
	material.ServiceName = runtimeService.ServiceName
	material.ServicePort = runtimeService.ServicePort
	material.ServiceGroupNamespace = runtimeService.GroupNamespace
	material.ServiceGroupName = runtimeService.GroupName
	material.ProjectNamespace = runtimeService.ProjectNamespace
	// enable tls by default
	material.K8SRouteOptions.EnableTLS = true
	switch strings.ToLower(runtimeService.BackendProtocol) {
	case "https":
		material.K8SRouteOptions.BackendProtocol = &[]k8s.BackendProtocl{k8s.HTTPS}[0]
	case "grpc":
		material.K8SRouteOptions.BackendProtocol = &[]k8s.BackendProtocl{k8s.GRPC}[0]
	case "grpcs":
		material.K8SRouteOptions.BackendProtocol = &[]k8s.BackendProtocl{k8s.GRPCS}[0]
	case "fastcgi":
		material.K8SRouteOptions.BackendProtocol = &[]k8s.BackendProtocl{k8s.FCGI}[0]
	}
	return material, nil
}

func (impl GatewayDomainServiceImpl) GiveRuntimeDomainToPackage(runtimeService *orm.GatewayRuntimeService, session *db.SessionHelper) (bool, error) {
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return false, err
	}
	serviceDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return false, err
	}
	var domains []gw.EndpointDomainDto
	for _, domain := range serviceDomains {
		domains = append(domains, gw.EndpointDomainDto{
			Domain: domain.Domain,
		})
	}
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return false, err
	}
	relationPackage, err := packageSession.GetByAny(&orm.GatewayPackage{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return false, err
	}
	if relationPackage == nil {
		return false, errors.New("not found")
	}
	projectName, err := impl.globalBiz.GetProjectNameFromCmdb(runtimeService.ProjectId)
	if err != nil {
		return false, err
	}
	changed := false
	if relationPackage != nil {
		packageDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
			PackageId: relationPackage.Id,
		})
		if err != nil {
			return false, err
		}
		packageAdds, _, _ := diffDomains(domains, packageDomains)
		for _, domainDto := range packageAdds {
			err = domainSession.Insert(&orm.GatewayDomain{
				Domain:      domainDto.Domain,
				ClusterName: runtimeService.ClusterName,
				Type:        orm.DT_PACKAGE,
				PackageId:   relationPackage.Id,
				ProjectId:   runtimeService.ProjectId,
				ProjectName: projectName,
				Workspace:   runtimeService.Workspace,
			})
			if err != nil {
				return false, err
			}
		}
		if len(packageAdds) > 0 {
			changed = true
		}

	}
	return changed, nil
}

func (impl GatewayDomainServiceImpl) TouchRuntimeDomain(runtimeService *orm.GatewayRuntimeService, material endpoint.EndpointMaterial, domains []gw.EndpointDomainDto, audits *[]apistructs.Audit, session *db.SessionHelper) (string, error) {
	// unique domain
	domains = uniqDomains(domains)

	var hosts []string
	for _, domain := range domains {
		hosts = append(hosts, domain.Domain)
	}
	err := impl.checkDomainsValid(runtimeService.ClusterName, hosts)
	if err != nil {
		return "", err
	}
	// get runtime service domain
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return "", err
	}
	serviceDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return "", err
	}
	// diff domain: adds, dels, updates
	adds, dels, updates := diffDomains(domains, serviceDomains)

	// get relation package
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return "", err
	}
	relationPackage, err := packageSession.GetByAny(&orm.GatewayPackage{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return "", err
	}
	if relationPackage == nil && runtimeService.UseApigw == 1 {
		packageBiz, err := NewGatewayOpenapiServiceImpl()
		if err != nil {
			return "", err
		}
		id, _, err := packageBiz.TouchRuntimePackageMeta(runtimeService, session)
		if err != nil {
			return "", err
		}
		relationPackage, err = packageSession.Get(id)
		if err != nil {
			return "", err
		}
		if relationPackage == nil {
			return "", errors.New("get relation package failed")
		}
	}
	projectName, err := impl.globalBiz.GetProjectNameFromCmdb(runtimeService.ProjectId)
	if err != nil {
		return "", err
	}
	// add runtime domain,  fail if used by other runtime or package no relation
	for _, domainDto := range adds {
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId:   runtimeService.ProjectId,
			Workspace:   runtimeService.Workspace,
			AppId:       runtimeService.AppId,
			ServiceName: runtimeService.ServiceName,
			RuntimeName: runtimeService.RuntimeName,
		}, apistructs.CreateServiceDomainTemplate, nil, map[string]interface{}{
			"domain": domainDto.Domain,
		})
		if audit != nil && audits != nil {
			*audits = append(*audits, *audit)
		}
		domainList, err := domainSession.SelectByAny(&orm.GatewayDomain{
			Domain:      domainDto.Domain,
			ClusterName: runtimeService.ClusterName,
		})
		if err != nil {
			return "", err
		}
		for _, domainObj := range domainList {
			switch domainObj.Type {
			case orm.DT_SERVICE_DEFAULT, orm.DT_SERVICE_CUSTOM:
				service, err := impl.runtimeDb.Get(domainObj.RuntimeServiceId)
				if err != nil {
					return "", err
				}
				serviceInfo := domainObj.RuntimeServiceId
				if service != nil {
					serviceInfo = fmt.Sprintf("ProjectId:%s, Workspace:%s, AppName:%s, ServiceName:%s, RuntimeName:%s",
						service.ProjectId, service.Workspace, service.AppName, service.ServiceName, service.RuntimeName)
				}
				return domainDto.Domain, errors.Errorf("domain already used by other service, domain:%s, details[%s]",
					domainDto.Domain, serviceInfo)
			case orm.DT_COMPONENT:
				return domainDto.Domain, errors.Errorf("domain already used by dice component, domain:%s, component:%s",
					domainDto.Domain, domainObj.ComponentName)
			case orm.DT_PACKAGE:
				pack, err := impl.packageDb.Get(domainObj.PackageId)
				if err != nil {
					return "", err
				}
				packageInfo := domainObj.PackageId
				if pack != nil {
					packageInfo = fmt.Sprintf("ProjectId:%s, Workspace:%s, EndpointName:%s",
						pack.DiceProjectId, pack.DiceEnv, pack.PackageName)
				}
				if relationPackage == nil || relationPackage.Id != domainObj.PackageId {
					return domainDto.Domain, errors.Errorf("domain already used by api-gateway endpoint, domain:%s, details[%s]",
						domainDto.Domain, packageInfo)
				}
			default:
				return domainDto.Domain, errors.New("domain already used")
			}
		}
		err = domainSession.Insert(&orm.GatewayDomain{
			Domain:           domainDto.Domain,
			ClusterName:      runtimeService.ClusterName,
			Type:             domainDto.Type,
			RuntimeServiceId: runtimeService.Id,
			ProjectId:        runtimeService.ProjectId,
			ProjectName:      projectName,
			Workspace:        runtimeService.Workspace,
		})
		if err != nil {
			return "", err
		}
	}
	az, err := impl.azDb.GetAzInfoByClusterName(runtimeService.ClusterName)
	if err != nil {
		return "", err
	}
	needRelatePackage := true
	if az.Type != orm.AT_K8S && az.Type != orm.AT_EDAS {
		needRelatePackage = false
	}
	// add relation package domain
	if relationPackage != nil && needRelatePackage {
		packageDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
			PackageId: relationPackage.Id,
		})
		if err != nil {
			return "", err
		}
		packageAdds, _, _ := diffDomains(domains, packageDomains)
		// 跟随runtime的域名增加
		for _, domainDto := range packageAdds {
			err = domainSession.Insert(&orm.GatewayDomain{
				Domain:      domainDto.Domain,
				ClusterName: runtimeService.ClusterName,
				Type:        orm.DT_PACKAGE,
				PackageId:   relationPackage.Id,
				ProjectId:   runtimeService.ProjectId,
				ProjectName: projectName,
				Workspace:   runtimeService.Workspace,
			})
			if err != nil {
				return "", err
			}
		}
		var packageDomainDeleted bool
		// 跟随runtime的域名删除
		for _, delObj := range dels {
			changed, err := domainSession.DeleteByAny(&orm.GatewayDomain{
				Domain:      delObj.Domain,
				ClusterName: runtimeService.ClusterName,
				Type:        orm.DT_PACKAGE,
				PackageId:   relationPackage.Id,
			})
			if err != nil {
				return "", err
			}
			if changed != 0 {
				packageDomainDeleted = true
			}
		}
		if len(packageAdds) > 0 || packageDomainDeleted {
			packageBiz, err := NewGatewayOpenapiServiceImpl()
			if err != nil {
				return "", err
			}
			err = packageBiz.RefreshRuntimePackage(relationPackage.Id, runtimeService, session)
			if err != nil {
				return "", err
			}
		}
	}

	for _, domainObj := range updates {
		err = domainSession.Update(&domainObj)
		if err != nil {
			return "", err
		}
	}

	// del runtime domain
	for _, domainObj := range dels {
		audit := common.MakeAuditInfo(impl.ReqCtx, common.ScopeInfo{
			ProjectId:   runtimeService.ProjectId,
			Workspace:   runtimeService.Workspace,
			AppId:       runtimeService.AppId,
			ServiceName: runtimeService.ServiceName,
			RuntimeName: runtimeService.RuntimeName,
		}, apistructs.DeleteServiceDomainTemplate, nil, map[string]interface{}{
			"domain": domainObj.Domain,
		})
		if audit != nil && audits != nil {
			*audits = append(*audits, *audit)
		}
		err = domainSession.Delete(domainObj.Id)
		if err != nil {
			return "", err
		}
	}

	for _, domain := range domains {
		material.Routes = append(material.Routes, endpoint.Route{
			Host: domain.Domain,
			Path: "/",
		})
	}
	needTouchEndpoint, factory, err := impl.acquireEndpointFactory(runtimeService)
	if needTouchEndpoint {
		if err != nil {
			return "", err
		}
		err = endpoint.TouchEndpoint(factory, material)
		if err != nil {
			return "", err
		}
	}
	return "", nil
}

func (impl GatewayDomainServiceImpl) RefreshRuntimeDomain(runtimeService *orm.GatewayRuntimeService, session *db.SessionHelper) error {
	material, err := MakeEndpointMaterial(runtimeService)
	if err != nil {
		return err
	}
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return err
	}
	domains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		RuntimeServiceId: runtimeService.Id,
	})
	if err != nil {
		return err
	}
	for _, domain := range domains {
		material.Routes = append(material.Routes, endpoint.Route{
			Host: domain.Domain,
			Path: "/",
		})
	}
	needTouchEndpoint, factory, err := impl.acquireEndpointFactory(runtimeService)
	if needTouchEndpoint {
		if err != nil {
			return err
		}
		err = endpoint.TouchEndpoint(factory, material)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl GatewayDomainServiceImpl) IsPackageDomainsDiff(packageId, clusterName string, domains []string, session *db.SessionHelper) (bool, error) {
	// unique domain
	domains = util.UniqStringSlice(domains)
	var domainDtos []gw.EndpointDomainDto
	for _, domain := range domains {
		domainDtos = append(domainDtos, gw.EndpointDomainDto{Domain: domain})
	}
	// get package domain
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return false, err
	}
	packageDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		PackageId: packageId,
	})
	if err != nil {
		return false, err
	}
	// diff domain: adds, dels
	adds, dels, _ := diffDomains(domainDtos, packageDomains)
	log.Debugf("package domains, adds:%+v, dels:%+v, packageId:%s, req domain: %+v, domainDto: %+v, packageDomains: %+v", adds, dels, packageId, domains, domainDtos, packageDomains) // output for debug
	return len(adds) != 0 || len(dels) != 0, nil
}

func (impl GatewayDomainServiceImpl) SetCloudapiDomain(pack *orm.GatewayPackage, domains []string) error {
	if pack.CloudapiGroupId == "" {
		return nil
	}
	orgId, err := impl.globalBiz.GetOrgId(pack.DiceProjectId)
	if err != nil {
		return err
	}
	for _, domain := range domains {
		descDomainReq := cloudapi.CreateDescribeDomainRequest()
		descDomainReq.GroupId = pack.CloudapiGroupId
		descDomainReq.DomainName = domain
		descDomainReq.SecurityToken = uuid.UUID()
		descDomainResp := cloudapi.CreateDescribeDomainResponse()
		err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, descDomainReq, descDomainResp)
		var domainNotFound bool
		if err != nil {
			rootError := errors.Cause(err)
			if serverErr, ok := rootError.(*aliyun_errors.ServerError); ok {
				if serverErr.HttpStatus() != 404 {
					return err
				}
				domainNotFound = true
			} else {
				return err
			}
		}
		if domainNotFound {
			setDomainReq := cloudapi.CreateSetDomainRequest()
			setDomainReq.GroupId = pack.CloudapiGroupId
			setDomainReq.DomainName = domain
			setDomainReq.SecurityToken = uuid.UUID()
			setDomainResp := cloudapi.CreateSetDomainResponse()
			err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, setDomainReq, setDomainResp)
			if err != nil {
				rootError := errors.Cause(err)
				if clientErr, ok := rootError.(*aliyun_errors.ClientError); ok {
					if clientErr.ErrorCode() == "InvalidDomainName" {
						log.Errorf("invalid domain cannot set to aliyun gateway, pack:%+v, domain:%s", pack, domain)
						continue
					}
				}
				if serverErr, ok := rootError.(*aliyun_errors.ServerError); ok {
					if serverErr.ErrorCode() == "InvalidICPLicense" {
						log.Errorf("invalid icp license cannot set to aliyun gateway, pack:%+v, domain:%s", pack, domain)
						continue
					}
				}
				return err
			}
		}
	}
	return nil
}

func (impl GatewayDomainServiceImpl) deleteCloudapiDomain(pack *orm.GatewayPackage, domain string) error {
	if pack.CloudapiGroupId == "" {
		return nil
	}
	orgId, err := impl.globalBiz.GetOrgId(pack.DiceProjectId)
	if err != nil {
		return err
	}
	descDomainReq := cloudapi.CreateDescribeDomainRequest()
	descDomainReq.GroupId = pack.CloudapiGroupId
	descDomainReq.DomainName = domain
	descDomainReq.SecurityToken = uuid.UUID()
	descDomainResp := cloudapi.CreateDescribeDomainResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, descDomainReq, descDomainResp)
	if err != nil {
		rootError := errors.Cause(err)
		if serverErr, ok := rootError.(*aliyun_errors.ServerError); ok {
			if serverErr.HttpStatus() != 404 {
				return err
			}
			// 不存在域名
			return nil
		} else {
			return err
		}
	}
	req := cloudapi.CreateDeleteDomainRequest()
	req.GroupId = pack.CloudapiGroupId
	req.DomainName = domain
	req.SecurityToken = uuid.UUID()
	resp := cloudapi.CreateDeleteDomainResponse()
	err = bundle.Bundle.DoRemoteAliyunAction(orgId, pack.DiceClusterName, vars.CloudapiEndpointType, vars.CloudapiEndpointMap, req, resp)
	if err != nil {
		return err
	}
	return nil
}

func (impl GatewayDomainServiceImpl) checkDomainsValid(clusterName string, domains []string) error {
	clusterInfo, err := bundle.Bundle.QueryClusterInfo(clusterName)
	if err != nil {
		return errors.WithStack(err)
	}
	isEdge := clusterInfo.Get(apistructs.DICE_IS_EDGE)
	var keepList []string
	if isEdge == "true" {
		keepList = config.ServerConf.EdgeDomainNameKeepList
	} else {
		keepList = config.ServerConf.CenterDomainNameKeepList
		for _, domain := range domains {
			if strings.HasSuffix(strings.Split(domain, ".")[0], "-org") {
				return errors.Errorf("invalid domain:%s, it's maintained by platform", domain)
			}
		}
	}
	clusterDomain := clusterInfo.Get(apistructs.DICE_ROOT_DOMAIN)
	for _, domain := range domains {
		if !strings.HasSuffix(domain, clusterDomain) {
			return nil
		}
		for _, keepDomain := range keepList {
			if strings.HasPrefix(domain, keepDomain+".") {
				return errors.Errorf("invalid domain:%s, it's maintained by platform", domain)
			}
		}
	}
	return nil
}

func (impl GatewayDomainServiceImpl) TouchPackageDomain(packageId, clusterName string, domains []string,
	session *db.SessionHelper) ([]string, error) {
	// unique domain
	domains = util.UniqStringSlice(domains)
	err := impl.checkDomainsValid(clusterName, domains)
	if err != nil {
		return nil, err
	}
	var domainDtos []gw.EndpointDomainDto
	for _, domain := range domains {
		domainDtos = append(domainDtos, gw.EndpointDomainDto{Domain: domain})
	}
	// get package domain
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return nil, err
	}
	packageDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		PackageId: packageId,
	})
	if err != nil {
		return nil, err
	}
	packageSession, err := impl.packageDb.NewSession(session)
	if err != nil {
		return nil, err
	}
	pack, err := packageSession.Get(packageId)
	if err != nil {
		return nil, err
	}
	if pack == nil {
		return nil, errors.New("invalid package")
	}
	projectName, err := impl.globalBiz.GetProjectNameFromCmdb(pack.DiceProjectId)
	if err != nil {
		return nil, err
	}
	go func() {
		defer util.DoRecover()
		err = impl.SetCloudapiDomain(pack, domains)
		if err != nil {
			log.Errorf("error happend: %+v", err)
		}
	}()
	// diff domain: adds, dels
	adds, dels, _ := diffDomains(domainDtos, packageDomains)
	// add package domain, fail if used by other package or runtime no relation
	for _, domainDto := range adds {
		domainList, err := domainSession.SelectByAny(&orm.GatewayDomain{
			Domain:      domainDto.Domain,
			ClusterName: clusterName,
		})
		if err != nil {
			return nil, err
		}
		if pack.Scene != orm.UNITY_SCENE {

			for _, domainObj := range domainList {
				switch domainObj.Type {
				case orm.DT_SERVICE_DEFAULT, orm.DT_SERVICE_CUSTOM:
					if domainObj.RuntimeServiceId != pack.RuntimeServiceId {
						return nil, errors.Errorf("domain already used by other service, domain:%s, service:%s",
							domainDto.Domain, domainObj.RuntimeServiceId)
					}
				case orm.DT_COMPONENT:
					return nil, errors.Errorf("domain already used by dice component, domain:%s, component:%s",
						domainDto.Domain, domainObj.ComponentName)
				case orm.DT_PACKAGE:
					return nil, errors.Errorf("domain already used by other api-gateway endpoints, domain:%s, endpoints:%s",
						domainDto.Domain, domainObj.PackageId)

				default:
					return nil, errors.New("domain already used")
				}
			}
		}
		err = domainSession.Insert(&orm.GatewayDomain{
			Domain:      domainDto.Domain,
			ClusterName: clusterName,
			Type:        orm.DT_PACKAGE,
			PackageId:   packageId,
			ProjectId:   pack.DiceProjectId,
			Workspace:   pack.DiceEnv,
			ProjectName: projectName,
		})
		if err != nil {
			return nil, err
		}

	}
	// del package domain
	for _, domainObj := range dels {
		err = domainSession.Delete(domainObj.Id)
		if err != nil {
			return nil, err
		}
	}
	go func() {
		defer util.DoRecover()
		for _, domainObj := range dels {
			err = impl.deleteCloudapiDomain(pack, domainObj.Domain)
			if err != nil {
				log.Errorf("error happend: %+v", err)
			}
		}
	}()

	return domains, nil
}

func (impl GatewayDomainServiceImpl) GetPackageDomains(packageId string, session ...*db.SessionHelper) ([]string, error) {
	var domainSession db.GatewayDomainService
	var err error
	if len(session) > 0 {
		domainSession, err = impl.domainDb.NewSession(session[0])
		if err != nil {
			return nil, err
		}
	} else {
		domainSession = impl.domainDb
	}

	packageDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		PackageId: packageId,
	})
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, domainDao := range packageDomains {
		res = append(res, domainDao.Domain)
	}
	return res, nil
}

func (impl GatewayDomainServiceImpl) GetTenantDomains(projectId, env string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	if projectId == "" || env == "" {
		return res.SetReturnCode(PARAMS_IS_NULL)
	}
	domainMap := map[string]bool{}
	result := []string{}
	var packages []orm.GatewayPackage
	var runtimes []orm.GatewayRuntimeService
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
	runtimes, err = impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
		ProjectId:   projectId,
		Workspace:   env,
		ClusterName: az,
	})
	if err != nil {
		goto failed
	}
	packages, err = impl.packageDb.SelectByAny(&orm.GatewayPackage{
		DiceProjectId:   projectId,
		DiceEnv:         env,
		DiceClusterName: az,
	})
	if err != nil {
		goto failed
	}
	for _, runtime := range runtimes {
		domains, err := impl.domainDb.SelectByAny(&orm.GatewayDomain{
			RuntimeServiceId: runtime.Id,
		})
		if err != nil {
			goto failed
		}
		for _, domain := range domains {
			if _, exist := domainMap[domain.Domain]; !exist {
				domainMap[domain.Domain] = true
				result = append(result, domain.Domain)
			}
		}
	}
	for _, pack := range packages {
		domains, err := impl.domainDb.SelectByAny(&orm.GatewayDomain{
			PackageId: pack.Id,
		})
		if err != nil {
			goto failed
		}
		for _, domain := range domains {
			if _, exist := domainMap[domain.Domain]; !exist {
				domainMap[domain.Domain] = true
				result = append(result, domain.Domain)
			}
		}
	}
	return res.SetSuccessAndData(result)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})

}

func (impl GatewayDomainServiceImpl) doesClusterSupportHttps(clusterName string) bool {
	clusterInfo, err := bundle.Bundle.QueryClusterInfo(clusterName)
	if err != nil {
		log.Errorf("err happened, err: %+v", errors.WithStack(err))
		return false
	}
	if clusterInfo.DiceProtocolIsHTTPS() {
		return true
	}
	return false
}

func (impl GatewayDomainServiceImpl) GetRuntimeDomains(runtimeId string) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	result := gw.RuntimeDomainsDto{}
	var rootDomain, appName, tenantGroup string
	var useHttps bool
	runtimeServices, err := impl.runtimeDb.SelectByAny(&orm.GatewayRuntimeService{
		RuntimeId:  runtimeId,
		IsEndpoint: 1,
	})
	if err != nil {
		goto failed
	}
	if len(runtimeServices) == 0 {
		log.Errorf("runtime id:%s maybe not ready", runtimeId)
		return res.SetSuccessAndData(result)
	}
	// acquire root domain
	{
		leader := runtimeServices[0]
		appName = leader.AppName
		var azInfo *orm.GatewayAzInfo
		var kongInfo *orm.GatewayKongInfo
		azInfo, err = impl.azDb.GetAzInfoByClusterName(leader.ClusterName)
		if err != nil {
			goto failed
		}
		if azInfo == nil {
			err = errors.New("cluster info not found")
			goto failed
		}
		rootDomain = azInfo.WildcardDomain
		kongInfo, _ = impl.kongDb.GetKongInfo(&orm.GatewayKongInfo{
			ProjectId: leader.ProjectId,
			Env:       leader.Workspace,
			Az:        leader.ClusterName,
		})
		if kongInfo != nil {
			tenantGroup = kongInfo.TenantGroup
		}
		useHttps = impl.doesClusterSupportHttps(leader.ClusterName)
	}
	for _, service := range runtimeServices {
		var domains []orm.GatewayDomain
		var relationPackage *orm.GatewayPackage
		// 内部地址为空的service已经下掉了，不要显示域名
		if service.InnerAddress == "" {
			continue
		}
		domains, err = impl.domainDb.SelectByAny(&orm.GatewayDomain{
			RuntimeServiceId: service.Id,
		})
		relationPackageId := ""
		if err != nil {
			goto failed
		}
		relationPackage, err = impl.packageDb.GetByAny(&orm.GatewayPackage{
			RuntimeServiceId: service.Id,
		})
		if err != nil {
			goto failed
		}
		if relationPackage != nil {
			relationPackageId = relationPackage.Id
		}
		var domainList gw.SortByTypeList
		for _, domain := range domains {
			runtimeDomain := gw.RuntimeDomain{
				Domain:      domain.Domain,
				DomainType:  gw.EDT_CUSTOM,
				AppName:     appName,
				PackageId:   relationPackageId,
				TenantGroup: tenantGroup,
				UseHttps:    useHttps,
			}
			switch domain.Type {
			case orm.DT_SERVICE_DEFAULT:
				if strings.HasSuffix(runtimeDomain.Domain, rootDomain) {
					runtimeDomain.RootDomain = "." + rootDomain
					runtimeDomain.CustomDomain = strings.TrimSuffix(runtimeDomain.Domain,
						runtimeDomain.RootDomain)
					runtimeDomain.DomainType = gw.EDT_DEFAULT
				}
			case orm.DT_SERVICE_CUSTOM:
			default:
				err = errors.Errorf("invalid domain: %+v", domain)
				goto failed
			}
			domainList = append(domainList, runtimeDomain)
		}
		sort.Sort(domainList)
		if len(domainList) == 0 || domainList[0].DomainType != gw.EDT_DEFAULT {
			domainList = append([]gw.RuntimeDomain{
				{
					DomainType: gw.EDT_DEFAULT,
					RootDomain: "." + rootDomain,
				},
			}, domainList...)
		}
		result[service.ServiceName] = domainList
	}
	return res.SetSuccessAndData(result)
failed:
	log.Errorf("error happened, err:%+v", err)
	return res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
}

func (impl GatewayDomainServiceImpl) UpdateRuntimeServiceDomain(orgId, runtimeId, serviceName string, reqDto *gw.ServiceDomainReqDto) *common.StandardResult {
	res := &common.StandardResult{Success: false}
	var err error
	var material endpoint.EndpointMaterial
	var runtimeService *orm.GatewayRuntimeService
	var domains []gw.EndpointDomainDto
	var existDomain, rootDomain string
	audits := []apistructs.Audit{}
	defer func() {
		if len(audits) == 0 {
			return
		}
		var errMsg string
		var result apistructs.Result
		if res.Success {
			result = apistructs.SuccessfulResult
		} else {
			result = apistructs.FailureResult
		}
		if res.Err != nil {
			errMsg = res.Err.Msg
		}
		for i := 0; i < len(audits); i++ {
			audits[i].Result = result
			audits[i].ErrorMsg = errMsg
		}
		err = bundle.Bundle.BatchCreateAuditEvent(&apistructs.AuditBatchCreateRequest{Audits: audits})

		if err != nil {
			log.Errorf("create batch audit failed, err:%+v", err)
		}
	}()
	session, err := db.NewSessionHelper()
	if err != nil {
		goto failed
	}
	if orgId == "" || runtimeId == "" || serviceName == "" {
		err = errors.Errorf("invalid arg, orgId:%s, runtimeId:%s, serviceName:%s", orgId, runtimeId, serviceName)
		goto failed
	}
	err = reqDto.CheckValid()
	if err != nil {
		goto failed
	}
	runtimeService, err = impl.runtimeDb.GetByAny(&orm.GatewayRuntimeService{
		RuntimeId:   runtimeId,
		ServiceName: serviceName,
		IsEndpoint:  1,
	})
	if err != nil {
		goto failed
	}
	if runtimeService == nil {
		err = errors.Errorf("endpoint service %s may not ready", serviceName)
		goto failed
	}
	// acquire root domain
	{
		var azInfo *orm.GatewayAzInfo
		azInfo, err = impl.azDb.GetAzInfoByClusterName(runtimeService.ClusterName)
		if err != nil {
			goto failed
		}
		if azInfo == nil {
			err = errors.New("cluster info not found")
			goto failed
		}
		rootDomain = azInfo.WildcardDomain
	}
	for i := 0; i < len(reqDto.Domains); i++ {
		domain := gw.EndpointDomainDto{
			Domain: reqDto.Domains[i],
			Type:   orm.DT_SERVICE_CUSTOM,
		}
		if i == 0 && strings.HasSuffix(reqDto.Domains[i], "."+rootDomain) {
			domain.Type = orm.DT_SERVICE_DEFAULT
		}
		domains = append(domains, domain)
	}
	err = impl.fillRuntimeService(runtimeService, orgId, reqDto.ReleaseId, session)
	if err != nil {
		goto failed
	}
	material, err = MakeEndpointMaterial(runtimeService)
	if err != nil {
		goto failed
	}
	existDomain, err = impl.TouchRuntimeDomain(runtimeService, material, domains, &audits, session)
	if err != nil {
		goto failed
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
	if existDomain == "" {
		res.SetErrorInfo(&common.ErrInfo{Msg: errors.Cause(err).Error()})
	} else {
		res.SetErrorInfo(&common.ErrInfo{
			Msg:   "域名已被占用: " + existDomain,
			EnMsg: "Domain name is already taken: " + existDomain,
		})
	}
	return res
}

func (impl GatewayDomainServiceImpl) CreateOrUpdateComponentIngress(req apistructs.ComponentIngressUpdateRequest) (res *common.StandardResult) {
	var err error
	var session *db.SessionHelper
	res = &common.StandardResult{Success: false}
	defer func() {
		if err != nil {
			log.Errorf("error happened: %+v", err)
			res.SetErrorInfo(&common.ErrInfo{
				Msg: errors.Cause(err).Error(),
			})
			if session != nil {
				_ = session.Rollback()
				session.Close()
			}
			return
		}
		if session != nil {
			_ = session.Commit()
			session.Close()
		}
	}()
	err = req.CheckValid()
	if err != nil {
		return
	}
	session, err = db.NewSessionHelper()
	if err != nil {
		return
	}
	var domains []gw.EndpointDomainDto
	var hosts []string
	for _, route := range req.Routes {
		domains = append(domains, gw.EndpointDomainDto{Domain: route.Domain})
		hosts = append(hosts, route.Domain)
	}
	err = impl.checkDomainsValid(req.ClusterName, hosts)
	if err != nil {
		return
	}
	domains = uniqDomains(domains)
	domainSession, err := impl.domainDb.NewSession(session)
	if err != nil {
		return
	}
	existDomains, err := domainSession.SelectByAny(&orm.GatewayDomain{
		ComponentName: req.ComponentName,
		IngressName:   req.IngressName,
		ClusterName:   req.ClusterName,
	})
	if err != nil {
		return
	}
	adds, dels, _ := diffDomains(domains, existDomains)
	for _, domainDto := range adds {
		domainList, err := domainSession.SelectByAny(&orm.GatewayDomain{
			Domain:      domainDto.Domain,
			ClusterName: req.ClusterName,
		})
		if err != nil {
			return
		}
		for _, domainObj := range domainList {
			if domainObj.ComponentName != req.ComponentName && domainObj.IngressName != req.IngressName {
				log.Errorf("domain already used, domainObj:%+v", domainObj)
				return
			}
		}
		err = domainSession.Insert(&orm.GatewayDomain{
			Domain:        domainDto.Domain,
			ClusterName:   req.ClusterName,
			Type:          orm.DT_COMPONENT,
			ComponentName: req.ComponentName,
			IngressName:   req.IngressName,
		})
		if err != nil {
			return
		}
	}
	for _, domainObj := range dels {
		err = domainSession.Delete(domainObj.Id)
		if err != nil {
			return
		}
	}
	material := endpoint.EndpointMaterial{
		K8SNamespace: req.K8SNamespace,
		ServiceName:  req.ComponentName,
		ServicePort:  req.ComponentPort,
		IngressName:  req.IngressName,
	}
	var routes []endpoint.Route
	for _, route := range req.Routes {
		routes = append(routes, endpoint.Route{
			Host: route.Domain,
			Path: route.Path,
		})
	}
	material.Routes = routes
	routeOptions := k8s.RouteOptions{
		RewriteHost:     req.RouteOptions.RewriteHost,
		RewritePath:     req.RouteOptions.RewritePath,
		UseRegex:        req.RouteOptions.UseRegex,
		EnableTLS:       *req.RouteOptions.EnableTLS,
		LocationSnippet: req.RouteOptions.LocationSnippet,
	}
	annotations := map[string]*string{}
	// disable log by default
	annotations["nginx.ingress.kubernetes.io/enable-access-log"] = &[]string{"false"}[0]
	for key, value := range req.RouteOptions.Annotations {
		annotations[key] = &value
	}
	routeOptions.Annotations = annotations
	material.K8SRouteOptions = routeOptions
	factory, err := impl.acquireComponenetEndpointFactory(req.ClusterName)
	if err != nil {
		_ = session.Rollback()
		session.Close()
		log.Errorf("acquireComponenetEndpointFactory failed, err:%+v", err)
	}
	if factory == nil {
		res.SetSuccessAndData(true)
		return
	}

	err = endpoint.TouchComponentEndpoint(factory, material)
	if err != nil {
		return
	}
	res.SetSuccessAndData(true)
	return
}

func (impl GatewayDomainServiceImpl) GetOrgDomainInfo(dice gw.DiceArgsDto, reqDto *gw.ManageDomainReq) (res *common.StandardResult) {
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
	if dice.OrgId == "" || (reqDto.Type != "" && reqDto.Type != gw.ServiceDomain && reqDto.Type != gw.GatewayDomain && reqDto.Type != gw.OtherDomain) {
		err = errors.New("invalid request")
		return
	}
	clusters, err := impl.globalBiz.GetClustersByOrg(dice.OrgId)
	if err != nil {
		return
	}
	if len(clusters) == 0 {
		res.SetSuccessAndData(common.NewPages(nil, 0))
		return
	}
	pageInfo := common.NewPage2(dice.PageSize, dice.PageNo)
	reqDto.ProjectID = dice.ProjectId
	reqDto.Workspace = dice.Env
	selectOptions := reqDto.GenSelectOptions()
	if reqDto.ClusterName == "" {
		selectOptions = append(selectOptions, orm.SelectOption{
			Type:   orm.Contains,
			Column: "cluster_name",
			Value:  clusters,
		})
	} else {
		clusterAllow := false
		for _, cluster := range clusters {
			if strings.EqualFold(cluster, reqDto.ClusterName) {
				clusterAllow = true
				break
			}
		}
		if !clusterAllow {
			err = errors.New("invalid cluster name")
			return
		}
		selectOptions = append(selectOptions, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "cluster_name",
			Value:  reqDto.ClusterName,
		})
	}
	page, err := impl.domainDb.GetPage(selectOptions, pageInfo)
	if err != nil {
		return
	}
	daos, ok := page.Result.([]orm.GatewayDomain)
	if !ok {
		err = errors.New("type convert failed")
		return
	}
	var list []gw.ManageDomainInfo
	for _, dao := range daos {
		dto := gw.ManageDomainInfo{
			ID:          dao.Id,
			Domain:      dao.Domain,
			ClusterName: dao.ClusterName,
			ProjectName: dao.ProjectName,
			Workspace:   dao.Workspace,
		}
		link := &gw.DomainLinkInfo{
			ProjectID: dao.ProjectId,
			Workspace: dao.Workspace,
		}
		switch dao.Type {
		case orm.DT_SERVICE_CUSTOM, orm.DT_SERVICE_DEFAULT:
			dto.Type = gw.ServiceDomain
			runtimeService, err := impl.runtimeDb.Get(dao.RuntimeServiceId)
			if err != nil {
				log.Errorf("get runtime service failed, err:%+v", err)
				continue
			}
			if runtimeService == nil {
				log.Errorf("runtime service not found")
				continue
			}
			dto.AppName = runtimeService.AppName
			link.AppID = runtimeService.AppId
			link.RuntimeID = runtimeService.RuntimeId
			link.ServiceName = runtimeService.ServiceName
			dto.Link = link
		case orm.DT_PACKAGE:
			dto.Type = gw.GatewayDomain
			link.TenantGroup = impl.globalBiz.GenTenantGroup(dao.ProjectId, dao.Workspace, dao.ClusterName)
			dto.Link = link
		default:
			dto.Type = gw.OtherDomain
		}
		list = append(list, dto)
	}
	res.SetSuccessAndData(common.NewPages(list, pageInfo.TotalNum))
	return
}
