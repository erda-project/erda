package events

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type PipelineTaskEvent struct {
	DefaultEvent
	IdentityInfo
	EventHeader apistructs.EventHeader
	Task        *spec.PipelineTask
	Pipeline    *spec.Pipeline
}

func (e *PipelineTaskEvent) Kind() EventKind {
	return EventKindPipelineTask
}

func (e *PipelineTaskEvent) Header() apistructs.EventHeader {
	return e.EventHeader
}

func (e *PipelineTaskEvent) Sender() string {
	return SenderPipeline
}

func (e *PipelineTaskEvent) Content() interface{} {
	content := apistructs.PipelineTaskEventData{
		PipelineTaskID:  e.Task.ID,
		PipelineID:      e.Pipeline.ID,
		ActionType:      e.Task.Type,
		ClusterName:     e.Pipeline.ClusterName,
		Status:          string(e.Task.Status),
		UserID:          e.UserID,
		CreatedAt:       e.Task.TimeCreated,
		QueueTimeSec:    costtimeutil.CalculateTaskQueueTimeSec(e.Task),
		CostTimeSec:     costtimeutil.CalculateTaskCostTimeSec(e.Task),
		OrgName:         e.Pipeline.GetOrgName(),
		ProjectName:     e.Pipeline.NormalLabels[apistructs.LabelProjectName],
		ApplicationName: e.Pipeline.NormalLabels[apistructs.LabelAppName],
		TaskName:        e.Task.Name,
		RuntimeID:       e.Task.RuntimeID(),
		ReleaseID:       e.Task.ReleaseID(),
	}
	return content
}

func (e *PipelineTaskEvent) String() string {
	return fmt.Sprintf("event: %s, action: %s, pipelineID: %d, pipelineTaskID: %d",
		e.EventHeader.Event, e.EventHeader.Action, e.Pipeline.ID, e.Task.ID)
}

func (e *PipelineTaskEvent) HandleWebhook() error {
	req := &apistructs.EventCreateRequest{}
	req.Sender = SenderPipeline
	req.EventHeader = e.Header()
	req.Content = e.Content()

	return e.DefaultEvent.bdl.CreateEvent(req)
}

const (
	WSTypePipelineTaskStatusUpdate = "PIPELINE_TASK_STATUS_UPDATE"
)

type WSPipelineTaskStatusUpdatePayload struct {
	wsHeader
	PipelineTaskID uint64                        `json:"pipelineTaskID"`
	Status         apistructs.PipelineStatus     `json:"status"`
	Result         apistructs.PipelineTaskResult `json:"result"`

	CostTimeSec int64 `json:"costTimeSec"`
}

func (e *PipelineTaskEvent) HandleWebSocket() error {
	state := e.Task.Status
	if e.Task.Type == "manual-review" {
		state = e.Task.Status.ChangeStateForManualReview()
	}

	payload := WSPipelineTaskStatusUpdatePayload{}
	payload.PipelineTaskID = e.Task.ID
	payload.PipelineID = e.Pipeline.ID
	payload.ApplicationID = e.Pipeline.Labels[apistructs.LabelAppID]
	payload.ProjectID = e.Pipeline.Labels[apistructs.LabelProjectID]
	payload.OrgID = e.Pipeline.Labels[apistructs.LabelOrgID]
	payload.Status = state
	payload.Result = e.Task.Result
	payload.CostTimeSec = e.Content().(apistructs.PipelineTaskEventData).CostTimeSec

	wsEvent := websocket.Event{
		Scope: apistructs.Scope{
			Type: apistructs.AppScope,
			ID:   e.Header().ApplicationID,
		},
		Type:    WSTypePipelineTaskStatusUpdate,
		Payload: payload,
	}

	return e.DefaultEvent.wsClient.EmitEvent(context.Background(), wsEvent)
}
