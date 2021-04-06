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

package metrics

import (
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func taskErrorLog(task spec.PipelineTask, format string, args ...interface{}) {
	logrus.WithField("type", "metrics").WithField("pipelineID", task.PipelineID).WithField("taskID", task.ID).Errorf(format, args...)
}

func taskDebugLog(task spec.PipelineTask, format string, args ...interface{}) {
	logrus.WithField("type", "metrics").WithField("pipelineID", task.PipelineID).WithField("taskID", task.ID).Debugf(format, args...)
}

func pipelineErrorLog(p spec.Pipeline, format string, args ...interface{}) {
	logrus.WithField("type", "metrics").WithField("pipelineID", p.ID).Errorf(format, args...)
}

func pipelineDebugLog(p spec.Pipeline, format string, args ...interface{}) {
	logrus.WithField("type", "metrics").WithField("pipelineID", p.ID).Debugf(format, args...)
}
