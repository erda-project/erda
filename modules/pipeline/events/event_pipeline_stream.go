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
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

// Valid values for event levels (new level could be added in future)
const (
	// EventLevelNormal represents information only and will not cause any problems
	EventLevelNormal string = "Normal"
	// EventLevelWarning represents events are to warn that something might go wrong
	EventLevelWarning string = "Warning"
)

type PipelineStreamEvent struct {
	DefaultEvent
	IdentityInfo
	EventHeader apistructs.EventHeader
	PipelineID  uint64
	Events      []*apistructs.PipelineEvent
}

func (e *PipelineStreamEvent) Kind() EventKind {
	return EventKindPipelineStream
}

func (e *PipelineStreamEvent) Header() apistructs.EventHeader {
	return e.EventHeader
}

func (e *PipelineStreamEvent) Sender() string {
	return SenderPipeline
}

func (e *PipelineStreamEvent) Content() interface{} {
	return e.Events
}

func (e *PipelineStreamEvent) String() string {
	eventJson, _ := json.Marshal(e.Events)
	return fmt.Sprintf("event: %s, action: %s, pipelineID: %d, event: %s",
		e.EventHeader.Event, e.EventHeader.Action, e.PipelineID, eventJson)
}

func (e *PipelineStreamEvent) HandleDB() error {
	err := e.dbClient.AppendPipelineEvent(e.PipelineID, e.Events)
	if err != nil {
		logrus.Errorf("pipeline event: failed to handleDB, pipelineID: %d, err: %v", e.PipelineID, err)
	}
	return nil
}
