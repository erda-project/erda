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
	"sort"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func ListAccess(req *apistructs.ListAccessReq, responsibleAssetIDs []string) (uint64, []*apistructs.ListAccessObj, error) {
	keyword := strutil.Concat("%", req.QueryParams.Keyword, "%")
	var accesses []*apistructs.APIAccessesModel

	if req.QueryParams.PageSize < 1 {
		req.QueryParams.PageSize = 15
	}
	if req.QueryParams.PageSize > 500 {
		req.QueryParams.PageSize = 500
	}
	if req.QueryParams.PageNo < 1 {
		req.QueryParams.PageNo = 1
	}
	if !req.QueryParams.Paging {
		req.QueryParams.PageSize = 500
		req.QueryParams.PageNo = 1
	}

	// 查出有访问管理且我负责的 asset
	rows, err := Sq().Raw(DistinctAssetIDFromAccess,
		req.OrgID,
		req.QueryParams.Keyword, req.QueryParams.Keyword, keyword,
		responsibleAssetIDs,
		req.QueryParams.PageSize, req.QueryParams.PageSize*(req.QueryParams.PageNo-1),
	).Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, nil, nil
		}
		logrus.Errorf("failed to Raw DistinctAssetIDFromAccess, err: err")
		return 0, nil, err
	}
	defer rows.Close()

	var assetIDs []string
	for rows.Next() {
		var assetID string
		if err := rows.Scan(&assetID); err != nil {
			return 0, nil, err
		}
		assetIDs = append(assetIDs, assetID)
	}
	assetIDs = strutil.DedupSlice(assetIDs)

	var total uint64
	if err := Sq().Raw(SelectFoundRows).Row().Scan(&total); err != nil {
		logrus.Errorf("failed to Raw SelectFoundRows, err: %v", err)
		return 0, nil, err
	}

	if err := Sq().Where("org_id = ?", req.OrgID).
		Where("asset_id IN (?)", assetIDs).
		Find(&accesses).
		Error; err != nil {
		logrus.Errorf("failed to Find accesses, err: %v", err)
		return 0, nil, err
	}

	m := make(map[string]*apistructs.ListAccessObj)
	for _, access := range accesses {
		data := apistructs.ListAccessObjChild{
			ID:             access.ID,
			SwaggerVersion: access.SwaggerVersion,
			AppCount:       0,
			ProjectID:      access.ProjectID,
			CreatorID:      access.CreatorID,
			CreatedAt:      access.CreatedAt,
			UpdatedAt:      access.UpdatedAt,
			Permission:     map[string]bool{"edit": false, "delete": false},
		}

		// client 计数
		var contracts []*apistructs.ContractModel
		if err := Sq().Order("updated_at DESC, created_at DESC").
			Find(&contracts, map[string]interface{}{
				"org_id":          access.OrgID,
				"asset_id":        access.AssetID,
				"swagger_version": access.SwaggerVersion,
			}).Error; err != nil || len(contracts) == 0 {
			data.AppCount = 0
		} else {
			var (
				clientPrimaries []uint64
				clientCount     uint64
			)
			for _, contract := range contracts {
				clientPrimaries = append(clientPrimaries, contract.ClientID)
			}
			_ = Sq().Model(new(apistructs.ClientModel)).
				Where("org_id = ?", req.OrgID).
				Where("id in (?)", clientPrimaries).Count(&clientCount).Error
			data.AppCount = clientCount
		}

		if obj, ok := m[access.AssetID]; ok {
			obj.Children = append(obj.Children, &data)
			obj.TotalChildren += 1
		} else {
			m[access.AssetID] = &apistructs.ListAccessObj{
				AssetID:       access.AssetID,
				AssetName:     access.AssetName,
				TotalChildren: 1,
				UpdatedAt:     access.UpdatedAt,
				Children:      []*apistructs.ListAccessObjChild{&data},
			}
		}
	}

	var list []*apistructs.ListAccessObj
	for _, v := range m {
		sort.Slice(v.Children, func(i, j int) bool {
			return v.Children[i].UpdatedAt.After(v.Children[j].UpdatedAt)
		})
		list = append(list, v)
	}
	sort.Slice(list, func(i, j int) bool {
		if len(list[i].Children) == 0 || len(list[j].Children) == 0 {
			return false
		}
		return list[i].Children[0].UpdatedAt.After(list[j].Children[0].UpdatedAt)
	})

	return total, list, nil
}
