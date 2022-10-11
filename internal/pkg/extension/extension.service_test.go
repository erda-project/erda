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
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/extension/db"
	"github.com/erda-project/erda/pkg/i18n"
)

var cli = mockClient()

func mockClient() *db.Client {
	client := &db.Client{
		DB: &gorm.DB{},
	}

	monkey.PatchInstanceMethod(reflect.TypeOf(client), "CreateExtension", func(_ *db.Client, extension *db.Extension) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "QueryExtensions", func(_ *db.Client, all bool, typ string, labels string) ([]db.Extension, error) {
		return []db.Extension{
			{
				Name: "extension",
			},
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetExtension", func(_ *db.Client, name string) (*db.Extension, error) {
		return &db.Extension{
			Name: name,
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "DeleteExtension", func(_ *db.Client, name string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetExtensionVersion", func(_ *db.Client, name string, version string) (*db.ExtensionVersion, error) {
		return &db.ExtensionVersion{Name: name, Version: version}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetExtensionDefaultVersion", func(_ *db.Client, name string) (*db.ExtensionVersion, error) {
		return &db.ExtensionVersion{Name: name, IsDefault: true}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "SetUnDefaultVersion", func(_ *db.Client, name string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "CreateExtensionVersion", func(_ *db.Client, version *db.ExtensionVersion) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "DeleteExtensionVersion", func(_ *db.Client, name, version string) error {
		return nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "QueryExtensionVersions", func(_ *db.Client, name string, all bool, orderByVersionDesc bool) ([]db.ExtensionVersion, error) {
		return []db.ExtensionVersion{
			{
				Name: name,
			},
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "GetExtensionVersionCount", func(_ *db.Client, name string) (int64, error) {
		return 1, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "QueryAllExtensions", func(_ *db.Client) ([]db.ExtensionVersion, error) {
		return []db.ExtensionVersion{
			{
				Name: "extension",
			},
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client), "ListExtensionVersions", func(_ *db.Client, names []string, all bool) (map[string][]db.ExtensionVersion, error) {
		res := make(map[string][]db.ExtensionVersion)
		for _, name := range names {
			res[name] = []db.ExtensionVersion{{Name: name}}
		}
		return res, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(client.DB), "Save", func(_ *gorm.DB, value interface{}) *gorm.DB {
		return &gorm.DB{
			Error: nil,
		}
	})

	return client
}

func TestCreateExtensionVersionByRequest(t *testing.T) {
	type arg struct {
		ext *pb.ExtensionVersionCreateRequest
	}
	testCases := []struct {
		name string
		arg  arg
	}{
		{
			name: "create action",
			arg: arg{
				ext: &pb.ExtensionVersionCreateRequest{
					Name:        "custom-script",
					Version:     "1.0",
					IsDefault:   true,
					ForceUpdate: true,
					SpecYml: `name: custom-script
version: "1.0"
type: action
displayName: ${{ i18n.displayName }}
category: custom_task
desc: ${{ i18n.desc }}
public: true
labels:
  autotest: true
  configsheet: true
  project_level_app: true
  eci_disable: true

supportedVersions: # Deprecated. Please use supportedErdaVersions instead.
  - ">= 3.5"
supportedErdaVersions:
  - ">= 1.0"

params:
  - name: command
    desc: ${{ i18n.params.command.desc }}
locale:
  zh-CN:
    desc: 运行自定义命令
    displayName: 自定义任务
    params.command.desc: 运行的命令
  en-US:
    desc: Run custom commands
    displayName: Custom task
    params.command.desc: Command
`,
					DiceYml: `### job 配置项
jobs:
  custom-script:
    image: registry.erda.cloud/custom-script-action:1.0
    resources:
      cpu: 0.1
      mem: 1024
      disk: 1024
`,
				},
			},
		},
	}
	p := &provider{
		db: cli,
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := p.CreateExtensionVersionByRequest(tc.arg.ext)
			assert.NoError(t, err)
		})
	}
}

func TestDeleteExtensionVersion(t *testing.T) {
	p := &provider{
		db: cli,
	}
	err := p.DeleteExtensionVersion("extension", "1.0")
	assert.NoError(t, err)
}

func TestGetExtension(t *testing.T) {
	p := &provider{
		db: cli,
	}
	ext, err := p.GetExtension("extension", "1.0", true)
	assert.NoError(t, err)
	assert.Equal(t, "extension", ext.Name)
}

func TestQueryExtensionList(t *testing.T) {
	p := &provider{
		db: cli,
	}
	exts, err := p.QueryExtensionList(true, apistructs.SpecActionType.String(), "")
	assert.NoError(t, err)
	assert.Equal(t, "extension", exts[0].Name)
}

func TestQueryExtensionVersions(t *testing.T) {
	p := &provider{
		db: cli,
	}
	_, err := p.QueryExtensionVersions(context.Background(), &pb.ExtensionVersionQueryRequest{
		Name: "extension",
	})
	assert.NoError(t, err)
}

func TestCreate(t *testing.T) {
	p := &provider{
		db: cli,
	}
	ext, err := p.Create(&pb.ExtensionCreateRequest{
		Name: "extension",
		Type: apistructs.SpecActionType.String(),
	})
	assert.NoError(t, err)
	assert.Equal(t, "extension", ext.Name)
}

func TestGetExtensionDefaultVersion(t *testing.T) {
	p := &provider{
		db: cli,
	}
	ext, err := p.GetExtensionDefaultVersion("extension", true)
	assert.NoError(t, err)
	assert.Equal(t, "extension", ext.Name)
}

func TestMenuExtWithLocale(t *testing.T) {
	p := &provider{
		db: cli,
	}
	menus, err := p.MenuExtWithLocale([]*pb.Extension{
		{Name: "extension", Type: apistructs.SpecActionType.String(), Category: "source_code_management"},
	}, &i18n.LocaleResource{}, true)
	t.Log(menus)
	assert.NoError(t, err)
	assert.Equal(t, "source_code_management", menus[apistructs.SpecActionType.String()][0].Name)
}
