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

package assetsvc

import (
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/dop/dbclient"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/swagger"
	"github.com/erda-project/erda/pkg/swagger/oas3"
	"github.com/erda-project/erda/pkg/swagger/oasconv"
)

// GetAsset 查询 API 资料
func (svc *Service) GetAsset(req *apistructs.GetAPIAssetReq) (*apistructs.GetAPIAssetResponse, error) {
	// 参数校验
	if req.OrgID == 0 {
		return nil, apierrors.GetAPIAsset.MissingParameter(apierrors.MissingOrgID)
	}
	if req.URIParams == nil {
		return nil, apierrors.GetAPIAsset.MissingParameter("no uri parameters")
	}
	if err := apistructs.ValidateAPIAssetID(req.URIParams.AssetID); err != nil {
		return nil, apierrors.GetAPIAsset.InvalidParameter(fmt.Errorf("assetID: %v", err))
	}

	// 查询
	assetModel, err := dbclient.GetAPIAsset(req)
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.GetAPIAsset.NotFound()
		}
		return nil, apierrors.GetAPIAsset.InternalError(err)
	}

	asset := apistructs.APIAssetsModel(*assetModel)

	// 每次查询后, 都同步下 project_name 和 app_name
	go func() {
		if asset.ProjectID == nil {
			return
		}
		if *asset.ProjectID == 0 {
			return
		}
		project, err2 := bdl.Bdl.GetProject(*asset.ProjectID)
		if err2 != nil {
			return
		}
		if asset.ProjectName == nil || *asset.ProjectName != project.Name {
			_ = dbclient.Sq().Model(new(apistructs.APIAssetsModel)).
				Where(map[string]interface{}{"org_id": asset.OrgID, "asset_id": asset.AssetID}).
				Updates(map[string]interface{}{"project_name": project.Name}).Error
		}
	}()

	go func() {
		if asset.AppID == nil {
			return
		}
		if *asset.AppID == 0 {
			return
		}
		app, err2 := bdl.Bdl.GetApp(*asset.AppID)
		if err2 != nil {
			return
		}
		if asset.AppName == nil || *asset.AppName != app.Name {
			_ = dbclient.Sq().Model(new(apistructs.APIAssetsModel)).
				Where(map[string]interface{}{"org_id": asset.OrgID, "asset_id": asset.AssetID}).
				Updates(map[string]interface{}{"app_name": app.Name})
		}
	}()

	// 按钮权限
	permission := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	written := writePermission(permission, &asset)

	hasAccess := svc.assetHasAccess(req.OrgID, asset.AssetID)
	hasInstantiation := svc.FirstRecord(new(apistructs.InstantiationModel), map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": req.URIParams.AssetID,
	}) == nil
	return &apistructs.GetAPIAssetResponse{
		Asset: &asset,
		Permission: map[string]bool{
			"delete":           written,
			"public":           written,
			"edit":             written,
			"request":          hasAccess && hasInstantiation,
			"hasAccess":        hasAccess,
			"hasInstantiation": hasInstantiation,
		},
	}, nil
}

// GetAssetVersion 查询 API 资料版本
func (svc *Service) GetAssetVersion(req *apistructs.GetAPIAssetVersionReq) (*apistructs.GetAssetVersionRsp, error) {
	// 参数校验
	if req.OrgID == 0 {
		return nil, apierrors.GetAPIAssetVersion.MissingParameter(apierrors.MissingOrgID)
	}
	if err := apistructs.ValidateAPIAssetID(req.URIParams.AssetID); err != nil {
		return nil, apierrors.GetAPIAsset.InvalidParameter(fmt.Errorf("assetID: %v", err))
	}

	var response apistructs.GetAssetVersionRsp

	// 查询
	version, err := dbclient.GetAPIAssetVersion(req)
	if err != nil {
		logrus.Errorf("failed to GetAPIAssetVersion, req: %+v: %v", req, err)
		if gorm.IsRecordNotFoundError(err) {
			return nil, apierrors.GetAPIAssetVersion.NotFound()
		}
		return nil, apierrors.GetAPIAssetVersion.InternalError(err)
	}
	response.Version = version

	if req.QueryParams.Asset {
		asset, err := dbclient.GetAPIAsset(&apistructs.GetAPIAssetReq{
			OrgID:     req.OrgID,
			Identity:  req.Identity,
			URIParams: &apistructs.GetAPIAssetURIPrams{AssetID: req.URIParams.AssetID},
		})
		if err != nil {
			return nil, apierrors.GetAPIAssetVersion.InternalError(err)
		}
		if asset.ProjectID != nil && *asset.ProjectID != 0 {
			if project, err := bdl.Bdl.GetProject(*asset.ProjectID); err == nil {
				asset.ProjectName = &project.Name
			}
		}
		if asset.AppID != nil && *asset.AppID != 0 {
			if app, err := bdl.Bdl.GetProject(*asset.AppID); err == nil {
				asset.AppName = &app.Name
			}
		}
		response.Asset = &apistructs.APIAssetsModel{
			BaseModel: apistructs.BaseModel{
				ID:        asset.ID,
				CreatedAt: asset.CreatedAt,
				UpdatedAt: asset.UpdatedAt,
				CreatorID: asset.CreatorID,
				UpdaterID: asset.UpdaterID,
			},
			OrgID:        asset.OrgID,
			AssetID:      asset.AssetID,
			AssetName:    asset.AssetName,
			Desc:         asset.Desc,
			Logo:         asset.Logo,
			ProjectID:    asset.ProjectID,
			ProjectName:  asset.ProjectName,
			AppID:        asset.AppID,
			AppName:      asset.AppName,
			Public:       asset.Public,
			CurVersionID: asset.CurVersionID,
			CurMajor:     asset.CurMajor,
			CurMinor:     asset.CurMinor,
			CurPatch:     asset.CurPatch,
		}
	}

	if req.QueryParams.Spec {
		spec, err := dbclient.GetAPIAssetVersionSpec(req)
		if err != nil {
			return nil, apierrors.GetAPIAssetVersion.InternalError(err)
		}

		response.Spec = &apistructs.APIAssetVersionSpecsModel{
			BaseModel: apistructs.BaseModel{
				ID:        spec.ID,
				CreatedAt: spec.CreatedAt,
				UpdatedAt: spec.UpdatedAt,
				CreatorID: spec.CreatorID,
				UpdaterID: spec.UpdaterID,
			},
			OrgID:        spec.OrgID,
			AssetID:      spec.AssetID,
			VersionID:    spec.VersionID,
			SpecProtocol: spec.SpecProtocol,
			Spec:         spec.Spec,
		}
	}

	// 查询是否有"实例"和"访问管理条目"
	var where = map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": response.Version.SwaggerVersion,
	}
	response.HasInstantiation = svc.FirstRecord(new(apistructs.InstantiationModel), where) == nil
	var access apistructs.APIAccessesModel
	response.HasAccess = svc.FirstRecord(&access, where) == nil
	if response.HasAccess {
		response.Access = &access
	}

	return &response, nil
}

// 查询 minor 下的 instantiation
// ok: 是否存在这样的实例化记录, true: 存在
func (svc *Service) GetInstantiation(req *apistructs.GetInstantiationsReq) (*apistructs.InstantiationModel, bool, *errorresp.APIError) {
	// 参数校验
	if req.OrgID == 0 {
		return nil, false, apierrors.GetInstantiations.MissingParameter(apierrors.MissingOrgID)
	}

	var model apistructs.InstantiationModel
	err := dbclient.OneInstantiation(&model, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": req.URIParams.SwaggerVersion,
		"minor":           req.URIParams.Minor,
	})
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, false, nil
		}
		return nil, false, apierrors.GetInstantiations.InternalError(err)
	}

	return &model, true, nil
}

func (svc *Service) DownloadSpecText(req *apistructs.DownloadSpecTextReq) ([]byte, *errorresp.APIError) {
	// 参数校验
	if req.OrgID == 0 {
		return nil, apierrors.DownloadSpecText.MissingParameter(apierrors.MissingOrgID)
	}

	model, err := dbclient.GetAPIAssetVersionSpec(&apistructs.GetAPIAssetVersionReq{
		OrgID:    req.OrgID,
		Identity: req.Identity,
		URIParams: &apistructs.AssetVersionDetailURI{
			AssetID:   req.URIParams.AssetID,
			VersionID: req.URIParams.VersionID,
		},
		QueryParams: &apistructs.GetAPIAssetVersionQueryParams{
			Asset: false,
			Spec:  true,
		},
	})
	if err != nil {
		return nil, apierrors.DownloadSpecText.InternalError(err)
	}

	if req.QueryParams.SpecProtocol == "" {
		req.QueryParams.SpecProtocol = oasconv.OAS2JSON.String()
	}

	data := []byte(model.Spec)

	switch {
	// spec 格式与所要求的一致, 直接返回
	case req.QueryParams.SpecProtocol == model.SpecProtocol:
		return data, nil

	case req.QueryParams.SpecProtocol == oasconv.OAS2JSON.String() &&
		model.SpecProtocol == oasconv.OAS2YAML.String(),
		req.QueryParams.SpecProtocol == oasconv.OAS3JSON.String() &&
			model.SpecProtocol == oasconv.OAS3YAML.String():
		j, err := oasconv.YAMLToJSON(data)
		if err != nil {
			return data, nil
		}
		return j, nil

	case req.QueryParams.SpecProtocol == oasconv.OAS2YAML.String() &&
		model.SpecProtocol == oasconv.OAS2JSON.String(),
		req.QueryParams.SpecProtocol == oasconv.OAS3YAML.String() &&
			model.SpecProtocol == oasconv.OAS3JSON.String():
		y, err := oasconv.JSONToYAML(data)
		if err != nil {
			return data, nil
		}
		return y, nil

	case req.QueryParams.SpecProtocol == oasconv.OAS2JSON.String():
		v3, err := oas3.LoadFromData([]byte(model.Spec))
		if err != nil {
			return data, nil
		}
		v2, err := oasconv.OAS3ConvTo2(v3)
		if err != nil {
			return data, nil
		}
		j, err := json.Marshal(v2)
		if err != nil {
			return data, nil
		}
		return j, nil
	case req.QueryParams.SpecProtocol == oasconv.OAS2YAML.String():
		v3, err := oas3.LoadFromData([]byte(model.Spec))
		if err != nil {
			return data, nil
		}
		v2, err := oasconv.OAS3ConvTo2(v3)
		if err != nil {
			return data, nil
		}
		j, err := json.Marshal(v2)
		if err != nil {
			return data, nil
		}
		y, err := oasconv.JSONToYAML(j)
		if err != nil {
			return data, nil
		}
		return y, nil
	case req.QueryParams.SpecProtocol == oasconv.OAS3JSON.String():
		v3, err := swagger.LoadFromData(data)
		if err != nil {
			return data, nil
		}
		j, err := json.Marshal(v3)
		if err != nil {
			return data, nil
		}
		return j, nil
	case req.QueryParams.SpecProtocol == oasconv.OAS3YAML.String():
		v3, err := swagger.LoadFromData(data)
		if err != nil {
			return data, nil
		}
		j, err := json.Marshal(v3)
		if err != nil {
			return data, nil
		}
		y, err := oasconv.JSONToYAML(j)
		if err != nil {
			return data, nil
		}
		return y, nil
	default:
		return data, nil
	}
}

func (svc *Service) GetMyClient(req *apistructs.GetClientReq) (*apistructs.ClientObj, *errorresp.APIError) {
	// 参数校验
	if req.OrgID == 0 {
		return nil, apierrors.GetClient.MissingParameter(apierrors.MissingOrgID)
	}
	if req.URIParams == nil {
		return nil, apierrors.GetClient.InvalidParameter("no URI parameter")
	}

	model, err := dbclient.GetMyClient(req, true)
	if err != nil {
		return nil, apierrors.GetClient.InternalError(err)
	}

	credentials, err := bdl.Bdl.GetClientCredentials(model.ClientID)
	if err != nil {
		return nil, apierrors.GetClient.InternalError(err)
	}

	return &apistructs.ClientObj{
		Client: model,
		SK: &apistructs.SK{
			ClientID:     credentials.ClientId,
			ClientSecret: credentials.ClientSecret,
		},
	}, nil
}

func (svc *Service) GetContract(req *apistructs.GetContractReq) (*apistructs.ClientModel, *apistructs.SK, *apistructs.ContractModel, *errorresp.APIError) {
	// 参数校验
	if req == nil || req.URIParams == nil {
		return nil, nil, nil, apierrors.GetContract.InvalidParameter("parameters is invalid")
	}
	if req.OrgID == 0 {
		return nil, nil, nil, apierrors.GetContract.InvalidParameter(apierrors.MissingOrgID)
	}

	// 查询客户端详情
	client, err := dbclient.GetMyClient(&apistructs.GetClientReq{
		OrgID:     req.OrgID,
		Identity:  req.Identity,
		URIParams: &apistructs.GetClientURIParams{ClientID: req.URIParams.ClientID},
	}, true)
	if err != nil {
		return nil, nil, nil, apierrors.GetContract.InternalError(err)
	}

	// 查询 sk
	credentials, err := bdl.Bdl.GetClientCredentials(req.URIParams.ClientID)
	if err != nil {
		return nil, nil, nil, apierrors.GetContract.InternalError(err)
	}

	// 查询合约
	contract, err := dbclient.GetContract(req)
	if err != nil {
		return nil, nil, nil, apierrors.GetContract.InternalError(err)
	}

	return client, &apistructs.SK{
		ClientID:     credentials.ClientId,
		ClientSecret: credentials.ClientSecret,
	}, contract, nil
}

func (svc *Service) GetAccess(req *apistructs.GetAccessReq) (map[string]interface{}, *errorresp.APIError) {
	var (
		asset  apistructs.APIAssetsModel
		access apistructs.APIAccessesModel
	)

	// 查询目标 access
	if err := svc.FirstRecord(&access, map[string]interface{}{
		"org_id": req.OrgID,
		"id":     req.URIParams.AccessID,
	}); err != nil {
		logrus.Errorf("failed to FistRecord access, err: %v", err)
		return nil, apierrors.GetAccess.InternalError(err)
	}

	// 查出对应的 asset
	if err := svc.FirstRecord(&asset, map[string]interface{}{
		"org_id":   req.OrgID,
		"asset_id": access.AssetID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord asset, err: %v", err)
		return nil, apierrors.GetAccess.InternalError(err)
	}

	// 查询 tenantGroupID
	tenantGroupID, err := bdl.Bdl.GetTenantGroupID(access.ProjectID, access.Workspace)
	if err != nil {
		logrus.Errorf("failed to GetTenantGroupID, err: %v", err)
		return nil, apierrors.GetAccess.InternalError(err)
	}

	// 查询流量入口
	var (
		endpointBindDomain []string
		endpointName       string
	)
	endpoint, err := bdl.Bdl.GetEndpoint(access.EndpointID)
	if err != nil {
		logrus.Errorf("failed to GetEndpoint, err: %v", err)
	} else {
		endpointBindDomain = endpoint.BindDomain
		endpointName = endpoint.Name
	}

	// 按钮权限
	rolesSet := bdl.FetchAssetRolesSet(req.OrgID, req.Identity.UserID)
	written := req.Identity.UserID == access.CreatorID || writePermission(rolesSet, &asset)

	return map[string]interface{}{
		"access": &apistructs.GetAccessRspAccess{
			ID:              access.ID,
			AssetID:         access.AssetID,
			AssetName:       access.AssetName,
			OrgID:           access.OrgID,
			SwaggerVersion:  access.SwaggerVersion,
			Major:           access.Major,
			Minor:           access.Minor,
			ProjectID:       access.ProjectID,
			ProjectName:     access.ProjectName,
			Workspace:       access.Workspace,
			EndpointID:      access.EndpointID,
			Authentication:  access.Authentication,
			Authorization:   access.Authorization,
			AddonInstanceID: access.AddonInstanceID,
			BindDomain:      endpointBindDomain,
			CreatorID:       access.CreatorID,
			UpdaterID:       access.UpdaterID,
			CreatedAt:       access.CreatedAt,
			UpdatedAt:       access.UpdatedAt,
			TenantGroupID:   tenantGroupID,
			EndpointName:    endpointName,
		},
		"tenantGroup": &apistructs.GetAccessRspTenantGroup{
			TenantGroupID: tenantGroupID,
		},
		"permission": map[string]bool{
			"edit":   written && endpoint != nil,
			"delete": written,
		},
	}, nil
}

func (svc *Service) GetSLA(req *apistructs.GetSLAReq) (*apistructs.GetSLARsp, *errorresp.APIError) {
	// 参数校验
	if req == nil || req.URIParams == nil {
		return nil, apierrors.GetSLA.InvalidParameter("无效的参数")
	}
	if req.OrgID == 0 {
		return nil, apierrors.GetSLA.InvalidParameter("无效的参数 OrgID")
	}

	var (
		access apistructs.APIAccessesModel
		sla    apistructs.SLAModel
		limits []*apistructs.SLALimitModel
	)

	if err := svc.FirstRecord(&access, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.SwaggerVersion,
		"swagger_version": req.URIParams.SwaggerVersion,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord access, err: %v", err)
		return nil, apierrors.GetSLA.InternalError(errors.New("查询访问管理失败"))
	}

	if err := svc.FirstRecord(&sla, map[string]interface{}{
		"id": req.URIParams.SLAID,
	}); err != nil {
		logrus.Errorf("failed to FirstRecord SLA, err: %v", err)
		return nil, apierrors.GetSLA.InternalError(errors.New("查询 SLA 失败"))
	}

	if err := svc.ListRecords(&limits, map[string]interface{}{
		"sla_id": req.URIParams.SLAID,
	}); err != nil {
		logrus.Errorf("failed to ListRecord limits, err: %v", err)
		return nil, apierrors.GetSLA.InternalError(errors.New("查询 SLA 限制条件失败"))
	}

	var rsp = apistructs.GetSLARsp{
		SLAModel:       apistructs.SLAModel{},
		Limits:         limits,
		AssetID:        req.URIParams.AssetID,
		AssetName:      access.AssetName,
		SwaggerVersion: req.URIParams.SwaggerVersion,
		Default:        access.DefaultSLAID != nil && *access.DefaultSLAID == req.URIParams.SLAID,
	}

	return &rsp, nil
}

func (svc *Service) FirstRecord(model interface{}, where map[string]interface{}) error {
	return dbclient.Sq().First(model, where).Error
}

func (svc *Service) ListRecords(models interface{}, where map[string]interface{}) error {
	return dbclient.Sq().Find(models, where).Error
}

func (svc *Service) GetProject(projectID uint64) (*apistructs.ProjectDTO, error) {
	return bdl.Bdl.GetProject(projectID)
}

func (svc *Service) GetApp(appID uint64) (*apistructs.ApplicationDTO, error) {
	return bdl.Bdl.GetApp(appID)
}
