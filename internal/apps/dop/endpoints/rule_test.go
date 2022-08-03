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

package endpoints

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	gomock "github.com/golang/mock/gomock"

	pb "github.com/erda-project/erda-proto-go/dop/rule/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func TestEndpoints_FireRule(t *testing.T) {
	e := &Endpoints{}
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(e), "SetConfigInfo",
		func(d *Endpoints, eventInfo EventInfo) (map[string]interface{}, error) {
			return map[string]interface{}{}, nil
		},
	)
	defer p1.Unpatch()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ruleService := NewMockRuleServiceServer(ctrl)
	ruleService.EXPECT().Fire(gomock.Any(), gomock.Any()).AnyTimes().Return(&pb.FireResponse{Output: []bool{true}}, nil)
	e.ruleExecutor = ruleService
	type args struct {
		ctx       context.Context
		content   interface{}
		eventInfo EventInfo
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				content: map[string]interface{}{
					"title": "1",
				},
				eventInfo: EventInfo{
					Scope:   "project",
					ScopeID: "22",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := e.FireRule(tt.args.ctx, tt.args.content, tt.args.eventInfo); (err != nil) != tt.wantErr {
				t.Errorf("Endpoints.FireRule() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEndpoints_SetConfigInfo(t *testing.T) {
	type args struct {
		eventInfo EventInfo
	}
	var bdl *bundle.Bundle
	p1 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetProject",
		func(d *bundle.Bundle, id uint64) (*apistructs.ProjectDTO, error) {
			return &apistructs.ProjectDTO{
				Name: "project",
			}, nil
		},
	)
	defer p1.Unpatch()

	p2 := monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp",
		func(d *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
			return &apistructs.ApplicationDTO{
				Name: "app",
			}, nil
		},
	)
	defer p2.Unpatch()

	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			args: args{
				eventInfo: EventInfo{
					EventHeader: apistructs.EventHeader{
						ProjectID:     "1",
						ApplicationID: "2",
						Event:         "issue",
					},
				},
			},
			want: map[string]interface{}{
				"project": map[string]interface{}{
					"name": "project",
				},
				"app": map[string]interface{}{
					"name": "app",
				},
			},
		},
	}
	e := &Endpoints{bdl: bdl}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := e.SetConfigInfo(tt.args.eventInfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("Endpoints.SetConfigInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
