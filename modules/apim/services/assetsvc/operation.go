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

package assetsvc

import (
	"encoding/json"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/dbclient"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (svc *Service) SearchOperations(req *apistructs.SearchOperationsReq) (results []*apistructs.APIOAS3IndexModel, apiError *errorresp.APIError) {
	// 查询用户可以查看 API 集市
	response, err := svc.PagingAsset(apistructs.PagingAPIAssetsReq{
		OrgID:    req.OrgID,
		Identity: req.Identity,
		QueryParams: &apistructs.PagingAPIAssetsQueryParams{
			Paging:        false,
			PageNo:        0,
			PageSize:      0,
			Keyword:       "",
			Scope:         "",
			HasProject:    false,
			LatestVersion: false,
			LatestSpec:    false,
			Instantiation: false,
		},
	})
	if err != nil {
		return nil, apierrors.SearchOperations.InternalError(err)
	}

	if response.Total == 0 {
		return nil, nil
	}

	var assetIDs []string
	for _, l := range response.List {
		assetIDs = append(assetIDs, l.Asset.AssetID)
	}

	// 查询这些集市下的所有版本
	var versions []*apistructs.APIAssetVersionsModel
	if find := dbclient.Sq().Where("org_id = ?", req.OrgID).
		Where("deprecated = ?", false).
		Where("asset_id IN (?)", assetIDs).
		Order("swagger_version DESC, major DESC, minor DESC, patch DESC").
		Find(&versions); find.Error != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, apierrors.SearchOperations.InternalError(err)
	}

	// 筛选出每个 swaggerVersion 下的最新版本
	var (
		versionsM  = make(map[string]*apistructs.APIAssetVersionsModel)
		versionIDs []uint64
	)
	for _, v := range versions {
		if _, ok := versionsM[v.SwaggerVersion]; ok {
			continue
		}
		versionsM[v.SwaggerVersion] = v
		versionIDs = append(versionIDs, v.ID)
	}

	// 在筛选出的 version 下搜索
	keyword := "%" + req.QueryParams.Keyword + "%"
	if find := dbclient.Sq().Where("asset_id like ? OR asset_name like ? OR operation_id like ? OR path like ? OR description like ?",
		keyword, keyword, keyword, keyword, keyword).
		Where("version_id IN (?)", versionIDs).
		Find(&results); find.Error != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, apierrors.SearchOperations.InternalError(err)
	}

	return results, nil
}

// node 包含 assert_id, info_version, path, method 四个字段的信息
func (svc *Service) GetOperation(req *apistructs.GetOperationReq) (data *apistructs.GetOperationResp, apiError *errorresp.APIError) {
	var index = apistructs.APIOAS3IndexModel{ID: req.URIParams.ID}
	if first := dbclient.Sq().First(&index); first.Error != nil {
		if gorm.IsRecordNotFoundError(first.Error) {
			return nil, apierrors.GetOperation.NotFound()
		}
		return nil, apierrors.GetOperation.InternalError(first.Error)
	}

	var fragment apistructs.APIOAS3FragmentModel
	if first := dbclient.Sq().First(&fragment, map[string]interface{}{"index_id": req.URIParams.ID}); first.Error != nil {
		if gorm.IsRecordNotFoundError(first.Error) {
			return nil, apierrors.GetOperation.NotFound()
		}
		return nil, apierrors.GetOperation.InternalError(first.Error)
	}

	var resp = apistructs.GetOperationResp{
		ID:          index.ID,
		AssetID:     index.AssetID,
		AssetName:   index.AssetName,
		Version:     index.InfoVersion,
		Path:        index.Path,
		Method:      index.Method,
		OperationID: index.OperationID,
		Operation:   json.RawMessage(fragment.Operation),
	}

	return &resp, nil
}
