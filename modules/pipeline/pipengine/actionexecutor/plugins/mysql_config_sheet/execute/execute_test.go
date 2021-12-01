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

package execute

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pkg/log_collector"
)

type testResult struct {
	AffectedRows int64
	InsertId     int64
}

func (res *testResult) LastInsertId() (int64, error) {
	return res.InsertId, nil
}

func (res *testResult) RowsAffected() (int64, error) {
	return res.AffectedRows, nil
}

func Test_exec(t *testing.T) {
	type args struct {
		executors []SqlExecutor
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				executors: []SqlExecutor{
					{
						Sql: "drop procedure if exists idata122",
						OutParams: []apistructs.APIOutParam{
							{
								Key:        "name",
								Expression: ".lastInsertId",
							},
						},
						Asserts: []APIAssert{
							{
								Arg:      "name",
								Operator: "=",
								Value:    "1",
							},
						},
					},
					{
						Sql: "create procedure idata122()\nbegin\n    declare i int;\n    set i = 500;\n    while(i <= 500)\n        do\n            #     insert into test_thousand values(i);\n            INSERT INTO erda.ads_tag_user_lulu (user_id, gender, age, mbr_level, mbr_type, register_channel,\n                                                            lifecycle, is_qiwei, is_public_fans, is_potential,\n                                                            city_level, payment_180d, shop_time, clothing_payment,\n                                                            avg_distinct, max_payment, login_cnt_7d, last_shop_day,\n                                                            is_tmall_shop, offline_shop)\n            VALUES (i, ELT(CEILING(rand() * 2), '男', '女'), FLOOR(1 + (RAND() * 101)),\n                    ELT(CEILING(rand() * 4), 'L1', 'L2', 'L3', 'L4'), ELT(CEILING(rand() * 2), '非付费会员', '付费会员'),\n                    ELT(CEILING(rand() * 4), '商城', '微信公众号', '微信小程序', '会员app'),\n                    ELT(CEILING(rand() * 4), '新增用户', '沉睡用户', '活跃用户', '流失用户'),\n                    ELT(CEILING(rand() * 11), '厦门', '无锡', '杭州', '上海', '铁岭', '开封', '德州', '扬州', '芜湖', '丽江', '宁波'),\n                    ELT(CEILING(rand() * 2), '是', '否'), ELT(CEILING(rand() * 2), '是', '否'),\n                    ELT(CEILING(rand() * 5), '一线城市', '二线城市', '三线城市', '四线城市', '五线城市'),\n                    CEILING(0 + (RAND() * 100000)), CEILING(0 + (RAND() * 100)), CEILING(0 + (RAND() * 5000)),\n                    CEILING(0 + (RAND() * 9)), CEILING(0 + (RAND() * 5000)), CEILING(0 + (RAND() * 20)),\n                    CEILING(0 + (RAND() * 365)), ELT(CEILING(rand() * 2), '是', '否'),\n                    ELT(CEILING(rand() * 5), '包', '帽子', '鞋', '打底裤', '衣服'));\n\n            set i = i + 1;\n        end while;\nend "},
					{
						Sql: "create procedure idata123()\nbegin\n    declare i int;\n    set i = 500;\n    while(i <= 500)\n        do\n            #     insert into test_thousand values(i);\n            INSERT INTO erda.ads_tag_user_lulu (user_id, gender, age, mbr_level, mbr_type, register_channel,\n                                                            lifecycle, is_qiwei, is_public_fans, is_potential,\n                                                            city_level, payment_180d, shop_time, clothing_payment,\n                                                            avg_distinct, max_payment, login_cnt_7d, last_shop_day,\n                                                            is_tmall_shop, offline_shop)\n            VALUES (i, ELT(CEILING(rand() * 2), '男', '女'), FLOOR(1 + (RAND() * 101)),\n                    ELT(CEILING(rand() * 4), 'L1', 'L2', 'L3', 'L4'), ELT(CEILING(rand() * 2), '非付费会员', '付费会员'),\n                    ELT(CEILING(rand() * 4), '商城', '微信公众号', '微信小程序', '会员app'),\n                    ELT(CEILING(rand() * 4), '新增用户', '沉睡用户', '活跃用户', '流失用户'),\n                    ELT(CEILING(rand() * 11), '厦门', '无锡', '杭州', '上海', '铁岭', '开封', '德州', '扬州', '芜湖', '丽江', '宁波'),\n                    ELT(CEILING(rand() * 2), '是', '否'), ELT(CEILING(rand() * 2), '是', '否'),\n                    ELT(CEILING(rand() * 5), '一线城市', '二线城市', '三线城市', '四线城市', '五线城市'),\n                    CEILING(0 + (RAND() * 100000)), CEILING(0 + (RAND() * 100)), CEILING(0 + (RAND() * 5000)),\n                    CEILING(0 + (RAND() * 9)), CEILING(0 + (RAND() * 5000)), CEILING(0 + (RAND() * 20)),\n                    CEILING(0 + (RAND() * 365)), ELT(CEILING(rand() * 2), '是', '否'),\n                    ELT(CEILING(rand() * 5), '包', '帽子', '鞋', '打底裤', '衣服'));\n\n            set i = i + 1;\n        end while;\nend "},
					{
						Sql: "call idata122()"},
					{
						Sql: "call idata123()"},
					{
						Sql: "INSERT INTO `dice_pipeline_template_versions` (`id`, `template_id`, `name`, `version`, `spec`, `readme`, `created_at`,\n                                               `updated_at`)\nVALUES (8, 8, 'custom', '1.0', 'name: custom\\nversion: \\\"1.0\\\"\\ndesc: 自定义模板\\n\\ntemplate: |\\n  version: 1.1\\n  stages:',\n        '', '2020-10-14 10:05:17', '2020-10-14 10:05:17'),\n       (9, 9, 'java-boot-gradle-dice', '1.0',\n        'name: java-boot-gradle-dice\\nversion: \\\"1.0\\\"\\ndesc: springboot gradle 打包构建部署到 dice 的模板\\n\\ntemplate: |\\n\\n  version: 1.1\\n  stages:\\n    - stage:\\n        - git-checkout:\\n            params:\\n              depth: 1\\n    - stage:\\n        - java-build:\\n            version: \\\"1.0\\\"\\n            params:\\n              build_cmd:\\n                - ./gradlew bootJar\\n              jdk_version: 8\\n              workdir: ${git-checkout}\\n    - stage:\\n        - release:\\n            params:\\n              dice_yml: ${git-checkout}/dice.yml\\n              services:\\n                dice.yml中的服务名:\\n                  image: registry.cn-hangzhou.aliyuncs.com/terminus/terminus-openjdk:v11.0.6\\n                  copys:\\n                    - ${java-build:OUTPUT:buildPath}/build/jar包的路径/jar包的名称:/target/jar包的名称\\n                  cmd: java -jar /target/jar包的名称\\n\\n    - stage:\\n        - dice:\\n            params:\\n              release_id: ${release:OUTPUT:releaseID}\\n\\n\\nparams:\\n\\n  - name: pipeline_version\\n    desc: 生成的pipeline的版本\\n    default: \\\"1.1\\\"\\n    required: false\\n\\n  - name: pipeline_cron\\n    desc: 定时任务的cron表达式\\n    required: false\\n\\n  - name: pipeline_scheduling\\n    desc: 流水线调度策略\\n    required: false\\n',\n        '', '2020-10-14 10:05:49', '2020-10-14 13:35:57'),\n       (10, 10, 'java-boot-maven-dice', '1.0',\n        'name: java-boot-maven-dice\\nversion: \\\"1.0\\\"\\ndesc: springboot maven 打包构建部署到 dice 的模板\\n\\ntemplate: |\\n\\n  version: 1.1\\n  stages:\\n    - stage:\\n        - git-checkout:\\n            params:\\n              depth: 1\\n\\n    - stage:\\n        - java-build:\\n            version: \\\"1.0\\\"\\n            params:\\n              build_cmd:\\n                - mvn package\\n              jdk_version: 8\\n              workdir: ${git-checkout}\\n\\n    - stage:\\n        - release:\\n            params:\\n              dice_yml: ${git-checkout}/dice.yml\\n              services:\\n                dice.yml中的服务名:\\n                  image: registry.cn-hangzhou.aliyuncs.com/terminus/terminus-openjdk:v11.0.6\\n                  copys:\\n                    - ${java-build:OUTPUT:buildPath}/target/jar包的名称:/target/jar包的名称\\n                  cmd: java -jar /target/jar包的名称\\n\\n    - stage:\\n        - dice:\\n            params:\\n              release_id: ${release:OUTPUT:releaseID}\\n\\n\\nparams:\\n\\n  - name: pipeline_version\\n    desc: 生成的pipeline的版本\\n    default: \\\"1.1\\\"\\n    required: false\\n\\n  - name: pipeline_cron\\n    desc: 定时任务的cron表达式\\n    required: false\\n\\n  - name: pipeline_scheduling\\n    desc: 流水线调度策略\\n    required: false\\n',\n        '', '2020-10-14 10:06:16', '2020-10-14 13:36:28'),\n       (11, 11, 'java-tomcat-maven-dice', '1.0',\n        'name: java-tomcat-maven-dice\\nversion: \\\"1.0\\\"\\ndesc: java maven 打包构建放入 tomcat 部署到 dice 的模板\\n\\ntemplate: |\\n\\n  version: \\\"1.1\\\"\\n  stages:\\n    - stage:\\n        - git-checkout:\\n            params:\\n              depth: 1\\n\\n    - stage:\\n        - java-build:\\n            version: \\\"1.0\\\"\\n            params:\\n              build_cmd:\\n                - mvn clean package\\n              jdk_version: 8\\n              workdir: ${git-checkout}\\n\\n    - stage:\\n        - release:\\n            params:\\n              dice_yml: ${git-checkout}/dice.yml\\n              services:\\n                dice_yaml中的服务名称:\\n                  image: tomcat:jdk8-openjdk-slim\\n                  copys:\\n                    - ${java-build:OUTPUT:buildPath}/target/war包的名称.war:/usr/local/tomcat/webapps\\n                  cmd: mv /usr/local/tomcat/webapps/war包的名称.war /usr/local/tomcat/webapps/ROOT.war && /usr/local/tomcat/bin/catalina.sh run\\n\\n    - stage:\\n        - dice:\\n            params:\\n              release_id: ${release:OUTPUT:releaseID}\\n\\n',\n        '', '2020-10-14 10:06:45', '2020-10-14 10:46:26'),\n       (12, 12, 'js-herd-release-dice', '1.0',\n        'name: js-herd-release-dice\\nversion: \\\"1.0\\\"\\ndesc: js 直接运行并部署到 dice 的模板\\n\\ntemplate: |\\n\\n  version: \\\"1.1\\\"\\n  stages:\\n    - stage:\\n        - git-checkout:\\n            alias: git-checkout\\n            version: \\\"1.0\\\"\\n    - stage:\\n        - js-build:\\n            alias: js-build\\n            version: \\\"1.0\\\"\\n            params:\\n              build_cmd:\\n                - cnpm i\\n              workdir: ${git-checkout}\\n\\n    - stage:\\n        - release:\\n            alias: release\\n            params:\\n              dice_yml: ${git-checkout}/dice.yml\\n              services:\\n                dice.yml中的服务名:\\n                  cmd: cd /root/js-build && ls && npm run dev\\n                  copys:\\n                    - ${js-build}:/root/\\n                  image: registry.cn-hangzhou.aliyuncs.com/terminus/terminus-herd:1.1.8-node12\\n    - stage:\\n        - dice:\\n            alias: dice\\n            params:\\n              release_id: ${release:OUTPUT:releaseID}\\n\\n',\n        '', '2020-10-14 10:07:11', '2020-10-14 10:07:11'),\n       (13, 13, 'js-spa-release-dice', '1.0',\n        'name: js-spa-release-dice\\nversion: \\\"1.0\\\"\\ndesc: js 进行打包构建到 nginx 并部署到 dice 的模板\\n\\ntemplate: |\\n\\n  version: \\\"1.1\\\"\\n  stages:\\n    - stage:\\n        - git-checkout:\\n            alias: git-checkout\\n            version: \\\"1.0\\\"\\n    - stage:\\n        - js-build:\\n            alias: js-build\\n            version: \\\"1.0\\\"\\n            params:\\n              build_cmd:\\n                - cnpm i\\n                - cnpm run build\\n              workdir: ${git-checkout}\\n\\n    - stage:\\n        - release:\\n            alias: release\\n            params:\\n              dice_yml: ${git-checkout}/dice.yml\\n              services:\\n                dice.yml中的服务名:\\n                  cmd: sed -i \\\"s^server_name .*^^g\\\" /etc/nginx/conf.d/nginx.conf.template && envsubst \\\"`printf \\'$%s\\' $(bash -c \\\"compgen -e\\\")`\\\" < /etc/nginx/conf.d/nginx.conf.template > /etc/nginx/conf.d/default.conf && /usr/local/openresty/bin/openresty -g \\'daemon off;\n\\'\\n\\n                  copys:\\n                    - ${js-build}/(build 产出的目录):/usr/share/nginx/html/\\n                    - ${js-build}/nginx.conf.template:/etc/nginx/conf.d/\\n                  image: registry.cn-hangzhou.aliyuncs.com/dice-third-party/terminus-nginx:0.2\\n    - stage:\\n        - dice:\\n            alias: dice\\n            params:\\n              release_id: ${release:OUTPUT:releaseID}\\n\\n','','2020-10-14 10:07:38','2020-10-14 10:07:38');"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client = sql.DB{}
			var tx = sql.Tx{}
			var rows = sql.Rows{}

			patch6 := monkey.PatchInstanceMethod(reflect.TypeOf(&rows), "Next", func(tx *sql.Rows) bool {
				return false
			})
			defer patch6.Unpatch()

			patch2 := monkey.PatchInstanceMethod(reflect.TypeOf(&tx), "QueryContext", func(tx *sql.Tx, ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
				return &rows, nil
			})
			defer patch2.Unpatch()

			patch3 := monkey.PatchInstanceMethod(reflect.TypeOf(&tx), "Rollback", func(tx *sql.Tx) error {
				return nil
			})
			defer patch3.Unpatch()

			patch4 := monkey.PatchInstanceMethod(reflect.TypeOf(&tx), "Exec", func(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
				return &testResult{
					1, 1,
				}, nil
			})
			defer patch4.Unpatch()

			patch5 := monkey.PatchInstanceMethod(reflect.TypeOf(&tx), "Commit", func(tx *sql.Tx) error {
				return nil
			})
			defer patch5.Unpatch()

			patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(&client), "BeginTx", func(client *sql.DB, ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
				return &tx, nil
			})
			defer patch1.Unpatch()
			var meta = &Meta{}

			var logger = logrus.Logger{}
			pathLog := monkey.PatchInstanceMethod(reflect.TypeOf(&logger), "IsLevelEnabled", func(log *logrus.Logger, level logrus.Level) bool {
				return false
			})
			defer pathLog.Unpatch()
			var entry = &logrus.Entry{
				Logger: &logger,
			}
			var ctx = context.WithValue(context.Background(), log_collector.CtxKeyLogger, entry)

			if err := exec(ctx, &client, meta, tt.args.executors); (err != nil) != tt.wantErr {
				t.Errorf("exec() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
