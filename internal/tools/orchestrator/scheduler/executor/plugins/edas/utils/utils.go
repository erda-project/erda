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

package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

func AppendCommonHeaders(source map[string]string) map[string]string {
	source["Pragma"] = "no-cache"
	source["Cache-Control"] = "no-cache"
	source["Connection"] = "keep-alive"
	return source
}

func CombineEDASAppGroup(sgType, sgID string) string {
	return fmt.Sprintf("%s-%s", sgType, sgID)
}

func CombineEDASAppNameWithGroup(group string, serviceName string) string {
	return fmt.Sprintf("%s-%s", group, serviceName)
}

func CombineEDASAppName(sgType, sgID, serviceName string) string {
	return CombineEDASAppNameWithGroup(CombineEDASAppGroup(sgType, sgID), serviceName)
}

func CombineEDASAppInfo(sgType, sgID, serviceName string) (string, string) {
	group := CombineEDASAppGroup(sgType, sgID)
	return group, CombineEDASAppNameWithGroup(group, serviceName)
}

func MakeEnvVariableName(str string) string {
	return strings.ToUpper(strings.Replace(str, "-", "_", -1))
}

func EnvToString(envs map[string]string) (string, error) {
	type appEnv struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	var appEnvs []appEnv
	var keys []string

	for env := range envs {
		keys = append(keys, env)
	}
	sort.Strings(keys)

	for _, k := range keys {
		appEnvs = append(appEnvs, appEnv{
			Name:  k,
			Value: envs[k],
		})
	}

	res, err := json.Marshal(appEnvs)
	if err != nil {
		return "", errors.Wrapf(err, "failed to json marshal map env")
	}

	return string(res), nil
}
