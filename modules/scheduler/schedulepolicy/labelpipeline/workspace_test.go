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

// The flag bit ("ENABLETAG": "true") has been set to enable tag scheduling,
// The cluster configuration has only basic configuration
// The label of the test runtime has workspace and no workspace
func TestWorkspaceLabelFilter1(t *testing.T) {
	// There is no fine configuration in the cluster configuration. If org is not configured in the basic configuration, then (all orgs) do not enable org scheduling
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
		// There is no workspace label in the label
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
		// There is a workspace label in the label, but worksapce scheduling is not enabled (there is no "ENABLE_WORKSPACE" setting in the cluster configuration)
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
	// The result is the same as result1, because there is no fine configuration in the cluster configuration and the org scheduling is not enabled in the basic configuration ("ENABLE_ORG" is not set)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)

	result = labelconfig.RawLabelRuleResult{}
	result2 = labelconfig.RawLabelRuleResult2{}
	li3 := &labelconfig.LabelInfo{
		// There is a workspace label in the label, but worksapce scheduling is not enabled (there is no "ENABLE_WORKSPACE" setting in the cluster configuration)
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
	// The result is the same as result1, because there is no fine configuration in the cluster configuration and the org scheduling is not enabled in the basic configuration ("ENABLE_ORG" is not set)
	assert.Zero(t, len(result.Likes))
	assert.Zero(t, len(result.UnLikes))
	assert.Zero(t, len(result.LikePrefixs))
	assert.Zero(t, len(result.ExclusiveLikes))
	assert.False(t, result.Flag)
	assert.Equal(t, []string{"workspace-"}, result.UnLikePrefixs)
}

// Open workspace scheduling is set in the basic configuration of the cluster (the recommended usage is to enable it in the fine configuration, and specify the subordinate org and workspace)
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
		// org label in label
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
		// org label in label
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

// Open workspace scheduling is set in the fine configuration of the cluster, and org and workspace are set in the runtime label
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

// Open workspace scheduling is set in the fine configuration of the cluster, and org and workspace are set in the runtime label
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

// Open workspace scheduling is set in the fine configuration of the cluster, and workspace is set in the runtime label
// But the workspace name set in the runtime label does not appear in the orgs of the cluster's fine configuration
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

// Test the compatibility with the "WORKSPACETAGS" tag, and note that "ENABLE_WORKSPACE" is not turned on
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

	// staging environment services
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

	// prod environment services
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
		// org label in label
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
		// org label in label
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
		// org label in label
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
		// org label in label
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

// Test settings STAGING_JOB_DEST and PROD_JOB_DEST
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
