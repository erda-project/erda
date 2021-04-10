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

func FirstOrCreateInstantiation(model *apistructs.InstantiationModel, params map[string]interface{}) error {
	return DB.FirstOrCreate(model, params).Error
}

func OneInstantiation(model *apistructs.InstantiationModel, params map[string]interface{}) error {
	return DB.First(model, params).Error
}

func UpdateInstantiation(req *apistructs.UpdateInstantiationReq) error {
	where := map[string]interface{}{
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": req.URIParams.SwaggerVersion,
		"id":              req.URIParams.InstantiationID,
		"org_id":          req.OrgID,
	}
	updates := map[string]interface{}{
		"type":         req.Body.Type,
		"url":          req.Body.URL,
		"project_id":   req.Body.ProjectID,
		"app_id":       req.Body.AppID,
		"runtime_id":   req.Body.RuntimeID,
		"service_name": req.Body.ServiceName,
		"workspace":    req.Body.Workspace,
	}

	return Sq().Model(new(apistructs.InstantiationModel)).
		Where(where).
		Updates(updates).
		Error
}
