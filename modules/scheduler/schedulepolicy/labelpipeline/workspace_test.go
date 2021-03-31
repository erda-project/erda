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

// 已设置了标志位 ("ENABLETAG": "true") 开启标签调度，
// 集群配置只有基本配置
// 测试 runtime 的 label 中带上了 workspace 和没有带上 workspace
func TestWorkspaceLabelFilter1(t *testing.T) {
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
		// lable 中没有 workspace 标签
		Label:          make(map[string]string),
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}
	li2 := &labelconfig.LabelInfo{
		// lable 中有 workspace 标签，但并未开启 worksapce 调度(集群配置中没有"ENABLE_WORKSPACE"的设置)
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	WorkspaceLabelFilter(&result, &result2, li2)
	// 结果同 result1，因为集群配置中没有精细配置且基本配置中没有开启 org 调度(没有设置"ENABLE_ORG")
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}
	li3 := &labelconfig.LabelInfo{
		// lable 中有 workspace 标签，但并未开启 worksapce 调度(集群配置中没有"ENABLE_WORKSPACE"的设置)
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
		Selectors: map[string]diceyml.Selectors{
			"placehold": {
				"org":       diceyml.Selector{Values: []string{"1"}},
				"workspace": diceyml.Selector{Values: []string{"test"}},
			},
		},
	}

	WorkspaceLabelFilter(&result, &result2, li3)
	// 结果同 result1，因为集群配置中没有精细配置且基本配置中没有开启 org 调度(没有设置"ENABLE_ORG")
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)
}

// 在集群的基本配置中设置了开启 workspace 调度（建议用法是在精细配置中开启，指定隶属的 org 及 workspace）
func TestWorkspaceLabelFilter2(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true"
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
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-test"}, result.ExclusiveLikes)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li1 := &labelconfig.LabelInfo{
		// lable 中有 org 标签
		Label: map[string]string{
			"DICE_ORG_NAME":  "1111",
			"DICE_WORKSPACE": "testttt",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
		Selectors: map[string]diceyml.Selectors{
			"placehold": {
				"org":       diceyml.Selector{Values: []string{"1"}},
				"workspace": diceyml.Selector{Values: []string{"test"}},
			},
		},
	}

	WorkspaceLabelFilter(&result, &result2, li1)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-test"}, result.ExclusiveLikes)

}

// 在集群的精细配置中设置了开启 workspace 调度，并且 runtime 的 label 中设置了 org 及 workspace
func TestWorkspaceLabelFilter3(t *testing.T) {
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
                        "name": "staging",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "2",
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
			"DICE_ORG_NAME":  "2",
			"DICE_WORKSPACE": "staging",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-12345",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs), "%+v", result.UnLikePrefixs)
	assert.Equal(t, []string{"workspace-staging"}, result.ExclusiveLikes)
}

// 在集群的精细配置中设置了开启 workspace 调度，但是 runtime 的 label 中没有设置 workspace
func TestWorkspaceLabelFilter4(t *testing.T) {
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
	assert.NotNil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		Label:          map[string]string{},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1111",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)
}

// 在集群的精细配置中设置了开启 workspace 调度，并且 runtime 的 label 中设置了 workspace
// 但是 runtime label 中设置的 workspace 名字没有出现在集群精细配置的 orgs 中
func TestWorkspaceLabelFilter5(t *testing.T) {
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
                            "CPU_SUBSCRIBE_RATIO": "5"
                        }
                    },
                    {
                        "name": "staging",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "3"
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
	assert.NotNil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "prod",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1234",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)
}

// 测试兼容"WORKSPACETAGS"标签的情况，注意"ENABLE_WORKSPACE"不开启
func TestWorkspaceLabelFilter6(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true",
		"WORKSPACETAGS": "staging"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	var result labelconfig.RawLabelRuleResult
	var result2 labelconfig.RawLabelRuleResult2

	// 预发环境的 服务
	li := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "staging",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1227",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-staging"}, result.ExclusiveLikes)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	// 生产环境的服务
	li2 := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "prod",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1227",
	}

	WorkspaceLabelFilter(&result, &result2, li2)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)
}

func TestWorkspaceLabelFilterForJob1(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/metronome",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true"
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
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-test"}, result.ExclusiveLikes)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li2 := &labelconfig.LabelInfo{
		// lable 中有 org 标签
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "test",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	WorkspaceLabelFilter(&result, &result2, li2)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-test"}, result.ExclusiveLikes)
}

func TestWorkspaceLabelFilterForJob2(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/metronome",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true"
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
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "staging",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-dev", "workspace-test"}, result.InclusiveLikes)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li2 := &labelconfig.LabelInfo{
		// lable 中有 org 标签
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "prod",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1222",
	}
	WorkspaceLabelFilter(&result, &result2, li2)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-dev", "workspace-test"}, result.InclusiveLikes)
}

// 测试设置 STAGING_JOB_DEST 和 PROD_JOB_DEST
func TestWorkspaceLabelFilterForJob3(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/metronome",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true",
		"PROD_JOB_DEST": "staging",
		"STAGING_JOB_DEST": "test,prod"
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
			"DICE_WORKSPACE": "staging",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1221",
	}

	WorkspaceLabelFilter(&result, &result2, li)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-test", "workspace-prod"}, result.InclusiveLikes)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}

	li2 := &labelconfig.LabelInfo{
		Label: map[string]string{
			"DICE_ORG_NAME":  "1",
			"DICE_WORKSPACE": "prod",
		},
		ExecutorName:   config.Name,
		ExecutorKind:   config.Kind,
		ExecutorConfig: &executortypes.ExecutorWholeConfigs{BasicConfig: config.Options, PlusConfigs: config.OptionsPlus},
		OptionsPlus:    config.OptionsPlus,
		ObjName:        "test-1222",
	}
	WorkspaceLabelFilter(&result, &result2, li2)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.UnLikePrefixs))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-staging"}, result.InclusiveLikes)
}
