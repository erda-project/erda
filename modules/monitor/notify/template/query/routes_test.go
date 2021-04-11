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

package query

import (
	"bou.ke/monkey"
	"fmt"
	"github.com/erda-project/erda-infra/modcom/api"
	"github.com/erda-project/erda/modules/monitor/notify/template/db"
	"github.com/erda-project/erda/modules/monitor/notify/template/model"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
	"gotest.tools/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

var p *provider

func TestMain(t *testing.M) {
	p = new(provider)
	p.N = new(db.NotifyDB)
	p.N.DB, _ = gorm.Open("mysql", "localhost:3306")
	p.N.DB.LogMode(true)
	p.C = new(config)
	p.C.Files = []string{"/Users/terminus/go/src/terminus.io/dice/monitor/conf/monitor/monitor/dice-configs/notity"}
	t.Run()
}

func Test_provider_createNotify(t *testing.T) {
	type args struct {
		r      *http.Request
		params model.CreateNotifyReq
	}
	tests := []struct {
		name     string
		provider provider
		args     args
		want     interface{}
	}{
		{
			name:     "testCreate",
			provider: *p,
			args: args{
				r: new(http.Request),
				params: model.CreateNotifyReq{
					ScopeID:       "16",
					Scope:         "app",
					TemplateID:    []string{"addon_mysql_memory"},
					NotifyName:    "pjytest456",
					NotifyGroupID: 1,
					Channels:      []string{"sms", "email"},
				},
			},
		},
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(new(db.NotifyDB)), "GetNotifyGroup", func(_ *db.NotifyDB, id int64) (*db.NotifyGroup, error) {
		return &db.NotifyGroup{
			TargetData: `[{"type":"dingding","values":[{"receiver":"https://oapi.dingtalk.com","secret":"e69d6e340a07c47fb"}]}]`,
		}, nil
	})
	monkey.PatchInstanceMethod(reflect.TypeOf(new(db.NotifyDB)), "CheckNotifyNameExist", func(_ *db.NotifyDB, scopeID, scopeName, notifyName string) (bool, error) {
		return false, nil
	})
	//monkey.PatchInstanceMethod(reflect.TypeOf(new(db.NotifyDB)), "CreateNotifyRecord", func(_ *db.NotifyDB, record *db.Notify) error {
	//	return nil
	//})
	monkey.Patch("CreateNotifyRecord", nil)

	for _, tt := range tests {
		p := &provider{
			C:    tt.provider.C,
			N:    tt.provider.N,
			L:    tt.provider.L,
			t:    tt.provider.t,
			bdl:  tt.provider.bdl,
			cmdb: tt.provider.cmdb,
		}
		res := p.createNotify(tt.args.r, tt.args.params)
		resp := res.(*api.Response)
		fmt.Println(resp.Success)
		assert.Equal(t, true, resp.Success)
	}
}

func initTemplateMap() {
	templateMap = make(map[string]model.Model)
	for _, file := range p.C.Files {
		f, err := os.Stat(file)
		if err != nil {
			fmt.Println(fmt.Errorf("fail to load notify file: %s", err))
		}
		if f.IsDir() {
			err := filepath.Walk(file, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				f, err := ioutil.ReadFile(p)
				var model model.Model
				err = yaml.Unmarshal(f, &model)
				if err != nil {
					return err
				}
				templateMap[model.ID] = model
				fmt.Printf("id:%v,content:%+v", model.ID, model)
				return nil
			})
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func Test_provider_getUserNotifyList(t *testing.T) {
	initTemplateMap()

	type args struct {
		r      *http.Request
		params struct {
			Scope   string `query:"scope" validate:"required"`
			ScopeID string `query:"scopeId" validate:"required"`
		}
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   interface{}
	}{
		{
			name:   "testGetUserList",
			fields: *p,
			args: args{
				r: new(http.Request),
				params: struct {
					Scope   string `query:"scope" validate:"required"`
					ScopeID string `query:"scopeId" validate:"required"`
				}{
					Scope:   "app",
					ScopeID: "16",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				C:    tt.fields.C,
				N:    tt.fields.N,
				L:    tt.fields.L,
				t:    tt.fields.t,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
			}
			if got := p.getUserNotifyList(tt.args.r, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getUserNotifyList() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_provider_notifyEnable(t *testing.T) {
	type args struct {
		r      *http.Request
		params struct {
			ID      int64  `param:"id" validate:"required"`
			ScopeID string `query:"scopeId" validate:"required"`
			Scope   string `query:"scope" validate:"required"`
		}
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   interface{}
	}{
		{
			name:   "testEnable",
			fields: *p,
			args: args{
				r: new(http.Request),
				params: struct {
					ID      int64  `param:"id" validate:"required"`
					ScopeID string `query:"scopeId" validate:"required"`
					Scope   string `query:"scope" validate:"required"`
				}{
					ID:      41,
					ScopeID: "16",
					Scope:   "app",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				C:    tt.fields.C,
				N:    tt.fields.N,
				L:    tt.fields.L,
				t:    tt.fields.t,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
			}
			if got := p.notifyEnable(tt.args.r, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("notifyEnable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_createUserDefineNotifyTemplate(t *testing.T) {
	type args struct {
		r      *http.Request
		params model.CreateUserDefineNotifyTemplate
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   interface{}
	}{
		{
			name:   "testCreateUserDefineTemplate",
			fields: *p,
			args: args{
				r: new(http.Request),
				params: model.CreateUserDefineNotifyTemplate{
					Name:    "app_test_2",
					Group:   "{{cluster_name}}-{{container_id}}",
					Trigger: []string{"alert", "recover"},
					Formats: []map[string]string{
						{"cpu_usage_percent_avg": "percent:1",
							"limit_value":  "fraction:1",
							"container_id": "string:6"},
						{"cpu_usage_percent_avg2": "percent:1",
							"limit_value2":  "fraction:1",
							"container_id2": "string:6"},
					},
					Title: []string{"【测试自定义模版】", "【测试自定义模版告警恢复】"},
					Template: []string{
						"【ElasticSearch实例CPU使用率异常告警】\n\n        组件: {{addon_type}}\n\n        实例: {{pod_namespace}} - {{pod_name}}\n\n        CPU平均使用率: {{cpu_usage_percent_avg}}",
						"【ElasticSearch实例CPU使用率异常告警恢复】\n\n        组件: {{addon_type}}\n\n        实例: {{pod_namespace}} - {{pod_name}}\n\n        CPU平均使用率: {{cpu_usage_percent_avg}}",
					},
					ScopeID: "16",
					Scope:   "app",
					Targets: []string{
						"dingding,email",
						"webhook",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				C:    tt.fields.C,
				N:    tt.fields.N,
				L:    tt.fields.L,
				t:    tt.fields.t,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
			}
			if got := p.createUserDefineNotifyTemplate(tt.args.r, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createUserDefineNotifyTemplate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_getNotifyDetail(t *testing.T) {
	type args struct {
		r      *http.Request
		params struct {
			Id int64 `param:"id"`
		}
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   interface{}
	}{
		{
			name:   "testGetNotifyDetail",
			fields: *p,
			args: args{
				r: new(http.Request),
				params: struct {
					Id int64 `param:"id"`
				}{
					Id: 44,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				C:    tt.fields.C,
				N:    tt.fields.N,
				L:    tt.fields.L,
				t:    tt.fields.t,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
			}
			if got := p.getNotifyDetail(tt.args.r, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getNotifyDetail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_getAllGroups(t *testing.T) {
	type args struct {
		r      *http.Request
		params struct {
			Scope   string `query:"scope"`
			ScopeId string `query:"scopeId"`
		}
	}
	tests := []struct {
		name   string
		fields provider
		args   args
		want   interface{}
	}{
		{
			name:   "testGetAllGroups",
			fields: *p,
			args: args{
				r: new(http.Request),
				params: struct {
					Scope   string `query:"scope"`
					ScopeId string `query:"scopeId"`
				}{
					Scope:   "app",
					ScopeId: "16",
				},
			},
		},
	}
	tests[0].args.r.Header = make(map[string][]string)
	tests[0].args.r.Header["Org-ID"] = []string{"1"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{
				C:    tt.fields.C,
				N:    tt.fields.N,
				L:    tt.fields.L,
				t:    tt.fields.t,
				bdl:  tt.fields.bdl,
				cmdb: tt.fields.cmdb,
			}
			if got := p.getAllGroups(tt.args.r, tt.args.params); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAllGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
