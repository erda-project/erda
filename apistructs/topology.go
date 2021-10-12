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

package apistructs

type ServiceDashboardResponse struct {
	Header
	Data ServiceDashboardData
}

type ServiceDashboardData struct {
	Data []ServiceDashboard
}

type ServiceDashboard struct {
	Id              string  `json:"service_id"`
	Name            string  `json:"service_name"`
	ReqCount        int64   `json:"req_count"`
	ReqErrorCount   int64   `json:"req_error_count"`
	ART             float64 `json:"avg_req_time"`           // avg response time
	RSInstanceCount string  `json:"running_instance_count"` // running / stopped
	RuntimeId       string  `json:"runtime_id"`
	RuntimeName     string  `json:"runtime_name"`
	ApplicationId   string  `json:"application_id"`
	ApplicationName string  `json:"application_name"`
}
