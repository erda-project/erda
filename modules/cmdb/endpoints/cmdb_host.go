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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/numeral"
)

// GetHost 获取指定集群下某个host的信息
func (e *Endpoints) GetHost(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetHost.NotLogin().ToResp(), nil
	}

	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrGetHost.NotLogin().ToResp(), nil
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrGetHost.InvalidParameter(err).ToResp(), nil
	}
	org, err := e.bdl.GetOrg(int64(orgID))
	if err != nil {
		return apierrors.ErrGetHost.InternalError(err).ToResp(), nil
	}

	// 操作鉴权
	req := apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.HostResource,
		Action:   apistructs.GetAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrGetHost.AccessDenied().ToResp(), nil
	}

	addr := vars["host"]
	if addr == "" {
		return apierrors.ErrGetHost.MissingParameter("host").ToResp(), nil
	}

	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrGetHost.MissingParameter("clusterName").ToResp(), nil
	}

	host, err := e.host.GetByClusterAndIP(clusterName, addr)
	if err != nil {
		return apierrors.ErrGetHost.InternalError(err).ToResp(), nil
	}
	if host == nil {
		return apierrors.ErrGetHost.NotFound().ToResp(), nil
	}

	// 若host 企业名称与当前企业名称不符, 返回错误
	if !strings.Contains(host.OrgName, org.Name) {
		return apierrors.ErrGetHost.NotFound().ToResp(), nil
	}

	return httpserver.OkResp(host)
}

// SyncHostResource 定时同步主机实际使用资源
func (e *Endpoints) SyncHostResource(interval time.Duration) {
	l := loop.New(loop.WithInterval(interval))
	l.Do(func() (bool, error) {
		clusters, err := e.cluster.ListCluster()
		if err != nil {
			logrus.Errorf("failed to get cluster list, %v", err)
		}
		for _, cluster := range *clusters {
			metrics, err := e.bdl.GetHostMetricInfo(cluster.Name)
			if err != nil {
				logrus.Errorf("failed to sync host usage info, %v", err)
				continue
			}
			for k, v := range metrics {
				// 若找不到对应 host, 待host信息写入DB后再同步
				if host, _ := e.host.GetByClusterAndPrivateIP(cluster.Name, k); host != nil {
					host.CpuUsage = numeral.Round(v.CPU*host.Cpus/100, 2)
					host.MemoryUsage = int64(v.Memory * float64(host.Memory) / 100)
					host.DiskUsage = int64(v.Disk * float64(host.Disk) / 100)
					host.Load5 = v.Load
					e.host.Update(host)
				}
			}
		}
		return false, nil
	})
}
