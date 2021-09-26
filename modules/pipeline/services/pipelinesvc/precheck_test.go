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

package pipelinesvc

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/precheck/prechecktype"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func Test_setItemForCheckRealDiceYml(t *testing.T) {
	type args struct {
		itemForCheck *prechecktype.ItemsForCheck
		p            *spec.Pipeline
		userID       string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				p: &spec.Pipeline{
					PipelineExtra: spec.PipelineExtra{
						PipelineYml: "version: \"1.1\"\nstages:\n  - stage:\n      - git-checkout:\n          alias: repo\n          description: 代码仓库克隆\n          params:\n            depth: 1\n  - stage:\n      - buildpack:\n          alias: bp-web\n          description: |-\n            平台内置的应用构建逻辑。\n            目前支持：\n            1. Java\n            2. NodeJS(Herd)\n            3. Single Page Application(SPA)\n            4. Dockerfile\n          params:\n            bp_args:\n              USE_AGENT: \"true\"\n            bp_repo: http://git.terminus.io/buildpacks/dice-bpack-termjava.git\n            bp_ver: master\n            context: ${repo}\n            modules:\n              - name: web\n  - stage:\n      - release:\n          alias: release\n          description: 用于打包完成时，向dicehub 提交完整可部署的dice.yml。用户若没在pipeline.yml里定义该action，CI会自动在pipeline.yml里插入该action\n          params:\n            check_diceyml: \"false\"\n            cross_cluster: \"true\"\n            dice_development_yml: ${repo}/dice_development.yml\n            dice_production_yml: ${repo}/dice_production.yml\n            dice_staging_yml: ${repo}/dice_staging.yml\n            dice_test_yml: ${repo}/dice_test.yml\n            dice_yml: ${repo}/dice.yml\n            replacement_images:\n              - ${bp-web}/pack-result\n  - stage:\n      - dice:\n          alias: dice\n          description: 用于 dice 平台部署应用服务\n          params:\n            release_id: ${release:OUTPUT:releaseID}\n",
						Extra: spec.PipelineExtraInfo{
							DiceWorkspace: "",
						},
						CommitDetail: apistructs.CommitDetail{
							RepoAbbr: "xxxx",
						},
					},
				},
				itemForCheck: &prechecktype.ItemsForCheck{
					Files: map[string]string{
						"dice_yml": "version: 2\nenvs:\n  TERMINUS_APP_NAME: \"TEST-global\"\n  TEST_PARAM: \"param_value\"\nservices:\n  web:\n    ports:\n      - 8080\n      - 20880\n    health_check:\n      exec:\n        cmd: \"echo 1\"\n    deployments:\n      replicas: 1\n    resources:\n      cpu: 0.1\n      mem: 512\n      disk: 0\n    expose:\n      - 20880",
					},
				},
				userID: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setItemForCheckRealDiceYml(tt.args.p, tt.args.itemForCheck, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("setItemForCheckRealDiceYml() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
