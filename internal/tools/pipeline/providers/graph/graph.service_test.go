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

package graph

import (
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"

	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
)

func Test_loadGraphActionNameAndLogo(t *testing.T) {
	actionStages := [][]*pb.PipelineYmlAction{
		{
			{
				Type: "git-checkout",
			},
		},
	}
	stages := make([]interface{}, 0)
	for _, stage := range actionStages {
		pbStages := make([]interface{}, 0)
		for _, action := range stage {
			pbAction, err := convertAction2Value(action)
			assert.NoError(t, err)
			pbStages = append(pbStages, pbAction.AsInterface())
		}
		stages = append(stages, pbStages)
	}
	stageList, err := structpb.NewList(stages)
	assert.NoError(t, err)
	yml := &pb.PipelineYml{
		Stages:  stageList,
		Version: "1.1",
	}
	bdl := bundle.New()
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "SearchExtensions", func(_ *bundle.Bundle, req apistructs.ExtensionSearchRequest) (map[string]apistructs.ExtensionVersion, error) {
		return map[string]apistructs.ExtensionVersion{
			"git-checkout": apistructs.ExtensionVersion{
				Spec: `name: git-checkout
version: "1.0"
type: action
category: source_code_management
logoUrl: //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/09/28/06e7346b-9377-47d4-8eb7-06a1f735691f.png
displayName: ${{ i18n.displayName }}
desc: ${{ i18n.desc }}
public: true
labels:
  new_workspace: true
  project_level_app: true
supportedVersions: # Deprecated. Please use supportedErdaVersions instead.
  - ">= 3.5"
supportedErdaVersions:
  - ">= 1.0"

params:
  - name: uri
    required: false
    desc: ${{ i18n.formProps.params.uri.labelTip }}
    default: ((gittar.repo))

outputs:
  - name: commit

formProps:
  - component: formGroup
    key: params
    componentProps:
      indentation: true
      showDivider: true
      title: ${{ i18n.formProps.params.componentProps.title }}
    group: params
    
locale:
  zh-CN:
    displayName: 代码克隆
  en-US:
    displayName: Git clone
`,
			},
		}, nil
	})
	defer pm1.Unpatch()

	s := &graphService{
		bdl: bdl,
	}
	err = s.loadGraphActionNameAndLogo(yml)
	assert.NoError(t, err)
	ymlbyte, err := yml.MarshalJSON()
	assert.NoError(t, err)
	t.Log(string(ymlbyte))
}
