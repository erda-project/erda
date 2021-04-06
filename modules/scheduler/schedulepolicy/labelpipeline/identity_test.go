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

// 对未开启标签调度的测试用例放在 policy_test.go 中执行

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 并且runtime 或者 job 未带任何 label 情况下
// 集群配置只有基本配置
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

	// 临时支持 project 标签
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	// any 的 flag 被置为 true，但是 any 标签只与 Likes 里的标签搭配组合，而 Likes 是空，所以 any 不会出现在约束条件中
	assert.True(t, result.Flag)
}

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 并且runtime 或者 job 有设置 label 情况下
// 集群配置只有基本配置
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

	// 临时支持 project 标签
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// 测试 addons 有状态服务
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

	// 临时支持 project 标签
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 并且runtime 或者 job 有设置 label 情况下
// 集群配置还有精细配置
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

	// 与 TestIdentityLabelFilter2 的结果相同，因为集群的精细配置只影响 ORG 和 WORKSPACE 标签
	IdentityFilter(&result, &result2, li)
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"service-stateless"}, result.Likes)
	assert.False(t, result.IsPlatform)
	assert.True(t, result.IsUnLocked)

	// 临时支持 project 标签
	assert.Equal(t, []string{"project-"}, result.UnLikePrefixs)
	assert.True(t, result.Flag)
}

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 并且 job 有设置 label 情况下
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
	// 当前 job 的 org, workspace 信息都存放在 ENV 里
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

// 测试 bigdata 类型的 job
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
	// 当前 job 的 org, workspace 信息都存放在 ENV 里
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
