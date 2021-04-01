package schedulepolicy

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/scheduler/conf"
	"github.com/erda-project/erda/modules/scheduler/executor/executortypes"
	"github.com/erda-project/erda/pkg/strutil"
)

// 未开启标签调度（即未设置标志位 "ENABLETAG": "true")，
// 无论runtime 或者 job 是否带有 label
func TestDisableLabelFilterChain(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10"
    }
}`)

	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	// runtime 未有任何标签
	r1 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-011",
		},
	}

	// runtime 带有标签
	r2 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-011",
			Labels: map[string]string{
				"DICE_ORG_NAME":    "1",
				"DICE_WORKSPACE":   "test",
				"ENABLE_WORKSPACE": "true",
				"ENABLE_ORG":       "true",
			},
		},
	}

	_, _, refinedConfigs_, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r1)
	assert.Nil(t, err)
	assert.Nil(t, refinedConfigs_)

	_, _, refinedConfigs_, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r2)
	assert.Nil(t, err)
	assert.Nil(t, refinedConfigs_)
}

// 集群已开启标签调度，即集群基本配置中有 "ENABLETAG": "false"
// 配置中基础配置中只配置 org，没有配置 workspace
func TestLabelFilterChain1(t *testing.T) {
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-011",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "test",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, 0, len(si.Likes))
	assert.Equal(t, 0, len(si.UnLikes))
	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"workspace-", "project-"},
		ExclusiveLikes: []string{"org-1"},
		Flag:           true,
	}, si)
	assert.Equal(t, "1", s2.Org)
	assert.True(t, s2.HasOrg)

	// constrains 按照 UnLikePrefixs，UnLikes，LikePrefixs，ExclusiveLikes，Likes 5个 slice 的顺序来组装
	// 其中每个 slice 又是按照 org, workspace, identity 三层来拼装
	// UnLikePrefixs 中会有 workspace- 和 project-
	// Unlikes 中会有 platform, locked
	// LikePrefixs 是空
	// ExclusiveLikes 中会有 org-1
	// Likes 是空
	// 其中 UnLikePrefixs，Unlikes 会转化成 DCOS 的 UNLIKE 语句
	// LikePrefixs, ExclusiveLikes, Likes 会转化成 DCOS 的 LIKE 语句
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-1\\b.*"}, constrains[4])
}

// 集群已开启标签调度，即集群基本配置中有 "ENABLETAG": "false"
// 集群配置中基础配置中只配置 workspace，没有配置 org
func TestLabelFilterChain2(t *testing.T) {
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-012",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "test",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-test"},
		Flag:           true,
	}, si)

	assert.Equal(t, []string{"test"}, s2.WorkSpaces)
	assert.True(t, s2.HasWorkSpace)
	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*"}, constrains[4])
}

// 集群已开启标签调度，即集群基本配置中有 "ENABLETAG": "false"
// 集群配置中精细配置中只配置 org，没有配置 workspace
func TestLabelFilterChain3(t *testing.T) {
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-013",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "test",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"workspace-", "project-"},
		ExclusiveLikes: []string{"org-1"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "1", s2.Org)
	// // 结果与 TestLabelFilterChain1 相同
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-1\\b.*"}, constrains[4])
}

// 集群已开启标签调度，即集群基本配置中有 "ENABLETAG": "false"
// 集群配置中精细配置中只配置 workspace，没有配置 org
func TestLabelFilterChain4(t *testing.T) {
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
                "options": {},
                "workspaces": [
                    {
                        "name": "test",
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-014",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "test",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "test", s2.WorkSpaces[0])
	assert.False(t, s2.HasOrg)
	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果与 TestLabelFilterChain2 相同
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*"}, constrains[4])
}

// 集群配置中精细配置中配置了 org 和 workspace
func TestLabelFilterChain5(t *testing.T) {
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
                "name": "test1",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "3",
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "prod",
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-017",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "test1",
				"DICE_WORKSPACE": "prod",
			},
		},
	}

	s2, si, refinedConfigs_, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-test1", "workspace-prod"},
		Flag:           true,
	}, si)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "prod", s2.WorkSpaces[0])

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-test1\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-prod\\b.*"}, constrains[4])

	assert.NotNil(t, refinedConfigs_)
	refinedConfigs, ok := refinedConfigs_.(map[string]string)
	assert.True(t, ok)

	cpu_subscribe_ratio, ok := refinedConfigs["CPU_SUBSCRIBE_RATIO"]
	assert.True(t, ok)
	// cpu 超卖比读的是该 runtime 隶属的 org（test1）下的 workspace （prod）下的 CPU_SUBSCRIBE_RATIO
	assert.Equal(t, "2", cpu_subscribe_ratio)
}

// 集群配置中精细配置中配置了 org 和 workspace
// label 中的 org 匹配上，但是 workspace 没有匹配上
func TestLabelFilterChain6(t *testing.T) {
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
                "name": "test1",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "4",
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "test",
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-018",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "test1",
				"DICE_WORKSPACE": "prod",
			},
		},
	}

	s2, si, refinedConfigs_, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"workspace-", "project-"},
		ExclusiveLikes: []string{"org-test1"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "test1", s2.Org)
	assert.False(t, s2.HasWorkSpace)

	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-test1\\b.*"}, constrains[4])

	// 精细配置中 org 匹配上了，workspace没有匹配上，cpu 超卖比被设置成该 org 下的超卖比
	refinedConfigs, ok := refinedConfigs_.(map[string]string)
	assert.True(t, ok)
	cpu_subscribe_ratio, ok := refinedConfigs["CPU_SUBSCRIBE_RATIO"]
	assert.True(t, ok)
	assert.Equal(t, "4", cpu_subscribe_ratio)
}

// 集群配置中精细配置中配置了 org 和 workspace
// label 中的 org 没匹配上（等同于 org 和 workspace 都没有匹配上）
func TestLabelFilterChain7(t *testing.T) {
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
                "name": "test1",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "2",
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "test",
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-018",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "test2",
				"DICE_WORKSPACE": "prod",
			},
		},
	}

	s2, si, refinedConfigs_, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:    true,
		Location:      map[string]interface{}{},
		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
	}, si)

	assert.False(t, s2.HasOrg)
	assert.False(t, s2.HasWorkSpace)
	assert.False(t, s2.HasProject)

	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])

	// 精细配置中 org 没有匹配上，所以 cpu 超卖比是默认集群配置超卖比
	_, ok := refinedConfigs_.(map[string]string)
	assert.False(t, ok)
}

// 集群配置中精细配置中配置了 org 和 workspace
// label 中的 org 没匹配上（等同于 org 和 workspace 都没有匹配上）
// runtime 带上 "SERVICE_TYPE": "STATELESS"
func TestLabelFilterChain8(t *testing.T) {
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
                "name": "test1",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "2",
                    "ENABLE_ORG": "true"
                },
                "workspaces": [
                    {
                        "name": "test",
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-018",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "test2",
				"DICE_WORKSPACE": "prod",
				"SERVICE_TYPE":   "STATELESS",
				"SPECIFIC_HOSTS": "192.168.1.2",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateless"},

		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
		SpecificHost:  []string{"192.168.1.2"},
	}, si)

	assert.Equal(t, 1, len(s2.SpecificHost))
	assert.Equal(t, "192.168.1.2", s2.SpecificHost[0])
	assert.True(t, s2.Stateless)
	assert.True(t, s2.PreferStateless)

	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateless\\b.*"}, constrains[5])
	// assert.Equal(t, []string{"hostname", "LIKE", "192.168.1.2"}, constrains[6])
}

// 测试兼容"WORKSPACETAGS"标签的情况，注意"ENABLE_WORKSPACE"不开启
func TestLabelFilterChain9(t *testing.T) {
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "test-011",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "test",
				"SERVICE_TYPE":   "STATELESS",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateless"},

		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
	}, si)

	assert.True(t, s2.Stateless)
	assert.True(t, s2.PreferStateless)
	assert.False(t, s2.HasWorkSpace)
	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateless\\b.*"}, constrains[5])

	r2 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-012",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "staging",
				"SERVICE_TYPE":   "STATELESS",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r2)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateless"},

		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-staging"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "staging", s2.WorkSpaces[0])
	assert.True(t, s2.Stateless)
	assert.True(t, s2.PreferStateless)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-staging\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateless\\b.*"}, constrains[5])

	r2a := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-012",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "staging",
				"SERVICE_TYPE":   "ADDONS",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r2a)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateful"},

		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-staging"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "staging", s2.WorkSpaces[0])
	assert.True(t, s2.Stateful)
	assert.True(t, s2.PreferStateful)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-staging\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateful\\b.*"}, constrains[5])

	r3 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-013",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "prod",
				"SERVICE_TYPE":   "STATELESS",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r3)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateless"},

		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
	}, si)

	assert.False(t, s2.HasWorkSpace)
	assert.True(t, s2.Stateless)
	assert.True(t, s2.PreferStateless)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateless\\b.*"}, constrains[5])

	r3a := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-013",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "1",
				"DICE_WORKSPACE": "prod",
				"SERVICE_TYPE":   "ADDONS",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r3a)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"service-stateful"},

		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
	}, si)

	assert.True(t, s2.Stateful)
	assert.False(t, s2.HasWorkSpace)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// // 结果及其顺序的分析参考 TestLabelFilterChain1
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bservice-stateful\\b.*"}, constrains[5])
}

// 多个企业情况下使用集群的配置（非精细化的配置）
func TestMultiOrg(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "MARATHON",
    "name": "MARATHONFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "CPU_SUBSCRIBE_RATIO": "10",
        "ENABLETAG": "true",
        "ENABLE_ORG": "true",
        "ENABLE_WORKSPACE": "true"
    },
    "optionsPlus": {
        "orgs": [
            {
                "name": "test1",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "3"
                },
                "workspaces": [
                    {
                        "name": "prod",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "2"
                        }
                    }
                ]
            },
            {
                "name": "hangzhou",
                "options": {
                    "CPU_SUBSCRIBE_RATIO": "5"
                },
                "workspaces": [
                    {
                        "name": "prod",
                        "options": {
                            "CPU_SUBSCRIBE_RATIO": "4"
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	r := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-017",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "hangzhou",
				"DICE_WORKSPACE": "prod",
			},
		},
	}

	s2, si, refinedConfigs_, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, r)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-hangzhou", "workspace-prod"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "hangzhou", s2.Org)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "prod", s2.WorkSpaces[0])
	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-hangzhou\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-prod\\b.*"}, constrains[4])

	assert.NotNil(t, refinedConfigs_)
	refinedConfigs, ok := refinedConfigs_.(map[string]string)
	assert.True(t, ok)

	cpu_subscribe_ratio, ok := refinedConfigs["CPU_SUBSCRIBE_RATIO"]
	assert.True(t, ok)
	// cpu 超卖比读的是该 runtime 隶属的 org（hangzhou）下的 workspace （prod）下的 CPU_SUBSCRIBE_RATIO
	assert.Equal(t, "4", cpu_subscribe_ratio)

	r2 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "prod-017",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "test1",
				"DICE_WORKSPACE": "dev",
			},
		},
	}

	s2, si, refinedConfigs_, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r2)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-test1", "workspace-dev"},
		Flag:           true,
	}, si)
	assert.True(t, s2.HasOrg)
	assert.Equal(t, "test1", s2.Org)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "dev", s2.WorkSpaces[0])

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-test1\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-dev\\b.*"}, constrains[4])

	assert.NotNil(t, refinedConfigs_)
	refinedConfigs, ok = refinedConfigs_.(map[string]string)
	assert.True(t, ok)

	cpu_subscribe_ratio, ok = refinedConfigs["CPU_SUBSCRIBE_RATIO"]
	assert.True(t, ok)
	// cpu 超卖比读的是该 runtime 隶属的 org（test1）下的配置，workspace没有匹配上精细配置
	assert.Equal(t, "3", cpu_subscribe_ratio)

	r3 := apistructs.ServiceGroup{
		ClusterName: "terminus-dev",
		Dice: apistructs.Dice{
			ID: "staging-0179",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "wrong",
				"DICE_WORKSPACE": "staging",
			},
		},
	}

	s2, si, refinedConfigs_, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, r3)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-wrong", "workspace-staging"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "wrong", s2.Org)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "staging", s2.WorkSpaces[0])

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-wrong\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-staging\\b.*"}, constrains[4])

	// 精细化配置为空
	assert.Nil(t, refinedConfigs_)
}

// 普通 job 和 大数据 job
func TestLabelFilterChainForJob(t *testing.T) {
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

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	j := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name:   "xx",
			Labels: map[string]string{},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, j)
	assert.Nil(t, err)

	assert.NotEqual(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs: []string{"org-", "workspace-", "project-"},
		Flag:          true,
	}, si)

	assert.True(t, s2.Job)
	assert.False(t, s2.HasOrg)
	assert.False(t, s2.HasWorkSpace)
	assert.False(t, s2.HasProject)

	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[5])

	// 大数据类型的 job
	j2 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"JOB_KIND": "bigdata",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j2)
	assert.Nil(t, err)

	assert.NotEqual(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"org-", "workspace-", "project-"},
		ExclusiveLikes: []string{"bigdata"},
	}, si)

	assert.True(t, s2.BigData)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bworkspace-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bbigdata\\b.*"}, constrains[5])
}

// metronome 开启了 workspace 调度并且job 带了相应的 label
func TestLabelFilterChainForJob2(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true"
    }
}`)

	// 测试环境的 job
	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	j := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "TEST",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, j)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-test"},
		Flag:           true,
	}, si)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "test", s2.WorkSpaces[0])
	assert.True(t, s2.Job)

	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[5])

	// 预发环境的job
	j2 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "STAGING",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j2)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"org-", "project-"},
		InclusiveLikes: []string{"workspace-dev", "workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "dev"))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "test"))

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-dev\\b.*|.*\\bworkspace-test\\b.*"}, constrains[5])

	// 生产环境的job
	j3 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "PROD",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j3)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"org-", "project-"},
		InclusiveLikes: []string{"workspace-dev", "workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.Job)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "dev"))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "test"))

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-dev\\b.*|.*\\bworkspace-test\\b.*"}, constrains[5])
}

// metronome 开启了 workspace 调度并且job 带了相应的 label
func TestLabelFilterChainForJob3(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "ENABLETAG": "true",
        "ENABLE_ORG": "true",
        "ENABLE_WORKSPACE": "true"
    }
}`)

	// 测试环境的 job
	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	j := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "TEST",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, j)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-xx", "workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "xx", s2.Org)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "test", s2.WorkSpaces[0])
	assert.True(t, s2.Job)
	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-xx\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[5])

	// 预发环境的job
	j2 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xy",
				"DICE_WORKSPACE": "STAGING",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j2)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-xy"},
		InclusiveLikes: []string{"workspace-dev", "workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.HasOrg)
	assert.Equal(t, "xy", s2.Org)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "dev"))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "test"))
	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-xy\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-dev\\b.*|.*\\bworkspace-test\\b.*"}, constrains[5])

	// 生产环境的job
	j3 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xyz",
				"DICE_WORKSPACE": "PROD",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j3)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"project-"},
		ExclusiveLikes: []string{"org-xyz"},
		InclusiveLikes: []string{"workspace-dev", "workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.Job)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "test"))
	assert.False(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "dev"))
	assert.True(t, s2.HasOrg)
	assert.Equal(t, "xyz", s2.Org)

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\borg-xyz\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-dev\\b.*|.*\\bworkspace-test\\b.*"}, constrains[5])
}

// 测试配置了 STAGING_JOB_DEST 和 PROD_JOB_DEST 的情况
func TestLabelFilterChainForJob4(t *testing.T) {
	var jsonBlob = []byte(`{
    "clusterName": "terminus-dev",
    "kind": "METRONOME",
    "name": "METRONOMEFORTERMINUSDEV",
    "options": {
        "ADDR": "http://master.mesos/service/marathon",
        "ENABLETAG": "true",
        "ENABLE_WORKSPACE": "true",
		"PROD_JOB_DEST": "staging",
		"STAGING_JOB_DEST": "test,prod"
    }
}`)

	// 测试环境的 job
	var config conf.ExecutorConfig
	err := json.Unmarshal(jsonBlob, &config)
	assert.Nil(t, err)
	assert.Nil(t, config.OptionsPlus)

	labelConfigs := &executortypes.ExecutorWholeConfigs{
		BasicConfig: config.Options,
		PlusConfigs: config.OptionsPlus,
	}

	j := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "TEST",
			},
		},
	}

	s2, si, _, err := LabelFilterChain(labelConfigs, config.Name, config.Kind, j)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"org-", "project-"},
		ExclusiveLikes: []string{"workspace-test"},
		Flag:           true,
	}, si)

	assert.True(t, s2.Job)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 1, len(s2.WorkSpaces))
	assert.Equal(t, "test", s2.WorkSpaces[0])
	// constrains, ok := constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[5])

	// 预发环境的job
	j2 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "STAGING",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j2)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked: true,
		Location:   map[string]interface{}{},
		Likes:      []string{"job"},

		UnLikePrefixs:  []string{"org-", "project-"},
		InclusiveLikes: []string{"workspace-test", "workspace-prod"},
		Flag:           true,
	}, si)

	assert.True(t, s2.Job)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, 2, len(s2.WorkSpaces))
	assert.True(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "prod"))
	assert.True(t, strutil.Contains(strutil.Concat(s2.WorkSpaces...), "test"))

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-test\\b.*|.*\\bworkspace-prod\\b.*"}, constrains[5])

	// 生产环境的job
	j3 := apistructs.Job{
		JobFromUser: apistructs.JobFromUser{
			Name: "xx",
			Labels: map[string]string{
				"DICE_ORG_NAME":  "xx",
				"DICE_WORKSPACE": "PROD",
			},
		},
	}

	s2, si, _, err = LabelFilterChain(labelConfigs, config.Name, config.Kind, j3)
	assert.Nil(t, err)

	assert.Equal(t, apistructs.ScheduleInfo{
		IsUnLocked:     true,
		Likes:          []string{"job"},
		Location:       map[string]interface{}{},
		UnLikePrefixs:  []string{"org-", "project-"},
		InclusiveLikes: []string{"workspace-staging"},
		Flag:           true,
	}, si)

	assert.True(t, s2.Job)
	assert.True(t, s2.HasWorkSpace)
	assert.Equal(t, "staging", s2.WorkSpaces[0])

	// constrains, ok = constrains_.([][]string)
	// assert.True(t, ok)

	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\borg-[^,]+\\b.*"}, constrains[0])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bproject-[^,]+\\b.*"}, constrains[1])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\bplatform\\b.*"}, constrains[2])
	// assert.Equal(t, []string{"dice_tags", "UNLIKE", ".*\\blocked\\b.*"}, constrains[3])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bany\\b.*|.*\\bjob\\b.*"}, constrains[4])
	// assert.Equal(t, []string{"dice_tags", "LIKE", ".*\\bworkspace-staging\\b.*"}, constrains[5])
}
