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

package pipelinesvc

import (
	"fmt"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func Test_passedDataWhenCreate_putPassedDataByPipelineYml(t *testing.T) {
	type fields struct {
		extMarketSvc     *extmarketsvc.ExtMarketSvc
		actionJobDefines map[string]*diceyml.Job
		actionJobSpecs   map[string]*apistructs.ActionSpec
	}
	type args struct {
		pipelineYml string
	}
	tests := []struct {
		name                 string
		fields               fields
		args                 args
		wantErr              bool
		wantActionJobDefines []string
		wantActionJobSpecs   []string
	}{
		{
			name: "normal",
			fields: fields{
				actionJobSpecs:   map[string]*apistructs.ActionSpec{},
				actionJobDefines: map[string]*diceyml.Job{},
			},
			args: args{
				pipelineYml: "version: \"1.1\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          description: 代码仓库克隆\n  - stage:\n      - java:\n          alias: java-demo\n          description: 针对 java 工程的编译打包任务，产出可运行镜像\n          version: \"1.0\"\n          params:\n            build_type: maven\n            container_type: spring-boot\n            jdk_version: \"11\"\n            target: ./target/docker-java-app-example.jar\n            workdir: ${git-checkout}\n  - stage:\n      - release:\n          alias: release\n          description: 用于打包完成时，向dicehub 提交完整可部署的dice.yml。用户若没在pipeline.yml里定义该action，CI会自动在pipeline.yml里插入该action\n          params:\n            dice_yml: ${git-checkout}/dice.yml\n            image:\n              java-demo: ${java-demo:OUTPUT:image}\n  - stage:\n      - dice:\n          alias: dice\n          description: 用于 dice 平台部署应用服务\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n",
			},
			wantActionJobDefines: []string{"git-checkout@1.0", "java@1.0", "release@1.0", "dice@1.0"},
			wantActionJobSpecs:   []string{"git-checkout@1.0", "java@1.0", "release@1.0", "dice@1.0"},
		},
		{
			name: "want_error",
			fields: fields{
				actionJobSpecs:   map[string]*apistructs.ActionSpec{},
				actionJobDefines: map[string]*diceyml.Job{},
			},
			args: args{
				pipelineYml: "version: \"1.1\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: git-checkout\n          description: 代码仓库克隆\n  - stage:\n      - java:\n          alias: java-demo\n          description: 针对 java 工程的编译打包任务，产出可运行镜像\n          version: \"1.0\"\n          params:\n            build_type: maven\n            container_type: spring-boot\n            jdk_version: \"11\"\n            target: ./target/docker-java-app-example.jar\n            workdir: ${git-checkout}\n  - stage:\n      - release:\n          alias: release\n          description: 用于打包完成时，向dicehub 提交完整可部署的dice.yml。用户若没在pipeline.yml里定义该action，CI会自动在pipeline.yml里插入该action\n          params:\n            dice_yml: ${git-checkout}/dice.yml\n            image:\n              java-demo: ${java-demo:OUTPUT:image}\n  - stage:\n      - dice:\n          alias: dice\n          description: 用于 dice 平台部署应用服务\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n",
			},
			wantActionJobDefines: []string{},
			wantActionJobSpecs:   []string{},
			wantErr:              true,
		},
		{
			name: "api-test_1.0",
			fields: fields{
				actionJobSpecs:   map[string]*apistructs.ActionSpec{},
				actionJobDefines: map[string]*diceyml.Job{},
			},
			args: args{
				pipelineYml: "version: '1.1'\nstages:\n  - - alias: api-test\n      type: api-test\n      description: 执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\n      version: '1.0'\n      params:\n        body:\n          type: none\n        method: GET\n        url: /api/user\n      resources: {}\n      displayName: 接口测试\n      logoUrl: >-\n        //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\n  - - alias: api-test1\n      type: api-test\n      description: 执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\n      version: '1.0'\n      params:\n        body:\n          type: none\n        method: GET\n        url: /api/user\n      resources: {}\n      displayName: 接口测试\n      logoUrl: >-\n        //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\nflatActions: null\nlifecycle: null\n",
			},
			wantActionJobDefines: []string{"api-test@1.0"},
			wantActionJobSpecs:   []string{"api-test@1.0"},
		},
		{
			name: "api-test_1.0_with_2.0",
			fields: fields{
				actionJobSpecs:   map[string]*apistructs.ActionSpec{},
				actionJobDefines: map[string]*diceyml.Job{},
			},
			args: args{
				pipelineYml: "version: '1.1'\nstages:\n  - - alias: api-test\n      type: api-test\n      description: 执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\n      version: '2.0'\n      params:\n        body:\n          type: none\n        method: GET\n        url: /api/user\n      resources: {}\n      displayName: 接口测试\n      logoUrl: >-\n        //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\n  - - alias: api-test1\n      type: api-test\n      description: 执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\n      version: '1.0'\n      params:\n        body:\n          type: none\n        method: GET\n        url: /api/user\n      resources: {}\n      displayName: 接口测试\n      logoUrl: >-\n        //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\nflatActions: null\nlifecycle: null\n",
			},
			wantActionJobDefines: []string{"api-test@1.0", "api-test@2.0"},
			wantActionJobSpecs:   []string{"api-test@1.0", "api-test@2.0"},
		},
		{
			name: "snippet",
			fields: fields{
				actionJobSpecs:   map[string]*apistructs.ActionSpec{},
				actionJobDefines: map[string]*diceyml.Job{},
			},
			args: args{
				pipelineYml: "version: '1.1'\nstages:\n  - - alias: api-test\n      type: api-test\n      description: 执行单个接口测试。上层可以通过 pipeline.yml 编排一组接口测试的执行顺序。\n      version: '1.0'\n      params:\n        body:\n          type: none\n        method: GET\n        url: /api/user\n      resources: {}\n      displayName: 接口测试\n      logoUrl: >-\n        //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/10/24195384-07b7-4203-93e1-666373639af4.png\n  - - alias: snippet\n      type: snippet\n      description: 嵌套流水线可以声明嵌套的其他 pipeline.yml\n      resources: {}\n      displayName: 嵌套流水线\n      logoUrl: >-\n        http://terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/10/22/410935c6-e399-463a-b87b-0b774240d12e.png\nflatActions: null\nlifecycle: null\n",
			},
			wantActionJobDefines: []string{"api-test@1.0"},
			wantActionJobSpecs:   []string{"api-test@1.0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			that := &passedDataWhenCreate{
				actionJobDefines: tt.fields.actionJobDefines,
				actionJobSpecs:   tt.fields.actionJobSpecs,
			}
			yml, err := pipelineyml.New([]byte(tt.args.pipelineYml))
			assert.NoError(t, err)

			var svc *extmarketsvc.ExtMarketSvc
			var guard *monkey.PatchGuard
			guard = monkey.PatchInstanceMethod(reflect.TypeOf(svc), "SearchActions", func(svc *extmarketsvc.ExtMarketSvc, items []string, ops ...extmarketsvc.OpOption) (map[string]*diceyml.Job, map[string]*apistructs.ActionSpec, error) {
				guard.Unpatch()
				defer guard.Restore()
				actionJobMap := make(map[string]*diceyml.Job)
				actionSpecMap := make(map[string]*apistructs.ActionSpec)
				for _, item := range items {
					actionJobMap[item] = &diceyml.Job{}
					actionSpecMap[item] = &apistructs.ActionSpec{}
				}
				if tt.wantErr {
					return nil, nil, fmt.Errorf("want error")
				}
				return actionJobMap, actionSpecMap, nil
			})
			that.extMarketSvc = svc

			if err := that.putPassedDataByPipelineYml(yml); (err != nil) != tt.wantErr {
				t.Errorf("putPassedDataByPipelineYml() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, len(that.actionJobDefines), len(tt.wantActionJobDefines), "wantActionJobDefines")
			assert.Equal(t, len(that.actionJobSpecs), len(tt.wantActionJobSpecs), "wantActionJobSpecs")

			for _, v := range tt.wantActionJobSpecs {
				assert.NotEmpty(t, that.actionJobSpecs[v])
			}
			for _, v := range tt.wantActionJobDefines {
				assert.NotEmpty(t, that.actionJobDefines[v])
			}
		})
	}
}
