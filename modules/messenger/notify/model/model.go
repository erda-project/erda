// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import "time"

type BaseModel struct {
	ID        int64     `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"createdAt" gorm:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"updated_at"`
}

type FilterStatusRequest struct {
	OrgId     int
	ScopeType string
	ScopeId   string
	StartTime string
	EndTime   string
}

type FilterStatusResult struct {
	Status string
	Count  int64
}

type NotifyValue struct {
	Field string
	Count int64
}

type QueryNotifyHistoriesRequest struct {
	PageNo      int64  `json:"pageNo"`
	PageSize    int64  `json:"pageSize"`
	Channel     string `json:"channel"`
	NotifyName  string `json:"notifyName"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Label       string `json:"label"`
	ClusterName string `json:"clusterName"`
	OrgID       int64  `json:"orgID"`
}
