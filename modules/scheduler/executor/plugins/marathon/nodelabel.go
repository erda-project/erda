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

package marathon

import (
	"fmt"

	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
)

func (m *Marathon) SetNodeLabels(setting executortypes.NodeLabelSetting, hosts []string, labels map[string]string) error {
	return fmt.Errorf("SetNodeLabels not implemented in marathon")
}

func convertLabels(key string) string {
	switch key {
	case "pack-job":
		return "pack"
	case "bigdata-job":
		return "bigdata"
	case "stateful-service":
		return "service-stateful"
	case "stateless-service":
		return "service-stateless"
	default:
		return key
	}
}
