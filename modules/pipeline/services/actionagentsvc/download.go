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
