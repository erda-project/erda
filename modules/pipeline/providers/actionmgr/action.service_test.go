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

package actionmgr

import (
	"context"

	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/action/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/actionmgr/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

func Test_actionsOrderByLocationIndex(t *testing.T) {
	type args struct {
		locations []string
		data      []*pb.Action
	}
	tests := []struct {
		name string
		args args
		want []*pb.Action
	}{
		{
			name: "test order",
			args: args{
				locations: []string{
					"fdp/",
					"default/",
				},
				data: []*pb.Action{
					{
						Location: "default/",
						Name:     "a",
					},
					{
						Location: "default/",
						Name:     "b",
					},
					{
						Location: "fdp/",
						Name:     "a",
					},
				},
			},
			want: []*pb.Action{
				{
					Location: "fdp/",
					Name:     "a",
				},
				{
					Location: "default/",
					Name:     "a",
				},
				{
					Location: "default/",
					Name:     "b",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := actionsOrderByLocationIndex(tt.args.locations, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actionsOrderByLocationIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actionService_Save(t *testing.T) {
	stringTime := "2017-08-30 16:40:41"

	loc, _ := time.LoadLocation("Local")

	nowTime, _ := time.ParseInLocation("2006-01-02 15:04:05", stringTime, loc)

	type args struct {
		ctx context.Context
		req *pb.PipelineActionSaveRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *db.PipelineAction
		wantErr bool
	}{
		{
			name: "test normal",
			args: args{
				ctx: apis.WithInternalClientContext(context.Background(), "test"),
				req: &pb.PipelineActionSaveRequest{
					Readme:   "test",
					Dice:     "### job 配置项\njobs:\n  java-builds:\n    image: registry.erda.cloud/erda-actions/java-build-action:1.0-20220316-5922882\n    resources:\n      cpu: 0.5\n      mem: 2048\n      disk: 1024",
					Spec:     "name: java-builds\nversion: \"1.0\"\ntype: action\ncategory: build_management\ndisplayName: ${{ i18n.displayName }}\nlogoUrl: //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/09/28/d74a0d23-523d-4451-9647-f32f47b2000d.png\ndesc: ${{ i18n.desc }}\nlabels:\n  maintainer: xxx\n  project_level_app: true\npublic: true\nsupportedVersions: # Deprecated. Please use supportedErdaVersions instead.\n  - \">= 3.9\"\nsupportedErdaVersions:\n  - \">= 1.0\"\n\nparams:\n  - name: jdk_version\n    desc: ${{ i18n.params.jdk_version.desc }}\n    required: true\n  - name: build_cmd\n    type: string_array\n    desc: ${{ i18n.params.build_cmd.desc }}\n    required: true\n  - name: workdir\n    desc: ${{ i18n.params.workdir.desc }}\n    default: \".\"\naccessibleAPIs:\n  # api compatibility check\n  - path: /api/gateway/check-compatibility\n    method: POST\n    schema: http\n\noutputs:\n  - name: buildPath\n    desc: ${{ i18n.outputs.buildPath.desc }}\n  - name: JAVA_OPTS\n    desc: ${{ i18n.outputs.JAVA_OPTS.desc }}\n\nlocale:\n  zh-CN:\n    desc: 针对 java 工程的编译打包任务\n    displayName: Java 工程编译打包\n    outputs.JAVA_OPTS.desc: 需要假如的监控agent入参\n    outputs.buildPath.desc: 包的位置\n    params.build_cmd.desc: 构建命令,如:./gradlew :user:build，mvn package\n    params.jdk_version.desc: 构建使用的jdk版本,支持8,11\n    params.workdir.desc: 工作目录一般为仓库路径 ${git-checkout}\n\n  en-US:\n    desc: Build and package a Java project\n    displayName: Java project build and package\n    outputs.JAVA_OPTS.desc: The parameters for monitoring agent\n    outputs.buildPath.desc: The path of the package\n    params.build_cmd.desc: The build command, such as ./gradlew :user:build, mvn package\n    params.jdk_version.desc: The JDK version to build, support 8, 11\n    params.workdir.desc: The working directory, usually the repository path ${git-checkout}\n",
					Location: "fdp/",
				},
			},
			want: &db.PipelineAction{
				ID:          "1",
				TimeCreated: nowTime,
				TimeUpdated: nowTime,
				Name:        "java-builds",
				Category:    "build_management",
				DisplayName: "Java 工程编译打包",
				LogoUrl:     "//terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/09/28/d74a0d23-523d-4451-9647-f32f47b2000d.png",
				Desc:        "针对 java 工程的编译打包任务",
				Readme:      "test",
				Dice:        "### job 配置项\njobs:\n  java-builds:\n    image: registry.erda.cloud/erda-actions/java-build-action:1.0-20220316-5922882\n    resources:\n      cpu: 0.5\n      mem: 2048\n      disk: 1024",
				Spec:        "name: java-builds\nversion: \"1.0\"\ntype: action\ncategory: build_management\ndisplayName: ${{ i18n.displayName }}\nlogoUrl: //terminus-paas.oss-cn-hangzhou.aliyuncs.com/paas-doc/2020/09/28/d74a0d23-523d-4451-9647-f32f47b2000d.png\ndesc: ${{ i18n.desc }}\nlabels:\n  maintainer: xxx\n  project_level_app: true\npublic: true\nsupportedVersions: # Deprecated. Please use supportedErdaVersions instead.\n  - \">= 3.9\"\nsupportedErdaVersions:\n  - \">= 1.0\"\n\nparams:\n  - name: jdk_version\n    desc: ${{ i18n.params.jdk_version.desc }}\n    required: true\n  - name: build_cmd\n    type: string_array\n    desc: ${{ i18n.params.build_cmd.desc }}\n    required: true\n  - name: workdir\n    desc: ${{ i18n.params.workdir.desc }}\n    default: \".\"\naccessibleAPIs:\n  # api compatibility check\n  - path: /api/gateway/check-compatibility\n    method: POST\n    schema: http\n\noutputs:\n  - name: buildPath\n    desc: ${{ i18n.outputs.buildPath.desc }}\n  - name: JAVA_OPTS\n    desc: ${{ i18n.outputs.JAVA_OPTS.desc }}\n\nlocale:\n  zh-CN:\n    desc: 针对 java 工程的编译打包任务\n    displayName: Java 工程编译打包\n    outputs.JAVA_OPTS.desc: 需要假如的监控agent入参\n    outputs.buildPath.desc: 包的位置\n    params.build_cmd.desc: 构建命令,如:./gradlew :user:build，mvn package\n    params.jdk_version.desc: 构建使用的jdk版本,支持8,11\n    params.workdir.desc: 工作目录一般为仓库路径 ${git-checkout}\n\n  en-US:\n    desc: Build and package a Java project\n    displayName: Java project build and package\n    outputs.JAVA_OPTS.desc: The parameters for monitoring agent\n    outputs.buildPath.desc: The path of the package\n    params.build_cmd.desc: The build command, such as ./gradlew :user:build, mvn package\n    params.jdk_version.desc: The JDK version to build, support 8, 11\n    params.workdir.desc: The working directory, usually the repository path ${git-checkout}\n",
				Location:    "fdp/",
				IsPublic:    true,
				IsDefault:   false,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &actionService{}

			var dbClient db.Client
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "ListPipelineAction", func(client *db.Client, req *pb.PipelineActionListRequest, ops ...mysqlxorm.SessionOption) ([]db.PipelineAction, error) {
				return nil, nil
			})
			defer patch.Unpatch()

			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(&dbClient), "InsertPipelineAction", func(client *db.Client, action *db.PipelineAction, ops ...mysqlxorm.SessionOption) error {
				action.ID = tt.want.ID
				action.SoftDeletedAt = 0
				action.TimeUpdated = tt.want.TimeUpdated
				action.TimeCreated = tt.want.TimeCreated
				return nil
			})
			defer patch1.Unpatch()

			s.dbClient = &dbClient

			got, err := s.Save(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotAction, _ := tt.want.Convert(false)
			assert.Equal(t, gotAction.ID, got.Action.ID)
		})
	}
}
