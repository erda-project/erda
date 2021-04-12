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

package spec

import (
	"time"
)

type CIV3BuildCache struct {
	ID          int64     `json:"id" xorm:"pk autoincr"`
	Name        string    `json:"name"`
	ClusterName string    `json:"clusterName"`
	LastPullAt  time.Time `json:"lastPullAt"`
	CreatedAt   time.Time `json:"createdAt" xorm:"created"`
	UpdatedAt   time.Time `json:"updatedAt" xorm:"updated"`
	DeletedAt   time.Time `xorm:"deleted"`
}

func (*CIV3BuildCache) TableName() string {
	return "ci_v3_build_caches"
}
