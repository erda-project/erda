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

package regex

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/apps/msp/apm/log-service/analysis/processors"
)

func Test_Process_With_ValidParams_Should_Success(t *testing.T) {
	var (
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
		"replaceKey": map[string]string{
			"level": "_level",
		},
	})
	p, _ := New(metricName, cfg)

	metric, fields, appendTags, replaceKey, err := p.(processors.Processor).Process("123")
	if metric != metricName {
		t.Errorf("Process error, expect metricName: %s, but got %s", metricName, metric)
	}
	if len(fields) == 0 {
		t.Errorf("fields should not empty")
	}
	if len(appendTags) == 0 {
		t.Errorf("appendTags should not empty")
	}
	if len(replaceKey) == 0 {
		t.Errorf("replaceKey should not empty")
	}
	metric, fields, appendTags, replaceKey, err = p.Process("abc")
	if err != ErrNotMatch {
		t.Errorf("should miss match")
	}

}

func Test_Process_With_InvalidLenOfKeys_Should_Fail(t *testing.T) {
	var (
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
	p, _ := New(metricName, cfg)

	_, _, _, _, err := p.(processors.Processor).Process("abc")
	if err != ErrNotMatch {
		t.Errorf("should miss match")
	}
}

func Test_Process_With_InvalidTypeOfKeys_Should_Fail(t *testing.T) {
	var (
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
	p, _ := New(metricName, cfg)

	_, _, _, _, err := p.(processors.Processor).Process("abc")
	if err != ErrNotMatch {
		t.Errorf("should miss match")
	}
}

func Test_processor_initRegexps(t *testing.T) {
	p := &processor{}

	// use default
	regexps, err := p.initRegexps("(.*)")
	assert.NoError(t, err)
	assert.NotNil(t, regexps.defaultReg)
	assert.Nil(t, regexps.zwaReg)

	// use zwa
	regexps, err = p.initRegexps("abc(?=de)")
	assert.NoError(t, err)
	assert.Nil(t, regexps.defaultReg)
	assert.NotNil(t, regexps.zwaReg)

	// invalid
	regexps, err = p.initRegexps("abc(?=de")
	assert.Error(t, err)
	assert.Nil(t, regexps.defaultReg)
	assert.Nil(t, regexps.zwaReg)
}
