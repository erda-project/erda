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
