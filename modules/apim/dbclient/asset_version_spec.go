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

package dbclient

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda/apistructs"
)

// APIAssetVersionSpec API 资料版本 Spec
// 因 Spec 为纯文本，可能很大，为了不影响查询性能，故与 APIAssetVersion 拆分
type APIAssetVersionSpecsModel apistructs.APIAssetVersionSpecsModel

func (m APIAssetVersionSpecsModel) TableName() string {
	return "dice_api_asset_version_specs"
}

func CreateOrUpdateAPIAssetVersionSpec(spec *APIAssetVersionSpecsModel) error {
	if spec.VersionID == 0 {
		return fmt.Errorf("missing versionID")
	}

	exist, err := GetAPIAssetVersionSpec(&apistructs.GetAPIAssetVersionReq{
		OrgID:    spec.OrgID,
		Identity: nil,
		URIParams: &apistructs.AssetVersionDetailURI{
			AssetID:   spec.AssetID,
			VersionID: spec.VersionID,
		},
		QueryParams: &apistructs.GetAPIAssetVersionQueryParams{
			Asset: false,
			Spec:  false,
		},
	})
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return err
		}
		// not exist, do create
		spec.UpdaterID = ""
		return DB.Create(spec).Error
	}
	// do update
	spec.CreatorID = exist.CreatorID
	spec.ID = exist.ID
	return DB.Save(spec).Error
}

func GetAPIAssetVersionSpec(req *apistructs.GetAPIAssetVersionReq) (*APIAssetVersionSpecsModel, error) {
	var spec APIAssetVersionSpecsModel
	if err := Sq().Where(map[string]interface{}{
		"org_id":     req.OrgID,
		"asset_id":   req.URIParams.AssetID,
		"version_id": req.URIParams.VersionID,
	}).First(&spec).Error; err != nil {
		return nil, err
	}
	return &spec, nil
}

func QuerySpecsFromVersions(orgID uint64, assetID string, versionIDs []string, m map[uint64]*apistructs.PagingAPIAssetVersionRspObj) error {
	var models []APIAssetVersionSpecsModel
	if err := Sq().
		Where(map[string]interface{}{
			"org_id":   orgID,
			"asset_id": assetID,
		}).Where("version_id in (?)", strings.Join(versionIDs, ", ")).
		Find(&models).
		Error; err != nil {
		return err
	}
	for _, v := range models {
		if obj, ok := m[v.VersionID]; ok {
			spec := apistructs.APIAssetVersionSpecsModel(v)
			obj.Spec = &spec
		}
	}
	return nil
}
