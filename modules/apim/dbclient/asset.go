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
	"github.com/erda-project/erda/apistructs"
)

// APIAsset is dice_api_assets model
type APIAssetsModel apistructs.APIAssetsModel

func (APIAssetsModel) TableName() string {
	return "dice_api_assets"
}

func GetAPIAsset(req *apistructs.GetAPIAssetReq) (*APIAssetsModel, error) {
	var asset APIAssetsModel
	if err := Sq().Where("`org_id` = ?", req.OrgID).Where("BINARY `asset_id` = ?", req.URIParams.AssetID).
		First(&asset).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func QueryAPILatestVersion(orgID uint64, assetID string) (*apistructs.APIAssetVersionsModel, error) {
	var model apistructs.APIAssetVersionsModel
	if err := Sq().Where(map[string]interface{}{
		"org_id":   orgID,
		"asset_id": assetID,
	}).Order("major DESC, minor DESC, patch DESC").
		First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func QueryVersionLatestSpec(orgID, versionID uint64) (*apistructs.APIAssetVersionSpecsModel, error) {
	var model apistructs.APIAssetVersionSpecsModel
	if err := Sq().Where(map[string]interface{}{
		"org_id":     orgID,
		"version_id": versionID,
	}).First(&model).Error; err != nil {
		return nil, err
	}
	return &model, nil
}

// DeleteAPIAssetByOrgAssetID deletes the APIAsset and APIAssetVersion record for giving assetID
func DeleteAPIAssetByOrgAssetID(orgID uint64, assetID string, cascade bool) error {
	tx := Tx()
	defer tx.RollbackUnlessCommitted()

	params := map[string]interface{}{
		"org_id":   orgID,
		"asset_id": assetID,
	}

	// if cascading delete, delete instance, spce text, version records
	if cascade {
		if err := tx.Delete(new(apistructs.InstantiationModel), params).Error; err != nil {
			return err
		}

		if err := tx.Delete(new(apistructs.APIAssetVersionSpecsModel), params).Error; err != nil {
			return err
		}

		if err := tx.Delete(new(apistructs.APIAssetVersionsModel), params).Error; err != nil {
			return err
		}
	}

	if err := tx.Delete(new(apistructs.APIAssetsModel), params).Error; err != nil {
		return err
	}

	tx.Commit()

	return nil
}
