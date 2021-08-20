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

package siteiplist

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	DefaultSelected = "success"
)

type ComponentList struct {
	ctxBundle protocol.ContextBundle
	component *apistructs.Component
}

type EdgeAppSiteIPInParam struct {
	ID       int64  `json:"id"`
	AppName  string `json:"appName"`
	SiteName string `json:"siteName"`
}

type EdgeAppSiteIPState struct {
	apistructs.EdgeViewGroupSelectState
	apistructs.EdgePageState
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
		selectScope = DefaultSelected
		inParam     = EdgeAppSiteIPInParam{}
		stateEntity = EdgeAppSiteIPState{}
	)
	i18nLocale := c.ctxBundle.Bdl.GetLocale(c.ctxBundle.Locale)
	jsonData, err := json.Marshal(c.ctxBundle.InParams)
	if err != nil {
		return fmt.Errorf("marshal id from inparams error: %v", err)
	}

	if err = json.Unmarshal(jsonData, &inParam); err != nil {
		return fmt.Errorf("unmarshal inparam to object error: %v", err)
	}

	jsonData, err = json.Marshal(c.component.State)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonData, &stateEntity); err != nil {
		return err
	} else {
		if stateEntity.PageSize != 0 && stateEntity.PageNo != 0 {
			reqPageSize = stateEntity.PageSize
			reqPageNo = stateEntity.PageNo
		}
		if stateEntity.ViewGroupSelected != "" {
			selectScope = stateEntity.ViewGroupSelected
		}
	}

	resList := make([]EdgeSiteMachineItem, 0)

	res, err := c.ctxBundle.Bdl.GetEdgeInstanceInfo(orgID, inParam.AppName, inParam.SiteName, identity)
	if err != nil {
		return err
	}

	for index, data := range res {
		item := EdgeSiteMachineItem{
			ID:        int64(index),
			IP:        data.ContainerIP,
			Address:   data.HostIP,
			CreatedAt: data.StartedAt.Format("2006-01-02 15:04:05"),
			Status:    getStatus(data.Phase),
			Operate:   getItemOperations(data.ContainerID, data.ContainerIP, data.Cluster, i18nLocale),
		}

		switch selectScope {
		case "success":
			if GetEdgeApplicationContainerStatus(data.Phase) == "success" {
				resList = append(resList, item)
			}
			break
		case "error":
			if GetEdgeApplicationContainerStatus(data.Phase) == "error" {
				resList = append(resList, item)
			}
			break
		default:
			resList = append(resList, item)
		}
	}

	total := len(resList)

	start := (reqPageNo - 1) * reqPageSize
	end := start + reqPageSize

	if end > len(resList) {
		resList = resList[start:]
	} else {
		resList = resList[start:end]
	}

	c.component.Data = map[string]interface{}{
		"list": resList,
	}

	c.component.State["total"] = total
	c.component.State["pageSize"] = reqPageSize
	c.component.State["pageNo"] = reqPageNo

	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentList{}
}
