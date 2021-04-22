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

package apistructs

import "time"

// PipelineEvent is k8s-event-like stream event.
type PipelineEvent struct {
	// Optional; this should be a short, machine understandable string that gives the reason
	// for this event being generated. For example, if the event is reporting that a container
	// can't start, the Reason might be "ImageNotFound".
	// +optional
	Reason string `json:"reason,omitempty"`

	// Optional. A human-readable description of the status of this operation.
	// +optional
	Message string `json:"message,omitempty"`

	// Optional. The component reporting this event. Should be a short machine understandable string.
	// +optional
	Source PipelineEventSource `json:"source,omitempty"`

	// The time at which the event was first recorded. (Time of server receipt is in TypeMeta.)
	// +optional
	FirstTimestamp time.Time `json:"firstTimestamp,omitempty"`

	// The time at which the most recent occurrence of this event was recorded.
	// +optional
	LastTimestamp time.Time `json:"lastTimestamp,omitempty"`

	// The number of times this event has occurred.
	// +optional
	Count int32 `json:"count,omitempty"`

	// Type of this event (Normal, Warning), new types could be added in the future.
	// +optional
	Type string `json:"type,omitempty"`
}

// PipelineEventSource represents the source from which an event is generated
type PipelineEventSource struct {
	// Component from which the event is generated.
	// +optional
	Component string `json:"component,omitempty"`

	// Node name on which the event is generated.
	// +optional
	Host string `json:"host,omitempty"`
}
