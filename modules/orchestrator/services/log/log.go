package log

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

type DeployLogHelper struct {
	DeploymentID uint64
	Bdl          *bundle.Bundle
}

func (d *DeployLogHelper) Log(content string) {
	content = "(orchestrator) " + content
	logrus.Debugf("deployment log -> %s", content)
	timestamp := time.Now().UnixNano()
	line := fmt.Sprintf("%s\n", content)
	lines := []apistructs.LogPushLine{
		{Source: "deploy", ID: strconv.FormatUint(d.DeploymentID, 10), Content: line, Timestamp: timestamp},
	}
	// TODO: buffer
	if err := d.Bdl.PushLog(&apistructs.LogPushRequest{Lines: lines}); err != nil {
		logrus.Errorf("[alert] failed to pushLog, deploymentId: %d, last err: %v",
			d.DeploymentID, err)
	}
}
