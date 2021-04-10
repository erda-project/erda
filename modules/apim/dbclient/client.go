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
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/strutil"
)

func ListMyClients(req *apistructs.ListMyClientsReq, orgManager bool) (total uint64, models []*apistructs.ClientModel, err error) {
	var (
		keyword = strutil.Concat("%", req.QueryParams.Keyword, "%")
	)

	sq := DB.Where("org_id = ?", req.OrgID).
		Where("? = true OR ? = creator_id", orgManager, req.Identity.UserID)
	if req.QueryParams.Keyword != "" {
		sq = sq.Where("name LIKE ? OR client_id LIKE ?", keyword, keyword)
	}
	sq = sq.Order("updated_at DESC")

	if req.QueryParams.Paging {
		if err := sq.Limit(req.QueryParams.PageSize).Offset((req.QueryParams.PageNo - 1) * req.QueryParams.PageSize).Find(&models).
			Limit(-1).Offset(0).Count(&total).
			Error; err != nil {
			return 0, nil, err
		}
	} else {
		if err := sq.Find(&models).Count(&total).Error; err != nil {
			return 0, nil, err
		}
	}

	return total, models, nil
}

// 查询本人名下某个 Client 详情. 注意传入的 ClientID 可以是 dice_api_clients 的主键, 也可以是 client_id 字段
func GetMyClient(req *apistructs.GetClientReq, orgManager bool) (*apistructs.ClientModel, error) {
	var (
		model apistructs.ClientModel
	)
	if err := DB.Where("org_id = ?", req.OrgID).
		Where("? in (client_id, id)", req.URIParams.ClientID).
		Where("? = creator_id OR ? = true", req.Identity.UserID, orgManager).
		First(&model).
		Error; err != nil {
		return nil, err
	}
	return &model, nil
}

func ListSwaggerVersionClients(req *apistructs.ListSwaggerVersionClientsReq) ([]*apistructs.ListSwaggerVersionClientOjb, error) {
	var (
		list      []*apistructs.ListSwaggerVersionClientOjb
		contracts []*apistructs.ContractModel
	)
	if err := DB.Find(&contracts, map[string]interface{}{
		"org_id":          req.OrgID,
		"asset_id":        req.URIParams.AssetID,
		"swagger_version": req.URIParams.SwaggerVersion,
	}).
		Error; err != nil {
		return nil, errors.Wrapf(err, "failed to Find contracts, asset_id: %s, swagger_version: %s", req.URIParams.AssetID, req.URIParams.SwaggerVersion)
	}

	var slaNames = make(map[uint64]string)
	for _, contract := range contracts {
		var (
			client apistructs.ClientModel
		)
		if err := DB.First(&client, map[string]interface{}{
			"org_id": req.OrgID,
			"id":     contract.ClientID,
		}).Error; err != nil {
			continue
		}

		obj := apistructs.ListSwaggerVersionClientOjb{
			Client: &client,
			Contract: &apistructs.ContractModelAdvance{
				ContractModel:     *contract,
				ClientName:        client.Name,
				ClientDisplayName: client.DisplayName,
				CurSLAName:        getSLAName(contract.CurSLAID, slaNames),
				RequestSLAName:    getSLAName(contract.RequestSLAID, slaNames),
			},
			Permission: map[string]bool{"edit": false},
		}
		list = append(list, &obj)
	}

	return list, nil
}

// 给定一个 slaID, 如果为空, 返回 "";
// 如果 slaID 对应的 slaName 已经在给定的集合中, 则返回集合中的结果;
// 否则查询库表获取 slaName, 并将结果记录在给定的集合中
func getSLAName(slaID *uint64, names map[uint64]string) string {
	if slaID == nil {
		return ""
	}

	if *slaID == 0 {
		return "无限制 SLA"
	}

	name, ok := names[*slaID]
	if ok {
		return name
	}

	var sla apistructs.SLAModel
	if err := DB.First(&sla, map[string]interface{}{"id": *slaID}).
		Error; err != nil {
		return ""
	}

	names[*slaID] = sla.Name

	return sla.Name
}
