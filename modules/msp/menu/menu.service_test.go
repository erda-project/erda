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

package menu

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/menu/pb"
)

func Test_menuService_GetMenu(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetMenuRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetMenuResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.menu.MenuService",
		// 			`
		// erda.msp.menu:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetMenuRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetMenuResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
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
			srv := hub.Service(tt.service).(pb.MenuServiceServer)
			got, err := srv.GetMenu(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("menuService.GetMenu() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("menuService.GetMenu() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_menuService_GetSetting(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetSettingRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetSettingResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.menu.MenuService",
		// 			`
		// erda.msp.menu:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetSettingRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetSettingResponse{
		// 				// TODO: setup fields.
		// 			},
		// 			false,
		// 		},
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
			srv := hub.Service(tt.service).(pb.MenuServiceServer)
			got, err := srv.GetSetting(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("menuService.GetSetting() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("menuService.GetSetting() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
