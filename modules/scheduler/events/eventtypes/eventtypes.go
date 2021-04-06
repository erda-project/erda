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

package eventtypes

type StatusEvent struct {
	Type string `json:"type"`
	// ID由{runtimeName}.{serviceName}.{dockerID}生成
	ID      string `json:"id,omitempty"`
	IP      string `json:"ip,omitempty"`
	Status  string `json:"status"`
	TaskId  string `json:"taskId,omitempty"`
	Cluster string `json:"cluster,omitempty"`
	Host    string `json:"host,omitempty"`
	Message string `json:"message,omitempty"`
}
