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
