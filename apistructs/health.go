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

type HealthStatus string

const (
	HealthStatusOk      HealthStatus = "ok"
	HealthStatusWarning HealthStatus = "warning"
	HealthStatusFail    HealthStatus = "fail"
)

// HealthResponse the response struct of health check
type HealthResponse struct {
	Name    string            `json:"name"`
	Status  HealthStatus      `json:"status"`
	Modules []Module          `json:"modules"`
	Tags    map[string]string `json:"tags"`
}

type Module struct {
	Name    string       `json:"name"`
	Status  HealthStatus `json:"status"`
	Message string       `json:"message"`
}
