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

func GenEnvKeyWithPrefix(envPrefix string, key string) string {
	return envPrefix + strings.NewReplacer(".", "_", "-", "_").Replace(strings.ToUpper(key))
}

func GenEnvKey(key string) string {
	return GenEnvKeyWithPrefix("", key)
}
