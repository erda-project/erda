package cmsvc

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pkg/gitflowutil"
)

func (s *CMSvc) MakeDefaultSecretNamespace(appID string) string {
	return fmt.Sprintf("%s-%s-default", apistructs.PipelineAppConfigNameSpacePreFix, appID)
}

func (s *CMSvc) MakeBranchPrefixSecretNamespace(appID, branch string) (string, error) {
	branchPrefix, err := gitflowutil.GetReferencePrefix(branch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", apistructs.PipelineAppConfigNameSpacePreFix, appID, branchPrefix), nil
}
