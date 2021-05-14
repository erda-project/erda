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

package processors_test

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors/regex" //
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

func ExampleProcessors() {
	var (
		scopeID string = "terminus"
		tags           = map[string]string{
			"dice_org_id":         "1",
			"dice_application_id": "2",
			"dice_service_name":   "abc",
		}
		metricName = "test_metric"
	)
	cfg, _ := json.Marshal(map[string]interface{}{
		"pattern": "(d+)",
		"keys": []*metrics.FieldDefine{
			{
				Key:  "ip",
				Type: "string",
			},
		},
	})
	ps := processors.New()
	err := ps.Add(scopeID, tags, metricName, "regexp", cfg)
	if err != nil {
		fmt.Println(err)
		return
	}
	list := ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
	})
	if len(list) != 0 {
		fmt.Println("Find error")
		return
	}
	list = ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "3",
		"dice_service_name":   "abc",
	})
	if len(list) != 0 {
		fmt.Println("Find error")
		return
	}
	list = ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
		"dice_service_name":   "abc",
	})
	if len(list) <= 0 {
		fmt.Println("Find error")
		return
	}
	fmt.Printf("Find %d\n", len(list))

	// Output:
	// Find 1
}
