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

package applicationlist

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	appconfigform "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/edge-application/components/appConfigForm"
)

type ComponentList struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
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
func (c *ComponentList) OperateChangePage(orgID int64, identity apistructs.Identity) (err error) {
	var (
		reqPageNo   = apistructs.EdgeDefaultPageNo
		reqPageSize = apistructs.EdgeDefaultPageSize
		appState    = apistructs.EdgeAppState{}
	)

	jsonData, err := json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &appState)
	if err != nil {
		return err
	}

	if appState.PageNo > 0 && appState.PageSize > 0 {
		reqPageNo = appState.PageNo
		reqPageSize = appState.PageSize
	}

	req := &apistructs.EdgeAppListPageRequest{
		OrgID:    orgID,
		PageNo:   reqPageNo,
		PageSize: reqPageSize,
	}

	res, err := c.ctxBundle.Bdl.ListEdgeApp(req, identity)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	if err != nil {
		return fmt.Errorf("list edge application error: %v", err)
	}

	resList := make([]EdgeAPPItem, 0)

	for _, data := range res.List {
		deployResource := data.Type
		if data.Image != "" {
			deployResource += fmt.Sprintf("(%s)", data.Image)
		} else if deployResource == "addon" && data.AddonName != "" && data.AddonVersion != "" {
			deployResource += fmt.Sprintf("(%s:%s)", data.AddonName, data.AddonVersion)
		}

		clusterInfo, err := c.ctxBundle.Bdl.GetCluster(strconv.Itoa(int(data.ClusterID)))
		if err != nil {
			return fmt.Errorf("get cluster(id: %d) info error: %v", data.ClusterID, err)
		}
		var item = EdgeAPPItem{
			ID:              int64(data.ID),
			ApplicationName: renderAppName(data.Name, int64(data.ID)),
			Cluster:         clusterInfo.Name,
			DeployResource:  deployResource,
			Operate:         getAPPItemOperate(data.Name, appconfigform.ConvertDeployResource(data.Type), int64(data.ID), i18nLocale),
		}
		resList = append(resList, item)
	}

	c.component.Data = map[string]interface{}{
		"list": resList,
	}

	c.component.State["total"] = res.Total
	c.component.State["pageSize"] = reqPageSize
	c.component.State["pageNo"] = reqPageNo

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

	err = c.ctxBundle.Bdl.DeleteEdgeApp(meta.Meta["id"], identity)
	if err != nil {
		return err
	}

	err = c.OperateChangePage(orgID, identity)
	if err != nil {
		return err
	}

	return
}

func (c *ComponentList) OperateReload(operationData map[string]interface{}) (err error) {
	var (
		meta = apistructs.EdgeEventMeta{}
	)

	jsonData, err := json.Marshal(operationData)
	if err != nil {
		return fmt.Errorf("marshal operation data error: %v", err)
	}

	err = json.Unmarshal(jsonData, &meta)
	if err != nil {
		return fmt.Errorf("unmarshal operation data error: %v", err)
	}

	c.component.State["addAppDrawerVisible"] = true
	c.component.State["appConfigFormVisible"] = true
	c.component.State["appID"] = meta.Meta["id"]

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentList{}
}
