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

package base

import (
	"fmt"
	"strings"
)

const (
	componentProviderNamePrefix = "component-protocol.components."
)

func MustGetScenarioAndCompNameFromProviderKey(providerKey string) (scenario, compName string) {
	scenario, compName, err := GetScenarioAndCompNameFromProviderKey(providerKey)
	if err != nil {
		panic(err)
	}

	return scenario, compName
}

func GetScenarioAndCompNameFromProviderKey(providerKey string) (scenario, compName string, err error) {
	if !strings.HasPrefix(providerKey, componentProviderNamePrefix) {
		return "", "", fmt.Errorf("invalid prefix")
	}
	ss := strings.SplitN(providerKey, ".", 4)
	if len(ss) != 4 {
		return "", "", fmt.Errorf("not standard provider key: %s", providerKey)
	}
	vv := strings.SplitN(ss[3], "@", 2)
	if len(vv) == 2 {
		compName = vv[1]
	} else {
		compName = ss[3]
	}
	return ss[2], compName, nil
}

func MakeComponentProviderName(scenario, compType string) string {
	return fmt.Sprintf("%s%s.%s", componentProviderNamePrefix, scenario, compType)
}
