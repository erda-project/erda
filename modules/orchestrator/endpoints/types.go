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

package endpoints

type RuntimeStatusEventReq struct {
	RuntimeName     string               `json:"runtimeName"`
	IsDeleted       bool                 `json:"isDeleted"`
	EventType       string               `json:"eventType"` // total(全量事件)/increment(增量事件)
	ServiceStatuses []ServiceStatusEvent `json:"serviceStatuses"`
}

type ServiceStatusEvent struct {
	ServiceName      string                `json:"serviceName"`
	Status           string                `json:"serviceStatus"`
	Replica          int                   `json:"replica"`
	InstanceStatuses []InstanceStatusEvent `json:"instanceStatuses"`
}

type InstanceStatusEvent struct {
	TaskId string                 `json:"id"` // TaskId
	IP     string                 `json:"ip"`
	Status string                 `json:"instanceStatus"`
	Stage  string                 `json:"stage"`
	Extra  map[string]interface{} `json:"extra"`
}
