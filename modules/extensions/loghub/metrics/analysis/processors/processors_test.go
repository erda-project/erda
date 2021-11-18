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

package processors_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors/regex"
	_ "github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors/regex" //
)

func TestExampleProcessors(t *testing.T) {
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
		"pattern": "(\\d+)",
		"keys": []*pb.FieldDefine{
			{
				Key:  "ip",
				Type: "string",
			},
		},
		"appendTags": map[string]string{
			"append_tag_1": "value_1",
		},
	})
	ps := processors.New()
	err := ps.Add(scopeID, tags, metricName, "regexp", cfg)
	if err != nil {
		t.Errorf("Add error: %s", err)
		return
	}
	list := ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
	})
	if len(list) != 0 {
		t.Errorf("Find error")
		return
	}
	list = ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "3",
		"dice_service_name":   "abc",
	})
	if len(list) != 0 {
		t.Errorf("Find error")
		return
	}
	list = ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
		"dice_service_name":   "abc",
	})
	if len(list) <= 0 {
		t.Errorf("Find error")
		return
	}
	fmt.Printf("Find %d\n", len(list))

	// Output:
	// Find 1

	metric, fields, appendTags, err := list[0].Process("123")
	if metric != metricName {
		t.Errorf("Process error, expect metricName: %s, but got %s", metricName, metric)
	}
	if len(fields) == 0 {
		t.Errorf("fields should not empty")
	}
	if len(appendTags) == 0 {
		t.Errorf("appendTags should not empty")
	}

	metric, fields, appendTags, err = list[0].Process("abc")
	if err != regex.ErrNotMatch {
		t.Errorf("should miss match")
	}

}

func Test_ExampleProcessors_With_InvalidLenOfKeys_Should_Fail(t *testing.T) {
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
		"pattern": "(\\d+)",
		"keys": []*pb.FieldDefine{
			{
				Key:  "ip",
				Type: "string",
			},
			{
				Key:  "extra",
				Type: "string",
			},
		},
		"appendTags": map[string]string{
			"append_tag_1": "value_1",
		},
	})
	ps := processors.New()
	err := ps.Add(scopeID, tags, metricName, "regexp", cfg)

	list := ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
		"dice_service_name":   "abc",
	})
	if len(list) <= 0 {
		t.Errorf("Find error")
		return
	}
	fmt.Printf("Find %d\n", len(list))

	// Output:
	// Find 1

	_, _, _, err = list[0].Process("abc")
	if err != regex.ErrNotMatch {
		t.Errorf("should miss match")
	}
}

func Test_ExampleProcessors_With_InvalidTypeOfKeys_Should_Fail(t *testing.T) {
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
		"pattern": "(\\S+)",
		"keys": []*pb.FieldDefine{
			{
				Key:  "ip",
				Type: "int",
			},
		},
		"appendTags": map[string]string{
			"append_tag_1": "value_1",
		},
	})
	ps := processors.New()
	err := ps.Add(scopeID, tags, metricName, "regexp", cfg)

	list := ps.Find("", scopeID, map[string]string{
		"dice_org_id":         "1",
		"dice_application_id": "2",
		"dice_service_name":   "abc",
	})
	if len(list) <= 0 {
		t.Errorf("Find error")
		return
	}
	fmt.Printf("Find %d\n", len(list))

	// Output:
	// Find 1

	_, _, _, err = list[0].Process("abc")
	if err != regex.ErrNotMatch {
		t.Errorf("should miss match")
	}
}
