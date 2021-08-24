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

	// if credential content is empty, use the latest credential data.
	// if credential change to agent from other type, clear credential info
	if req.Credential.Content != "" || (mc.Type != apistructs.ManageProxy && strings.ToUpper(req.CredentialType) == ProxyType) {
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
