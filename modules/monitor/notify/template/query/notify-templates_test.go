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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/erda-project/erda/modules/monitor/notify/template/db"
	"github.com/erda-project/erda/modules/monitor/notify/template/model"
	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

var p *provider

func TestMain(t *testing.M) {
	p = new(provider)
	p.N = new(db.NotifyDB)
	p.N.DB, _ = gorm.Open("mysql", "localhost:3306")
	p.N.DB.LogMode(true)
	p.C = new(config)
	p.C.Files = []string{"/Users/terminus/go/src/terminus.io/github.com/erda-project/erda/pkg/dice-configs/notity/notify"}
	initTemplateMap()
	t.Run()
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

func Test_getAllNotifyTemplates(t *testing.T) {
	tests := []struct {
		name     string
		wantList []model.Model
	}{
		{
			name: "test_getAllNotifyTemplates",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotList := getAllNotifyTemplates(); gotList != nil {
				fmt.Printf("getAllNotifyTemplates() = %+v", gotList)
			}
		})
	}
}

func Test_getNotifyTemplateList(t *testing.T) {
	type args struct {
		scope string
		name  string
		nType string
	}
	tests := []struct {
		name     string
		args     args
		wantList []*model.GetNotifyRes
	}{
		{
			name: "test_getNotifyTemplateList",
			args: args{
				scope: "project",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotList := getNotifyTemplateList(tt.args.scope, tt.args.name, tt.args.nType); gotList != nil {
				fmt.Printf("getNotifyTemplateList() = %+v", gotList)
			}
		})
	}
}

func TestToNotifyConfig(t *testing.T) {
	type args struct {
		c *model.CreateUserDefineNotifyTemplate
	}
	tests := []struct {
		name    string
		args    args
		want    *db.NotifyConfig
		wantErr bool
	}{
		{
			name: "test_toNotifyConfig",
			args: args{
				c: &model.CreateUserDefineNotifyTemplate{
					Name:     "test_notify",
					Group:    "notify_group",
					Trigger:  []string{"notify"},
					Formats:  []map[string]string{},
					Title:    []string{"pipeline start"},
					Template: []string{"### {{projectName}}/{{appName}} pipeline {{pipelineID}} start"},
					Scope:    "project",
					ScopeID:  "3",
					Targets:  []string{"dingding,email,mbox"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToNotifyConfig(tt.args.c)
			if err != nil {
				t.Errorf("ToNotifyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("ToNotifyConfig() got is nil")
			}
		})
	}
}
