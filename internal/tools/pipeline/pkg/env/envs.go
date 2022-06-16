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

package env

import (
	"strings"
)

type ParamSource string

const (
	SystemParamSource ParamSource = "system"
	UserParamSource   ParamSource = "user"
)

type TaskParam struct {
	Name   string      `json:"name"`
	Value  string      `json:"value"`
	Source ParamSource `json:"source"`
}

// some pipeline basic envs
const (
	PublicEnvPipelineID        = "PIPELINE_ID"
	PublicEnvTaskID            = "PIPELINE_TASK_ID"
	PublicEnvTaskName          = "PIPELINE_TASK_NAME"
	PublicEnvTaskLogID         = "PIPELINE_TASK_LOG_ID"
	PublicEnvPipelineDebugMode = "PIPELINE_DEBUG_MODE"
	PublicEnvPipelineTimeBegin = "PIPELINE_TIME_BEGIN_TIMESTAMP"
)

const (
	EnvPipelineSecretPrefix = "PIPELINE_SECRET_"
)

// TryRevertPipelineEnvKey try to revert pipeline env key, to lower and replace "_" to "."
func TryRevertPipelineEnvKey(envPrefix string, key string) string {
	return strings.Replace(
		strings.ToLower(
			strings.TrimPrefix(key, envPrefix)),
		"_", ".", -1)
}

func GenPipelineEnvKey(envPrefix string, key string) string {
	return envPrefix +
		strings.Replace(
			strings.Replace(
				strings.ToUpper(key),
				".", "_", -1),
			"-", "_", -1)
}

func GetTaskSourceParamsFromEnv(envPrefix string, envs map[string]string) map[string]string {
	params := make(map[string]string)
	for k, v := range envs {
		if strings.HasPrefix(k, envPrefix) {
			params[TryRevertPipelineEnvKey(envPrefix, k)] = v
		}
	}
	return params
}
