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
