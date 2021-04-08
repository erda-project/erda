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
