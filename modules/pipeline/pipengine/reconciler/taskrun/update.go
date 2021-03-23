package taskrun

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/pkg/loop"
)

func (tr *TaskRun) fetchLatestTask() error {
	latest, err := tr.DBClient.GetPipelineTask(tr.Task.ID)
	if err != nil {
		return err
	}
	*(tr.Task) = *(&latest)
	return nil
}

// Update must update without error
func (tr *TaskRun) Update() {
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: start update")
	defer rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: end update")

	// db
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: start update task to db")
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		if err := tr.DBClient.UpdatePipelineTask(tr.Task.ID, tr.Task); err != nil {
			rlog.TWarnf(tr.P.ID, tr.Task.ID, "failed to update taskRun, err: %v, will continue until update success", err)
			return false, err
		}
		return true, nil
	})
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: end update task to db")

	// event
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: start emit task event")
	events.EmitTaskEvent(tr.Task, tr.P)
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: end emit task event")

	// metrics
	rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: start emit task metrics")
	defer rlog.TDebugf(tr.Task.PipelineID, tr.Task.ID, "taskRun: end emit task metrics")
	// go metrics.TaskCounterTotalAdd(*tr.Task, 1)
}

func (tr *TaskRun) AppendLastMsg(msg string) error {
	if msg == "" {
		return nil
	}
	if err := tr.fetchLatestTask(); err != nil {
		return err
	}
	tr.Task.Result.Errors = append(tr.Task.Result.Errors, apistructs.ErrorResponse{Msg: msg})
	if err := tr.DBClient.UpdatePipelineTaskResult(tr.Task.ID, tr.Task.Result); err != nil {
		logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q append last message failed, err: %v",
			tr.P.ID, tr.Task.Name, err)
		return err
	}
	return nil
}

func (tr *TaskRun) fetchLatestPipelineStatus() error {
	status, err := tr.DBClient.GetPipelineStatus(tr.P.ID)
	if err != nil {
		return err
	}
	tr.QueriedPipelineStatus = status
	return nil
}

func (tr *TaskRun) EnsureFetchLatestPipelineStatus() {
	var latestPStatus apistructs.PipelineStatus
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		latestPStatus, err = tr.DBClient.GetPipelineStatus(tr.P.ID)
		if err != nil {
			rlog.TWarnf(tr.P.ID, tr.Task.ID, "failed to get latest pipeline status, err: %v, continue fetch", err)
			return false, err
		}
		return true, nil
	})
	tr.QueriedPipelineStatus = latestPStatus
}
