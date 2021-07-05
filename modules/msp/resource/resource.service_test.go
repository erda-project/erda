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

package resource

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/resource/pb"
	"reflect"
	"testing"
)

func Test_resourceService_CreateResource(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateResourceRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateResourceResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			"case 1",
			"erda.msp.resource.ResourceService",
			`
erda.msp.resource:
`,
			args{
				context.TODO(),
				&pb.CreateResourceRequest{
					// TODO: setup fields
				},
			},
			&pb.CreateResourceResponse{
				// TODO: setup fields.
			},
			false,
		},
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
			srv := hub.Service(tt.service).(pb.ResourceServiceServer)
			got, err := srv.CreateResource(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("resourceService.CreateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("resourceService.CreateResource() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
