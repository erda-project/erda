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

package actionagentsvc

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

const agentDownloadScript = "download_file_from_image"

func (s *ActionAgentSvc) downloadAgent(clusterInfo apistructs.ClusterInfoData, agentImage string, agentMD5 string) error {
	scriptParams := map[string]string{
		"IMAGE_NAME":      agentImage,
		"IMAGE_FILE_PATH": "/opt/action/agent",
		"FILE_NAME":       "agent",
		"EXPECT_MD5":      agentMD5,
		"EXECUTABLE":      "true",
	}
	err := RunScript(clusterInfo, agentDownloadScript, scriptParams)
	if err != nil {
		logrus.Errorf("[alert] failed to download action agent, cluster: %s, err: %v", clusterInfo.Get(apistructs.DICE_CLUSTER_NAME), err)
		return apierrors.ErrDownloadActionAgent.InternalError(err)
	}
	return nil
}
