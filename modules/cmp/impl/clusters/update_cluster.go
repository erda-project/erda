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

package clusters

import (
	"net/http"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (c *Clusters) UpdateCluster(req apistructs.CMPClusterUpdateRequest, header http.Header) error {
	var (
		mc  *apistructs.ManageConfig
		err error
	)

	cluster, err := c.bdl.GetCluster(req.ClusterUpdateRequest.Name)
	if err != nil {
		return err
	}

	mc = cluster.ManageConfig

	// if credential content is empty, use latest credential data.
	if req.Credential.Content != "" || strings.ToUpper(req.CredentialType) == ProxyType {
		// parse manage config from credential info
		mc, err = ParseManageConfigFromCredential(req.CredentialType, req.Credential)
		if err != nil {
			return err
		}
	}

	var newSchedulerConfig *apistructs.ClusterSchedConfig

	if req.Type != apistructs.EDAS {
		newSchedulerConfig = cluster.SchedConfig
		newSchedulerConfig.CPUSubscribeRatio = req.SchedulerConfig.CPUSubscribeRatio
	} else {
		newSchedulerConfig = req.SchedulerConfig
	}

	// TODO: support tag switch, current force true
	// e.g. modules/scheduler/impl/cluster/hook.go line:136
	newSchedulerConfig.EnableTag = true

	return c.bdl.UpdateCluster(apistructs.ClusterUpdateRequest{
		Name:            cluster.Name,
		DisplayName:     req.DisplayName,
		Type:            cluster.Type,
		Description:     req.Description,
		WildcardDomain:  req.WildcardDomain,
		SchedulerConfig: newSchedulerConfig,
		ManageConfig:    mc,
	}, map[string][]string{
		httputil.InternalHeader: {"cmp"},
	})
}
