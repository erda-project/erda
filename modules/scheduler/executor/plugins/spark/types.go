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

package spark

type SparkCreateRequest struct {
	AppResource          string            `json:"appResource"`
	Action               string            `json:"action"`
	ClientSparkVersion   string            `json:"clientSparkVersion"`
	MainClass            string            `json:"mainClass,omitempty"`
	AppArgs              []string          `json:"appArgs"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	SparkProperties      map[string]string `json:"sparkProperties,omitempty"`
}
type SparkResponse struct {
	Action             string `json:"action"`
	ServerSparkVersion string `json:"serverSparkVersion"`
	SubmissionId       string `json:"submissionId,omitempty"`
	Success            bool   `json:"success,omitempty"`
	DriverState        string `json:"driverState,omitempty"`
	Message            string `json:"message,omitempty"`
}
