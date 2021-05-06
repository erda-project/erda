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
	"time"
)

type PipelineGCInfo struct {
	CreatedAt time.Time `json:"createdAt,omitempty"`
	GCAt      time.Time `json:"gcAt,omitempty"`
	TTL       uint64    `json:"ttl,omitempty"`
	LeaseID   string    `json:"leaseID,omitempty"`
	Data      []byte    `json:"data,omitempty"`
}

func MakePipelineGCInfo(ttl uint64, leaseID string, data []byte) PipelineGCInfo {
	now := time.Now()
	return PipelineGCInfo{
		CreatedAt: now,
		GCAt:      now.Add(time.Second * time.Duration(ttl)),
		TTL:       ttl,
		LeaseID:   leaseID,
		Data:      data,
	}
}

type PipelineGCDBOption struct {
	NeedArchive bool `json:"needArchive"`
}
