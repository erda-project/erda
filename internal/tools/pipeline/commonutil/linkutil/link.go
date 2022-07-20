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

package linkutil

import (
	"context"
	"fmt"
	"strconv"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/clusterinfo"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

func GetPipelineLink(orgClient org.ClientInterface, p spec.Pipeline) (bool, string) {
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
	orgResp, err := orgClient.GetOrg(apis.WithInternalClientContext(context.Background(), discover.Pipeline()), &orgpb.GetOrgRequest{
		IdOrName: strconv.FormatUint(orgID, 10),
	})
	if err != nil {
		return false, ""
	}
	org := orgResp.Data

	// get domain protocol
	clusterInfo, err := clusterinfo.GetClusterInfoByName(p.ClusterName)
	if err != nil {
		return false, ""
	}
	protocol := clusterInfo.CM.Get(apistructs.DICE_PROTOCOL)
	if protocol == "" {
		return false, ""
	}

	return true, fmt.Sprintf("%s://%s/workBench/projects/%d/apps/%d/pipeline/%d", protocol, org.Domain, projectID, appID, p.ID)
}
