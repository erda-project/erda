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

package cms

import (
	"fmt"

	"github.com/erda-project/erda/modules/pkg/gitflowutil"
)

const PipelineAppConfigNameSpacePrefix = "pipeline-secrets-app"

func MakeAppDefaultSecretNamespace(appID string) string {
	return fmt.Sprintf("%s-%s-default", PipelineAppConfigNameSpacePrefix, appID)
}

func MakeAppBranchPrefixSecretNamespace(appID, branch string) (string, error) {
	branchPrefix, err := gitflowutil.GetReferencePrefix(branch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", PipelineAppConfigNameSpacePrefix, appID, branchPrefix), nil
}
