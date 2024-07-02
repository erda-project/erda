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

package extension

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/alecthomas/assert"

	"github.com/erda-project/erda/apistructs"
)

func TestAddSyncExtension(t *testing.T) {
	var s = &provider{}
	file := NewFileExtensionSource(s)
	RegisterExtensionSource(file)

	patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "InitExtension", func(s *provider, addr string, forceUpdate bool) error {
		return fmt.Errorf("test")
	})
	defer patch1.Unpatch()

	dir, err := os.MkdirTemp(os.TempDir(), "*")
	assert.NoError(t, err)

	err = AddSyncExtension(dir)
	assert.Error(t, err)
}

func TestFileExtensionSource_add(t *testing.T) {

	var s = &provider{}
	file := NewFileExtensionSource(s)

	patch1 := monkey.PatchInstanceMethod(reflect.TypeOf(s), "InitExtension", func(s *provider, addr string, forceUpdate bool) error {
		return nil
	})
	defer patch1.Unpatch()

	dir, err := os.MkdirTemp(os.TempDir(), "*")
	assert.NoError(t, err)

	err = file.add(dir)
	assert.NoError(t, err)
}

func Test_isDir(t *testing.T) {
	dir, err := os.MkdirTemp(os.TempDir(), "*")
	assert.NoError(t, err)
	got := isDir(dir)
	assert.True(t, got)

	file, err := os.CreateTemp(os.TempDir(), "*")
	assert.NoError(t, err)
	got = isDir(file.Name())
	assert.True(t, !got)
}

func TestGitExtensionSource_match(t *testing.T) {
	type args struct {
		addr string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_http",
			args: args{
				addr: "http://test.git",
			},
			want: true,
		},
		{
			name: "test_https",
			args: args{
				addr: "http://test.git",
			},
			want: true,
		},
		{
			name: "test_git",
			args: args{
				addr: "git@github.com:erda-project/erda.git",
			},
			want: true,
		},
		{
			name: "test_file",
			args: args{
				addr: "/test/aaa",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GitExtensionSource{}
			if got := g.match(tt.args.addr); got != tt.want {
				t.Errorf("match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertFieldI18n(t *testing.T) {
	var key = "${{ i18n.displayName }}"
	var localeMap = map[string]map[string]string{
		"zh-CN": {
			"desc":                           "Erda MySQL Migration 工具",
			"displayName":                    "Erda MySQL 数据迁移",
			"outputs.success.desc":           "是否成功",
			"params.database.desc":           "要执行 migration 的库名",
			"params.lint_config.desc":        "Erda MySQL Migration Lint 的配置文件",
			"params.migrationdir.desc":       "脚本存放目录",
			"params.modules.desc":            "要进行数据迁移的模块, 为空时对 migrationDir 目录下的所有模块进行数据迁移",
			"params.mysql_host.desc":         "mysql 服务地址",
			"params.mysql_password.desc":     "mysql 密码",
			"params.mysql_port.desc":         "mysql 服务端口",
			"params.mysql_username.desc":     "mysql 用户名",
			"params.retry_timeout.desc":      "连接数据库最长超时时间",
			"params.skip_lint.desc":          "跳过 Erda MySQL 规约检查. 注意标明 \"MIGRATION_BASE\" 的脚本总是不会被检查.",
			"params.skip_migration.desc":     "跳过预执行和正式执行",
			"params.skip_pre_migration.desc": "跳过预执行",
			"params.skip_sandbox.desc":       "跳过沙盒预演",
			"params.workdir.desc":            "工作目录, 如 ${git-checkout}",
		},
		"en-US": {
			"desc":                           "Erda MySQL Migration tool",
			"displayName":                    "Erda MySQL data migration",
			"outputs.success.desc":           "whether succeed",
			"params.database.desc":           "The name of the library to execute migration",
			"params.lint_config.desc":        "Configuration file of Erda MySQL migration lint",
			"params.migrationdir.desc":       "Script directory",
			"params.modules.desc":            "The module to be migrated, or all modules in the migrationDir directory if empty.",
			"params.mysql_host.desc":         "Mysql host",
			"params.mysql_password.desc":     "Mysql password",
			"params.mysql_port.desc":         "Mysql port",
			"params.mysql_username.desc":     "Mysql user name",
			"params.retry_timeout.desc":      "The maximum timeout for connecting to the database",
			"params.skip_lint.desc":          "Skip Erda MySQL protocol check. Note that scripts marked \"MIGRATION_BASE\" will not be checked.",
			"params.skip_migration.desc":     "Skip pre-execution and formal execution",
			"params.skip_pre_migration.desc": "Skip pre-execution",
			"params.skip_sandbox.desc":       "Skip the sandbox preview",
			"params.workdir.desc":            "Working directory, such as ${git-checkout}",
		},
	}
	data := ConvertFieldI18n(key, localeMap)
	t.Log(data)
}

func TestIsExtensionPublic(t *testing.T) {
	tests := []struct {
		name string
		spec *apistructs.Spec
		want bool
	}{
		{
			name: "public extension",
			spec: &apistructs.Spec{Public: true},
			want: true,
		},
		{
			name: "private extension",
			spec: &apistructs.Spec{Public: false},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := isExtensionPublic(test.spec)
			if res != test.want {
				t.Errorf("got %t, want %t", res, test.want)
			}
		})
	}
}
