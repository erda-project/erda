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

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/extension/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
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
	monkey.PatchInstanceMethod(reflect.TypeOf(client.DB), "IsExtensionPublicVersionExist", func(_ *gorm.DB, name string) (bool, error) {
		if name == "deprecated-addon" {
			return false, nil
		}
		return true, nil
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
		db:  cli,
		bdl: &bundle.Bundle{},
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

func Test_extensionService_SearchExtensions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionSearchRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionSearchResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionSearchRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionSearchResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.SearchExtensions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.SearchExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.SearchExtensions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_CreateExtension(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionCreateResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionCreateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionCreateResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.CreateExtension(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.CreateExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.CreateExtension() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryExtensionsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryExtensionsResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.QueryExtensionsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.QueryExtensionsResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensionsMenu(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryExtensionsMenuRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryExtensionsMenuResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.QueryExtensionsMenuRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.QueryExtensionsMenuResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensionsMenu(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensionsMenu() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensionsMenu() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_CreateExtensionVersion222(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionVersionCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionVersionCreateResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionVersionCreateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionVersionCreateResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.CreateExtensionVersion(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.CreateExtensionVersion222() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.CreateExtensionVersion222() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_GetExtensionVersion(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExtensionVersionRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExtensionVersionResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExtensionVersionRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExtensionVersionResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.GetExtensionVersion(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.GetExtensionVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.GetExtensionVersion() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensionVersions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionVersionQueryRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionVersionQueryResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionVersionQueryRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionVersionQueryResponse{
		//				// TODO: setup fields.
		//			},
		//			false,
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := servicehub.New()
			events := hub.Events()
			go func() {
				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
			}()
			err := <-events.Started()
			if err != nil {
				t.Error(err)
				return
			}
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensionVersions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensionVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensionVersions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_isExtensionPublic(t *testing.T) {
	p := &provider{
		db: cli,
	}

	type args struct {
		name     string
		isPublic bool
	}

	tests := []struct {
		name   string
		args   args
		expect bool
	}{
		{
			name: "update processing extension is private",
			args: args{
				name:     "mysql",
				isPublic: false,
			},
			expect: true,
		},
		{
			name: "all private",
			args: args{
				name:     "deprecated-addon",
				isPublic: false,
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.isExtensionPublic(tt.args.name, tt.args.isPublic)
			assert.NoError(t, err)
			if got != tt.expect {
				t.Fatalf("expect: %v, got: %v", tt.expect, got)
			}
		})
	}
}
