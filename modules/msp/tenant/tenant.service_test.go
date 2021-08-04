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

package tenant

import (
	context "context"
	reflect "reflect"
	testing "testing"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
)

func Test_tenantService_CreateTenant(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateTenantRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateTenantResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.TenantService",
		//			`
		//erda.msp.tenant:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.CreateTenantRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.CreateTenantResponse{
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
			srv := hub.Service(tt.service).(pb.TenantServiceServer)
			got, err := srv.CreateTenant(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("tenantService.CreateTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("tenantService.CreateTenant() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_tenantService_GetTenant(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetTenantRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetTenantResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.TenantService",
		//			`
		//erda.msp.tenant:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetTenantRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetTenantResponse{
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
			srv := hub.Service(tt.service).(pb.TenantServiceServer)
			got, err := srv.GetTenant(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("tenantService.GetTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("tenantService.GetTenant() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_generateTenantID(t *testing.T) {
	type args struct {
		projectID  string
		tenantType string
		workspace  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"case1", args{"1", "msp", "DEFAULT"}, GenerateTenantID("1", "msp", "DEFAULT")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateTenantID(tt.args.projectID, tt.args.tenantType, tt.args.workspace); got != tt.want {
				t.Errorf("GenerateTenantID() = %v, want %v", got, tt.want)
			}
		})
	}
}
