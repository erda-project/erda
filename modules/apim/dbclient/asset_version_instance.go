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
