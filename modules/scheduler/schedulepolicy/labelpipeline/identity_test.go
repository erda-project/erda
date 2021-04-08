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

package labelpipeline

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
)

// The test cases that are not enabled for tag scheduling are placed in policy_test.go for execution

// The flag bit ("ENABLETAG": "true") has been set to enable tag scheduling,
// And when the runtime or job does not carry any label
// The cluster configuration has only basic configuration
func TestIdentityLabelFilter1(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		Label:          make(map[string]string),
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	// Temporary support for the project tag
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	// The flag of any is set to true, but the any label is only combined with the label in Likes, and Likes is empty, so any will not appear in the constraints
	assert.True(t, result.Flag)
}

// The flag bit ("ENABLETAG": "true") has been set to enable tag scheduling,
// And when the runtime or job has a label set
// The cluster configuration has only basic configuration
func TestIdentityLabelFilter2(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2
	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "dev",
			"SERVICE_TYPE":   "STATELESS"},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"service-stateless"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	// Temporary support for the project tag
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// Test addons stateful service
func TestIdentityLabelFilter2B(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2
	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"SERVICE_TYPE": "ADDONS"},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"service-stateful"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	// Temporary support for the project tag
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// The flag bit ("ENABLETAG": "true") has been set to enable tag scheduling,
// And when the runtime or job has a label set
// Cluster configuration and fine configuration
func TestIdentityLabelFilter3(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true"
    },
    "optionsPlus": {
        "orgs": [
            {
                "name": "1",
                "options": {
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "test",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "3",
                            "ENABLE_WORKSPACE": "true"
                        }
                    }
                ]
            }
        ]
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.NotNil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2
	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "dev",
			"SERVICE_TYPE":   "STATELESS"},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	// Same result as TestIdentityLabelFilter2, because the fine configuration of the cluster only affects ORG and WORKSPACE labels
	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"service-stateless"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	// Temporary support for the project tag
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// The flag bit ("ENABLETAG": "true") has been set to enable tag scheduling,
// And the job has a label set
func TestIdentityLabelFilter4(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "ENABLETAG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2
	// The org and workspace information of the current job are stored in ENV
	li := &labelconfig.LabelInfo{
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"job"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}
	li2 := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "xx",
			"DICE_WORKSPACE": "TEST",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li2)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"job"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// Test bigdata type job
func TestIdentityLabelFilter6(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "ENABLETAG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2
	// The org and workspace information of the current job are stored in ENV
	li := &labelconfig.LabelInfo{
		Label:          map[string]string{"JOB_KIND": "bigdata"},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Equal(t, []string{"bigdata"}, result.ExclusiveLikes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.False(t, result.Flag)
}

func TestBigdataAndAny(t *testing.T) {
	matchTags := make([]string, 0)
	println(strings.Join(matchTags, ","))
	x1 := strings.Join(matchTags, ",")

	x := strings.Split(x1, ",")

	x = append(x, "bigdata")
	y := strings.Join(x, ",")

	assert.Equal(t, ",bigdata", y)
}
