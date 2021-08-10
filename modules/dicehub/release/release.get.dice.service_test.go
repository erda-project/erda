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

package release

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
)

func Test_releaseGetDiceService_PullDiceYAML(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseGetDiceYmlRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseGetDiceYmlResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseGetDiceService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseGetDiceYmlRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseGetDiceYmlResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseGetDiceServiceServer)
			got, err := srv.PullDiceYAML(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseGetDiceService.PullDiceYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseGetDiceService.PullDiceYAML() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseGetDiceService_GetDiceYAML(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseGetDiceYmlRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseGetDiceYmlResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseGetDiceService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseGetDiceYmlRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseGetDiceYmlResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseGetDiceServiceServer)
			got, err := srv.GetDiceYAML(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseGetDiceService.GetDiceYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseGetDiceService.GetDiceYAML() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
