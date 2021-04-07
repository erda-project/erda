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
