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

package settings

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/core/monitor/settings/pb"
)

func Test_settingsService_GetSettings(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetSettingsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetSettingsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.settings.SettingsService",
		// 			`
		// erda.core.monitor.settings:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetSettingsRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetSettingsResponse{
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
			srv := hub.Service(tt.service).(pb.SettingsServiceServer)
			got, err := srv.GetSettings(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("settingsService.GetSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("settingsService.GetSettings() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_settingsService_PutSettings(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.PutSettingsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.PutSettingsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.settings.SettingsService",
		// 			`
		// erda.core.monitor.settings:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.PutSettingsRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.PutSettingsResponse{
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
			srv := hub.Service(tt.service).(pb.SettingsServiceServer)
			got, err := srv.PutSettings(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("settingsService.PutSettings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("settingsService.PutSettings() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_settingsService_RegisterMonitorConfig(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.RegisterMonitorConfigRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.RegisterMonitorConfigResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.core.monitor.settings.SettingsService",
		// 			`
		// erda.core.monitor.settings:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.RegisterMonitorConfigRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.RegisterMonitorConfigResponse{
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
			srv := hub.Service(tt.service).(pb.SettingsServiceServer)
			got, err := srv.RegisterMonitorConfig(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("settingsService.RegisterMonitorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("settingsService.RegisterMonitorConfig() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
