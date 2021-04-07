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

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func EmitPipelineEvent(p *spec.Pipeline, userID string) {
	event := &PipelineEvent{DefaultEvent: defaultEvent}

	// EventHeader
	event.EventHeader.Event = string(EventKindPipeline)
	event.EventHeader.Action = string(p.Status)

	event.EventHeader.ApplicationID = p.Labels[apistructs.LabelAppID]
	event.EventHeader.ProjectID = p.Labels[apistructs.LabelProjectID]
	event.EventHeader.OrgID = p.Labels[apistructs.LabelOrgID]
	event.EventHeader.Env = p.Extra.DiceWorkspace.String()

	// Identity
	event.UserID = userID
	event.InternalClient = p.Extra.InternalClient

	// Pipeline
	event.Pipeline = p

	mgr.ch <- event
}

func EmitTaskEvent(task *spec.PipelineTask, p *spec.Pipeline) {
	event := &PipelineTaskEvent{DefaultEvent: defaultEvent}

	// EventHeader
	event.EventHeader.Event = string(EventKindPipelineTask)
	event.EventHeader.Action = string(task.Status)

	event.EventHeader.ApplicationID = p.Labels[apistructs.LabelAppID]
	event.EventHeader.ProjectID = p.Labels[apistructs.LabelProjectID]
	event.EventHeader.OrgID = p.Labels[apistructs.LabelOrgID]
	event.EventHeader.Env = p.Extra.DiceWorkspace.String()

	// Identity
	event.UserID = p.GetRunUserID()
	event.InternalClient = p.Extra.InternalClient

	// Task
	event.Task = task

	// Pipeline
	event.Pipeline = p

	mgr.ch <- event
}

func EmitTaskRuntimeEvent(task *spec.PipelineTask, p *spec.Pipeline) {
	event := &PipelineTaskRuntimeEvent{DefaultEvent: defaultEvent}

	// EventHeader
	event.EventHeader.Event = string(EventKindPipelineTaskRuntime)
	event.EventHeader.Action = "update"

	event.EventHeader.ApplicationID = p.Labels[apistructs.LabelAppID]
	event.EventHeader.ProjectID = p.Labels[apistructs.LabelProjectID]
	event.EventHeader.OrgID = p.Labels[apistructs.LabelOrgID]
	event.EventHeader.Env = p.Extra.DiceWorkspace.String()

	// Identity
	event.UserID = p.GetRunUserID()
	event.InternalClient = p.Extra.InternalClient

	// Task
	event.Task = task

	// Pipeline
	event.Pipeline = p

	// RuntimeID
	event.RuntimeID = task.RuntimeID()

	mgr.ch <- event
}
