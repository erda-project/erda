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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/modules/scheduler/schedulepolicy/labelconfig"
	"github.com/erda-project/erda/pkg/parser/diceyml"
)

// 对以下情形做了测试:

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 集群配置只有基本配置
// 测试 runtime 的 label 中带上了 org 和没有带上 org
func TestOrgLabelFilter1(t *testing.T) {
	// 集群配置中无精细配置，基本配置中未配置 org，则（所有 org）都不开启 org 调度
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
		// lable 中没有 org 标签
		Label:          make(map[string]string),
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	OrgLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, 0, len(result.UnLikePrefixs))
	//assert.Equal(t, labelconfig.ORG_VALUE_PREFIX, result.UnLikePrefixs[0])

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li2 := &labelconfig.LabelInfo{
		// lable 中有 org 标签，但并未开启 org 调度(没有"ENABLE_ORG"的设置)
		Label: map[string]string{
			"DICE_ORG_NAME":  "org-1",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	OrgLabelFilter(&result, &result2, li2)
	// 结果同 result1，因为集群配置中没有精细配置且基本配置中没有开启 org 调度(集群配置中没有设置"ENABLE_ORG")
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, 1, len(result.UnLikePrefixs))
	assert.Equal(t, labelconfig.ORG_VALUE_PREFIX, result.UnLikePrefixs[0])

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li3 := &labelconfig.LabelInfo{
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-12345",
		Selectors: map[string]diceyml.Selectors{
			"placehold": {
				"org":       diceyml.Selector{Values: []string{"org-1"}},
				"workspace": diceyml.Selector{Values: []string{"test"}},
			},
		},
	}

	OrgLabelFilter(&result, &result2, li3)
	// 结果同 result1，因为集群配置中没有精细配置且基本配置中没有开启 org 调度(集群配置中没有设置"ENABLE_ORG")
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, 1, len(result.UnLikePrefixs))
	assert.Equal(t, labelconfig.ORG_VALUE_PREFIX, result.UnLikePrefixs[0])
}

// 在集群的基本配置中设置了开启 org 调度（建议用法是在精细配置中开启，见TestOrgLabelFilter3）
func TestOrgLabelFilter2(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true",
        "ENABLE_ORG": "true"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		// lable 中有 org 标签
		Label: map[string]string{
			"DICE_ORG_NAME":  "1xx",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	OrgLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"org-1xx"}, result.ExclusiveLikes)
}

// 在集群的精细配置中设置了开启 org 调度，并且 runtime 的 label 中设置了 org
func TestOrgLabelFilter3(t *testing.T) {
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
                "name": "2",
                "options": {
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "test",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "2"
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

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "2",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	OrgLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.Equal(t, []string{"org-2"}, result.ExclusiveLikes)
}

// 在集群的精细配置中设置了开启 org 调度，但是 runtime 的 label 中没有设置了 org
func TestOrgLabelFilter4(t *testing.T) {
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
                            "CPU_SUBSCRIBE_RATIO": "2"
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

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li2 := &labelconfig.LabelInfo{
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1111",
	}

	OrgLabelFilter(&result, &result2, li2)
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.ExclusiveLikes))
	//assert.Equal(t, []string{"org-"}, result.UnLikePrefixs)
}

// 在集群的精细配置中设置了开启 org 调度，并且 runtime 的 label 中设置了 org
// 但是 runtime label 中设置的 org 名字没有出现在集群精细配置的 orgs 中
func TestOrgLabelFilter5(t *testing.T) {
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
                            "CPU_SUBSCRIBE_RATIO": "2"
                        }
                    }
                ]
            },
            {
                "name": "test",
                "options": {
                    "ENABLE_ORG": "true"
                }
            }
        ]
    }
}`)
	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "3",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	OrgLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"org-"}, result.UnLikePrefixs)

	li1 := &labelconfig.LabelInfo{
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
		Selectors: map[string]diceyml.Selectors{
			"placehold": {
				"org":       diceyml.Selector{Values: []string{"3"}},
				"workspace": diceyml.Selector{Values: []string{"test"}},
			},
		},
	}
	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	OrgLabelFilter(&result, &result2, li1)
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"org-"}, result.UnLikePrefixs)

}
