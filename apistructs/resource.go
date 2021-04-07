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

// /api/scheduler/resource
// method: get
// 查询调度资源
type SchedulerResourceFecthRequest struct {
	Cluster  string            `json:"cluster"`
	Resource SchedulerResource `json:"resource"`
	Attr     Attribute         `json:"attribute"`
	Extra    map[string]string `json:"extra,omitempty"`
}

type SchedulerResource struct {
	CPU  float64 `json:"cpus"`
	Mem  float64 `json:"mem"`
	Disk float64 `json:"disk"`
}

// Attribute dice_tags like & unlike
type Attribute struct {
	Likes   []string `json:"like"`
	UnLikes []string `json:"unlike"`
}
