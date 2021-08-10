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

func Test_releaseService_CreateRelease(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseCreateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseCreateResponseData
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseCreateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseCreateResponseData{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.CreateRelease(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.CreateRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.CreateRelease() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_UpdateRelease(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseUpdateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseDataResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseUpdateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseDataResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.UpdateRelease(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.UpdateRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.UpdateRelease() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_UpdateReleaseReference(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseReferenceUpdateRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseDataResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseReferenceUpdateRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseDataResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.UpdateReleaseReference(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.UpdateReleaseReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.UpdateReleaseReference() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_GetIosPlist(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetIosPlistRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetIosPlistResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetIosPlistRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetIosPlistResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.GetIosPlist(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.GetIosPlist() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.GetIosPlist() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_GetRelease(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetIosPlistRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseGetResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetIosPlistRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseGetResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.GetRelease(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.GetRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.GetRelease() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_DeleteRelease(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetIosPlistRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseDataResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetIosPlistRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseDataResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.DeleteRelease(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.DeleteRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.DeleteRelease() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_ListRelease(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseListRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseUserDataResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseListRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseUserDataResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.ListRelease(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.ListRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.ListRelease() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_ListReleaseName(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ListReleaseNameRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ListReleaseNameResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ListReleaseNameRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ListReleaseNameResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.ListReleaseName(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.ListReleaseName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.ListReleaseName() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_GetLatestReleases(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetLatestReleasesRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetLatestReleasesResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetLatestReleasesRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetLatestReleasesResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.GetLatestReleases(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.GetLatestReleases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.GetLatestReleases() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_releaseService_ReleaseGC(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.ReleaseGCRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.ReleaseDataResponse
		wantErr  bool
	}{
		//		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.core.dicehub.release.ReleaseService",
		//			`
		//erda.core.dicehub.release:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.ReleaseGCRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.ReleaseDataResponse{
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
			srv := hub.Service(tt.service).(pb.ReleaseServiceServer)
			got, err := srv.ReleaseGC(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("releaseService.ReleaseGC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("releaseService.ReleaseGC() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
