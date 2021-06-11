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

package configcenter

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/configcenter/pb"
)

func Test_configCenterService_GetGroups(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetGroupRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetGroupResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.configcenter.ConfigCenterService",
		// 			`
		// erda.msp.configcenter:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetGroupRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetGroupResponse{
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
			srv := hub.Service(tt.service).(pb.ConfigCenterServiceServer)
			got, err := srv.GetGroups(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("configCenterService.GetGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("configCenterService.GetGroups() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_configCenterService_GetGroupProperties(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetGroupPropertiesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetGroupPropertiesResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.configcenter.ConfigCenterService",
		// 			`
		// erda.msp.configcenter:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetGroupPropertiesRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetGroupPropertiesResponse{
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
			srv := hub.Service(tt.service).(pb.ConfigCenterServiceServer)
			got, err := srv.GetGroupProperties(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("configCenterService.GetGroupProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("configCenterService.GetGroupProperties() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_configCenterService_SaveGroupProperties(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.SaveGroupPropertiesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.SaveGroupPropertiesResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.configcenter.ConfigCenterService",
		// 			`
		// erda.msp.configcenter:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.SaveGroupPropertiesRequest{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.SaveGroupPropertiesResponse{
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
			srv := hub.Service(tt.service).(pb.ConfigCenterServiceServer)
			got, err := srv.SaveGroupProperties(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("configCenterService.SaveGroupProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("configCenterService.SaveGroupProperties() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
