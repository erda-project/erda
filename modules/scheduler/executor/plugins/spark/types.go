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
