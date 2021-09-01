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

package cmp_dashboard_workload_detail

import (
	"fmt"
	"strings"

	"github.com/erda-project/erda/apistructs"
)

func ParseWorkloadID(id string) (apistructs.K8SResType, string, string, error) {
	splits := strings.Split(id, "_")
	if len(splits) != 3 {
		return "", "", "", fmt.Errorf("invalid workload id: %s", id)
	}
	return apistructs.K8SResType(splits[0]), splits[1], splits[2], nil
}
