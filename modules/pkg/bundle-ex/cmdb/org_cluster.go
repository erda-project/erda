// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmdb

import (
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
)

// OrgClusterRelationDTO 企业对应集群关系结构
type OrgClusterRelationDTO struct {
	ID          uint64    `json:"id"`
	OrgID       uint64    `json:"orgId"`
	OrgName     string    `json:"orgName"`
	ClusterID   uint64    `json:"clusterId"`
	ClusterName string    `json:"clusterName"`
	Creator     string    `json:"creator"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type orgClusterRelResp struct {
	apistructs.Header
	Data []*OrgClusterRelationDTO `json:"data"`
}

// QueryAllOrgClusterRelation 获取所有的企业集群关联关系
func (c *Cmdb) QueryAllOrgClusterRelation() ([]*OrgClusterRelationDTO, error) {
	var resp orgClusterRelResp
	hresp, err := c.hc.Get(c.url).
		Path("/api/orgs/clusters/relations").
		Header("User-ID", c.operatorID).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !hresp.IsOK() || !resp.Success {
		return nil, fmt.Errorf("status: %d, error: %v", hresp.StatusCode(), resp.Error)
	}
	return resp.Data, nil
}
