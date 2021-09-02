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

package extension

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/dicehub/extension/pb"
)

func Test_extensionService_SearchExtensions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionSearchRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionSearchResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionSearchRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionSearchResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.SearchExtensions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.SearchExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.SearchExtensions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_CreateExtension(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionCreateResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionCreateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionCreateResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.CreateExtension(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.CreateExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.CreateExtension() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryExtensionsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryExtensionsResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.QueryExtensionsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.QueryExtensionsResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensionsMenu(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.QueryExtensionsMenuRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.QueryExtensionsMenuResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.QueryExtensionsMenuRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.QueryExtensionsMenuResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensionsMenu(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensionsMenu() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensionsMenu() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_CreateExtensionVersion222(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionVersionCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionVersionCreateResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionVersionCreateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionVersionCreateResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.CreateExtensionVersion(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.CreateExtensionVersion222() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.CreateExtensionVersion222() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_GetExtensionVersion(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetExtensionVersionRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetExtensionVersionResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetExtensionVersionRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetExtensionVersionResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.GetExtensionVersion(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.GetExtensionVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.GetExtensionVersion() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_extensionService_QueryExtensionVersions(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ExtensionVersionQueryRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ExtensionVersionQueryResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.extension.ExtensionService",
		//			`
		//erda.core.dicehub.extension:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ExtensionVersionQueryRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ExtensionVersionQueryResponse{
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
			srv := hub.Service(tt.service).(pb.ExtensionServiceServer)
			got, err := srv.QueryExtensionVersions(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("extensionService.QueryExtensionVersions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("extensionService.QueryExtensionVersions() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
