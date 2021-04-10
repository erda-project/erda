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

func ListContracts(req *apistructs.ListContractsReq) (uint64, []*apistructs.ContractModelAdvance, error) {
	var (
		models []*apistructs.ContractModel
		total  uint64
	)

	DB.Where(map[string]interface{}{
		"org_id":    req.OrgID,
		"client_id": req.URIParams.ClientID,
	}).
		Where("status in (?)", req.QueryParams.Status).
		Order("updated_at DESC").
		Offset((req.QueryParams.PageNo - 1) * req.QueryParams.PageNo).Limit(req.QueryParams.PageSize).
		Find(&models).
		Offset(0).Limit(-1).
		Count(&total)

	advances := make([]*apistructs.ContractModelAdvance, len(models))
	names := make(map[uint64]string)
	for i, v := range models {
		var client apistructs.ClientModel
		DB.First(&client, map[string]interface{}{"id": v.ClientID})

		advances[i] = &apistructs.ContractModelAdvance{
			ContractModel:     *v,
			ClientName:        client.Name,
			ClientDisplayName: client.DisplayName,
			CurSLAName:        getSLAName(v.CurSLAID, names),
			RequestSLAName:    getSLAName(v.RequestSLAID, names),
			EndpointName:      "",
			ProjectID:         0,
			Workspace:         "",
		}
	}

	return total, advances, nil
}

func GetContract(req *apistructs.GetContractReq) (*apistructs.ContractModel, error) {
	var (
		model apistructs.ContractModel
	)

	if err := Sq().Where(map[string]interface{}{
		"org_id":    req.OrgID,
		"client_id": req.URIParams.ClientID,
		"id":        req.URIParams.ContractID,
	}).First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func ListContractRecords(req *apistructs.ListContractRecordsReq) ([]*apistructs.ContractRecordModel, error) {
	var (
		models []*apistructs.ContractRecordModel
	)

	if err := Sq().Where(map[string]interface{}{
		"org_id":      req.OrgID,
		"contract_id": req.URIParams.ContractID,
	}).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	return models, nil
}
