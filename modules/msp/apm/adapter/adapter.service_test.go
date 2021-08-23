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

package adapter

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/apm/adapter/pb"
)

func Test_adapterService_GetAdapters(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetAdaptersRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetAdaptersResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.apm.adapter.AdapterService",
		//			`
		//erda.msp.apm.adapter:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetAdapterRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetAdapterResponse{
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
			srv := hub.Service(tt.service).(pb.AdapterServiceServer)
			got, err := srv.GetAdapters(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("adapterService.GetAdapters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("adapterService.GetAdapters() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
