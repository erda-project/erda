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

// PipelineAppliedResources represents multi-kind applied resources.
type PipelineAppliedResources struct {

	// Limits is the minimal enough resource to run pipeline
	// calculate minResource
	// 1 2 (2)
	// 2 3 (3)
	// 4   (4)
	// => max(1,2,2,3,4) = 4
	Limits PipelineAppliedResource `json:"limits"`

	// Requests is the minimal resource to run pipeline
	// calculate maxResource
	// 1 2 (3)
	// 2 3 (5)
	// 4   (4)
	// => max((1+2), (2+3), (4)) = 5
	Requests PipelineAppliedResource `json:"requests"`
}

// PipelineAppliedResource represents one kind of applied resources.
type PipelineAppliedResource struct {
	CPU      float64 `json:"cpu"`
	MemoryMB float64 `json:"memoryMB"`
}
