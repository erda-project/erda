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

package example

import (
	context "context"
	reflect "reflect"
	testing "testing"

	pb "github.com/erda-project/erda-proto-go/examples/pb"
)

func Test_userService_GetUser(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetUserRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp *pb.GetUserResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &userService{
				p: tt.fields.p,
			}
			gotResp, err := s.GetUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userService.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("userService.GetUser() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}

func Test_userService_UpdateUser(t *testing.T) {
	type fields struct {
		p *provider
	}
	type args struct {
		ctx context.Context
		req *pb.GetUserRequest
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp *pb.UpdateUserResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &userService{
				p: tt.fields.p,
			}
			gotResp, err := s.UpdateUser(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("userService.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("userService.UpdateUser() = %v, want %v", gotResp, tt.wantResp)
			}
		})
	}
}
