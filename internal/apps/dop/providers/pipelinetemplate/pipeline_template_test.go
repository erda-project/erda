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

package pipelinetemplate

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/dop/pipelinetemplate/pb"
	dbclient "github.com/erda-project/erda/internal/apps/dop/providers/pipelinetemplate/db"
)

func TestQueryPipelineTemplates(t *testing.T) {
	db := &dbclient.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "QueryByPipelineTemplates", func(_ *dbclient.DBClient, template *dbclient.DicePipelineTemplate, pageSize int, pageNo int) ([]dbclient.DicePipelineTemplate, int, error) {
		return []dbclient.DicePipelineTemplate{
			{
				Name:      "custom",
				ScopeType: "dice",
				ScopeId:   "0",
			},
		}, 1, nil
	})
	defer pm1.Unpatch()

	p := &ServiceImpl{
		db: db,
	}
	res, err := p.QueryPipelineTemplates(context.Background(), &pb.PipelineTemplateQueryRequest{
		ScopeType: "dice",
		ScopeID:   "0",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Data.Data))
}

func TestGetPipelineTemplateVersion(t *testing.T) {
	db := &dbclient.DBClient{}
	pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTemplate", func(_ *dbclient.DBClient, name string, scopeType string, scopeId string) (*dbclient.DicePipelineTemplate, error) {
		return &dbclient.DicePipelineTemplate{
			Name:      "custom",
			ScopeType: "dice",
			ScopeId:   "0",
		}, nil
	})
	defer pm1.Unpatch()
	pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTemplateVersion", func(_ *dbclient.DBClient, version string, templateId uint64) (*dbclient.DicePipelineTemplateVersion, error) {
		return &dbclient.DicePipelineTemplateVersion{
			Name:    "custom",
			Version: "1.0",
		}, nil
	})
	defer pm2.Unpatch()

	p := &ServiceImpl{
		db: db,
	}
	res, err := p.GetPipelineTemplateVersion(context.Background(), &pb.PipelineTemplateVersionGetRequest{
		Name:      "custom",
		ScopeType: "dice",
		ScopeID:   "0",
		Version:   "1.0",
	})
	assert.NoError(t, err)
	assert.Equal(t, "custom", res.Data.Name)
}

func TestRenderPipelineTemplate(t *testing.T) {
	type arg struct {
		spec string
	}
	tests := []struct {
		name string
		arg  arg
		want string
	}{
		{
			name: "custom",
			arg: arg{
				spec: `name: custom
version: "1.0"
desc: 自定义模板

template: |
  version: 1.1
  stages:`,
			},
			want: "custom",
		},
		{
			name: "java-boot-maven-dice",
			arg: arg{
				spec: `name: java-boot-maven-dice
version: "1.0"
desc: springboot maven 打包构建部署到 dice 的模板

template: |

  version: 1.1
  stages:
    - stage:
        - git-checkout:
            params:
              depth: 1

    - stage:
        - java-build:
            version: "1.0"
            params:
              build_cmd:
                - mvn package
              jdk_version: 8
              workdir: ${git-checkout}

    - stage:
        - release:
            params:
              dice_yml: ${git-checkout}/dice.yml
              services:
                dice.yml中的服务名:
                  image: registry.cn-hangzhou.aliyuncs.com/terminus/terminus-openjdk:v11.0.6
                  copys:
                    - ${java-build:OUTPUT:buildPath}/target/jar包的名称:/target/jar包的名称
                  cmd: java -jar /target/jar包的名称

    - stage:
        - dice:
            params:
              release_id: ${release:OUTPUT:releaseID}


params:

  - name: pipeline_version
    desc: 生成的pipeline的版本
    default: "1.1"
    required: false

  - name: pipeline_cron
    desc: 定时任务的cron表达式
    required: false

  - name: pipeline_scheduling
    desc: 流水线调度策略
    required: false
`,
			},
			want: "java-boot-maven-dice",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &dbclient.DBClient{}
			pm1 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTemplateVersion", func(_ *dbclient.DBClient, version string, templateId uint64) (*dbclient.DicePipelineTemplateVersion, error) {
				return &dbclient.DicePipelineTemplateVersion{
					Name:    tt.name,
					Version: "1.0",
					Spec:    tt.arg.spec,
				}, nil
			})
			defer pm1.Unpatch()
			pm2 := monkey.PatchInstanceMethod(reflect.TypeOf(db), "GetPipelineTemplate", func(_ *dbclient.DBClient, name string, scopeType string, scopeId string) (*dbclient.DicePipelineTemplate, error) {
				return &dbclient.DicePipelineTemplate{
					Name:      tt.name,
					ScopeType: "dice",
					ScopeId:   "0",
				}, nil
			})
			defer pm2.Unpatch()
			p := &ServiceImpl{
				db: db,
			}
			got, err := p.RenderPipelineTemplate(context.Background(), &pb.PipelineTemplateRenderRequest{
				Name:      tt.name,
				ScopeID:   "0",
				ScopeType: "dice",
				Version:   "1.0",
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got.Data.Version.Name)
		})
	}
}
