// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package details_apis

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/pkg/transport/http"
	"github.com/erda-project/erda-infra/providers/httpserver"
	api "github.com/erda-project/erda/pkg/common/httpapi"
	"github.com/erda-project/erda/pkg/common/permission"
)

func (p *provider) getOrgIDByClusters(ctx httpserver.Context) (string, error) {
	req := ctx.Request()
	idStr := api.OrgID(req)
	_, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("Org-ID is not number")
	}

	resp, err := p.cmdb.GetOrgClusterRelationsByOrg(ctx.Request().Context(), idStr)
	if err != nil {
		return "", err
	}
	clustersMap := make(map[string]bool, len(resp.Data))
	for _, item := range resp.Data {
		clustersMap[item.ClusterName] = true
	}
	clusters := req.URL.Query()["in_cluster_name"]
	if len(clusters) <= 0 {
		return "", fmt.Errorf("in_cluster_name must not be empty")
	}
	for _, cluster := range clusters {
		if !clustersMap[cluster] {
			return "", fmt.Errorf("not found cluster '%s'", cluster)
		}
	}
	return idStr, nil
}

func (p *provider) OrgIDByCluster(key string) permission.ValueGetter {
	return func(ctx context.Context, req interface{}) (string, error) {
		orgIdValue := permission.OrgIDValue()
		orgIdStr, err := orgIdValue(ctx, req)
		_, err = strconv.ParseUint(orgIdStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("Org-ID is not number")
		}
		request := http.ContextRequest(ctx)
		cluster := request.URL.Query().Get(key)
		if len(cluster) <= 0 {
			return "", fmt.Errorf("cluster must not be empty")
		}
		err = p.checkOrgIDsByCluster(ctx, orgIdStr, cluster)
		if err != nil {
			return "", err
		}
		return orgIdStr, nil
	}
}

func (p *provider) checkOrgIDsByCluster(ctx context.Context, orgID string, clusterName string) error {
	resp, err := p.cmdb.GetOrgClusterRelationsByOrg(ctx, orgID)
	if err != nil {
		return err
	}
	for _, item := range resp.Data {
		if item.ClusterName == clusterName {
			return nil
		}
	}
	return fmt.Errorf("not found cluster '%s'", clusterName)
}
