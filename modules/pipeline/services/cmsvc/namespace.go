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
