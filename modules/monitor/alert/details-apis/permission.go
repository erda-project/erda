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

package details_apis

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda-infra/providers/httpserver"
)

func (p *provider) getOrgIDByClusters(ctx httpserver.Context) (string, error) {
	req := ctx.Request()
	idStr := api.OrgID(req)
	orgID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return "", fmt.Errorf("Org-ID is not number")
	}

	resp, err := p.cmdb.QueryAllOrgClusterRelation()
	if err != nil {
		return "", err
	}
	clustersMap := make(map[string]bool, len(resp))
	for _, item := range resp {
		if item.OrgID == orgID {
			clustersMap[item.ClusterName] = true
		}
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
