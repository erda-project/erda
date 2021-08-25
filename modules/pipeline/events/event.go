// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package events

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
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
	dbClient *dbclient.Client
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
func (*DefaultEvent) HandleDB() error        { return nil }

const (
	SenderPipeline = "pipeline"
)

type EventKind string

const (
	EventKindPipeline            EventKind = "pipeline"
	EventKindPipelineTask        EventKind = "pipeline_task"
	EventKindPipelineTaskRuntime EventKind = "pipeline_task_runtime"
	EventKindPipelineStream      EventKind = "pipeline_stream"
)
