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

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	AliyunRegistry = "aliyun-registry"
)

// TaskInspect inspect k8s job, like kubectl describe
// contains task object and events
type TaskInspect struct {
	Object interface{}       `json:"object"`
	Events *corev1.EventList `json:"events"`
	Desc   string            `json:"desc"`
}
