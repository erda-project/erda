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

type ScheduleLabelListRequest struct{}
type ScheduleLabelListResponse struct {
	Header
	Data ScheduleLabelListData `json:"data"`
}
type ScheduleLabelListData struct {
	// map-key: label name
	// map-value: is this label a prefix?
	Labels map[string]bool `json:"labels"`
}

type ScheduleLabelSetRequest struct {
	// 对于 dcos 的 tag, 由于只有 key, 则 tag 中的 value 都为空
	Tags        map[string]string `json:"tag"`
	Hosts       []string          `json:"hosts"`
	ClusterName string            `json:"clustername"`
	ClusterType string            `json:"clustertype"`
	SoldierURL  string            `json:"soldierURL"`
}

type ScheduleLabelSetResponse struct {
	Header
}
