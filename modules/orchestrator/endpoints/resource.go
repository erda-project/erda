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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// 获取项目资源信息，包括service和addon
func (e *Endpoints) GetProjectResource(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 校验 body 合法性
	var projectIDs []uint64
	if err := json.NewDecoder(r.Body).Decode(&projectIDs); err != nil {
		return apierrors.ErrProjectResource.InvalidParameter(err).ToResp(), nil
	}
	resourceMap := map[uint64]apistructs.ProjectResourceItem{}
	if len(projectIDs) == 0 {
		return httpserver.OkResp(resourceMap)
	}
	logrus.Infof("get project resource with project ids: %+v", projectIDs)

	serviceResource, err := e.resource.GetProjectServiceResource(projectIDs)
	if err != nil {
		return apierrors.ErrProjectResource.InvalidParameter(err).ToResp(), nil
	}
	addonResource, err := e.resource.GetProjectAddonResource(projectIDs)
	if err != nil {
		return apierrors.ErrProjectResource.InvalidParameter(err).ToResp(), nil
	}
	for _, v := range projectIDs {

		addonResourceItem := (*addonResource)[v]
		serviceResourceItem := (*serviceResource)[v]
		resourceMap[v] = apistructs.ProjectResourceItem{
			MemServiceUsed: serviceResourceItem.MemServiceUsed,
			CpuServiceUsed: serviceResourceItem.CpuServiceUsed,
			MemAddonUsed:   addonResourceItem.MemAddonUsed,
			CpuAddonUsed:   addonResourceItem.CpuAddonUsed,
		}
	}
	return httpserver.OkResp(resourceMap)

}

// 获取集群中service和addon的数量
func (e *Endpoints) GetClusterResourceReference(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 校验 body 合法性
	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrClusterResource.MissingParameter("clusterName").ToResp(), nil
	}
	orgID := r.URL.Query().Get("orgId")
	if orgID == "" {
		return apierrors.ErrClusterResource.MissingParameter("orgID").ToResp(), nil
	}
	addonRef, err := e.db.CountAddonReferenceByClusterAndOrg(clusterName, orgID)
	if err != nil {
		return apierrors.ErrClusterResource.InternalError(err).ToResp(), nil
	}
	serviceRef, err := e.db.CountServiceReferenceByClusterAndOrg(clusterName, orgID)
	if err != nil {
		return apierrors.ErrClusterResource.InternalError(err).ToResp(), nil
	}
	logrus.Infof("cluster resource ref, addon: %d, service: %d", addonRef, serviceRef)
	return httpserver.OkResp(apistructs.ResourceReferenceData{
		AddonReference:   int64(addonRef),
		ServiceReference: int64(serviceRef),
	})

}
