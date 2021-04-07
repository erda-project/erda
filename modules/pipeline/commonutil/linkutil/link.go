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

package linkutil

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func GetPipelineLink(bdl *bundle.Bundle, p spec.Pipeline) (bool, string) {
	// org id
	orgID, err := strconv.ParseUint(p.Labels[apistructs.LabelOrgID], 10, 64)
	if err != nil {
		return false, ""
	}
	// project id
	projectID, err := strconv.ParseUint(p.Labels[apistructs.LabelProjectID], 10, 64)
	if err != nil {
		return false, ""
	}
	// app id
	appID, err := strconv.ParseUint(p.Labels[apistructs.LabelAppID], 10, 64)
	if err != nil {
		return false, ""
	}

	// get org domain
	org, err := bdl.GetOrg(orgID)
	if err != nil {
		return false, ""
	}

	// get domain protocol
	clusterInfo, err := bdl.QueryClusterInfo(p.ClusterName)
	if err != nil {
		return false, ""
	}
	protocol := clusterInfo.Get(apistructs.DICE_PROTOCOL)
	if protocol == "" {
		return false, ""
	}

	return true, fmt.Sprintf("%s://%s/workBench/projects/%d/apps/%d/pipeline/%d", protocol, org.Domain, projectID, appID, p.ID)
}
