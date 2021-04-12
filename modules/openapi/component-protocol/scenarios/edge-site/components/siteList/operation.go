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

package sitelist

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type ComponentList struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type EdgeSiteState struct {
	apistructs.EdgeSiteState
	IsFirstFilter bool `json:"isFirstFilter"`
}

func (c *ComponentList) SetBundle(ctxBundle protocol.ContextBundle) error {
	if ctxBundle.Bdl == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.ctxBundle = ctxBundle
	return nil
}

func (c *ComponentList) SetComponent(component *apistructs.Component) error {
	if component == nil {
		return fmt.Errorf("invalie bundle")
	}
	c.component = component
	return nil
}

// OperateChangePage
func (c *ComponentList) OperateChangePage(orgID int64, reList bool, identity apistructs.Identity) (err error) {
	var (
		reqPageNo   = apistructs.EdgeDefaultPageNo
		reqPageSize = apistructs.EdgeDefaultPageSize
		stateEntity = EdgeSiteState{}
	)

	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonData, &stateEntity); err != nil {
		return err
	}

	if stateEntity.IsFirstFilter {
		reqPageNo = 1
	} else if !stateEntity.IsFirstFilter && stateEntity.PageNo != 0 {
		reqPageNo = stateEntity.PageNo
	}

	if stateEntity.PageSize != 0 {
		reqPageSize = stateEntity.PageSize
	}

	if reList {
		totalListReq := &apistructs.EdgeSiteListPageRequest{
			OrgID:     orgID,
			NotPaging: true,
		}

		if stateEntity.SearchCondition != "" {
			totalListReq.Search = stateEntity.SearchCondition
		}

		allSite, err := c.ctxBundle.Bdl.ListEdgeSite(totalListReq, identity)

		if err != nil {
			return err
		}

		if allSite.Total <= (reqPageNo-1)*reqPageSize && reqPageNo-1 > 0 {
			reqPageNo--
		}
	}

	req := &apistructs.EdgeSiteListPageRequest{
		OrgID:    orgID,
		PageNo:   reqPageNo,
		PageSize: reqPageSize,
		Search:   stateEntity.SearchCondition,
	}

	res, err := c.ctxBundle.Bdl.ListEdgeSite(req, identity)

	if err != nil {
		return fmt.Errorf("list edge site error: %v", err)
	}

	resList := make([]EdgeSiteItem, 0)

	for _, data := range res.List {
		item := EdgeSiteItem{
			ID:             data.ID,
			NodeNum:        data.NodeCount,
			SiteName:       renderSiteName(data.ClusterName, data.Name, data.ID),
			RelatedCluster: data.ClusterName,
			Operate:        getSiteItemOperate(data),
		}

		resList = append(resList, item)
	}

	c.component.Data = map[string]interface{}{
		"list": resList,
	}

	c.component.State["total"] = res.Total
	c.component.State["pageSize"] = reqPageSize
	c.component.State["pageNo"] = reqPageNo
	c.component.State["isFirstFilter"] = false

	return
}

func (c *ComponentList) OperateDelete(orgID int64, operationData interface{}, identity apistructs.Identity) (err error) {
	var (
		meta = apistructs.EdgeEventMeta{}
	)

	jsonData, err := json.Marshal(operationData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &meta)
	if err != nil {
		return err
	}

	err = c.ctxBundle.Bdl.DeleteEdgeSite(meta.Meta["id"], identity)
	if err != nil {
		return err
	}

	err = c.OperateChangePage(orgID, true, identity)
	if err != nil {
		return err
	}

	return
}

func (c *ComponentList) OperateReload(operationData interface{}, operation string) (err error) {
	var (
		meta = apistructs.EdgeEventMeta{}
	)

	jsonData, err := json.Marshal(operationData)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &meta)
	if err != nil {
		return err
	}

	if operation == apistructs.EdgeOperationAdd {
		c.component.State["sitePreviewVisible"] = true
		c.component.State["siteAddDrawerVisible"] = true
	} else if operation == apistructs.EdgeOperationUpdate {
		c.component.State["siteFormModalVisible"] = true
	}

	c.component.State["siteID"] = meta.Meta["id"]

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentList{}
}
