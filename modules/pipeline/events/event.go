package events

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pkg/websocket"
)

type Event interface {
	Kind() EventKind
	Header() apistructs.EventHeader
	Sender() string
	Content() interface{}
	String() string
	Hook
}

type DefaultEvent struct {
	bdl      *bundle.Bundle
	wsClient *websocket.Publisher
}

type IdentityInfo struct {
	UserID         string `json:"userID"`
	InternalClient string `json:"internalClient"`
}

func (*DefaultEvent) HandleWebhook() error   { return nil }
func (*DefaultEvent) HandleWebSocket() error { return nil }
func (*DefaultEvent) HandleDingDing() error  { return nil }
func (*DefaultEvent) HandleHTTP() error      { return nil }

const (
	SenderPipeline = "pipeline"
)

type EventKind string

const (
	EventKindPipeline            EventKind = "pipeline"
	EventKindPipelineTask        EventKind = "pipeline_task"
	EventKindPipelineTaskRuntime EventKind = "pipeline_task_runtime"
)
