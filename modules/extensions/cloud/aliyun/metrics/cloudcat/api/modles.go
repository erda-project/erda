// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package api

// Metric .
type Metric struct {
	Name      string                 `json:"name"`
	Timestamp uint64                 `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

// OrgInfo .
type OrgInfo struct {
	OrgId        string
	OrgName      string
	AccessKey    string
	AccessSecret string
}

// ProjectMeta .
type ProjectMeta struct {
	Description string
	Labels      map[string]interface{}
	Namespace   string
}

type MetricMeta struct {
}

type DataPoints string

// type DataPoint struct {
// 	Timestamp  uint64      `json:"timestamp"`
// 	UserId     string      `json:"userId"`
// 	InstanceId string      `json:"instanceId"`
// 	Average    interface{} `json:"Average,omitempty"`
// 	Maximum    interface{} `json:"Maximum,omitempty"`
// 	Minimum    interface{} `json:"Minimum,omitempty"`
// 	Value      interface{} `json:"Value,omitempty"`
// }
