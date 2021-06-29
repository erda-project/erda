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

package apis

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
)

func Test_checkerV1Service_CreateCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateCheckerV1Response
		wantErr  bool
	}{
		// 		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.CreateCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.CreateCheckerV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.CreateCheckerV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.CreateCheckerV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.CreateCheckerV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_UpdateCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.UpdateCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.UpdateCheckerV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.UpdateCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.UpdateCheckerV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.UpdateCheckerV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.UpdateCheckerV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.UpdateCheckerV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_DeleteCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.DeleteCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.DeleteCheckerV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.DeleteCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.DeleteCheckerV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.DeleteCheckerV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.DeleteCheckerV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.DeleteCheckerV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_DescribeCheckersV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.DescribeCheckersV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.DescribeCheckersV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.DescribeCheckersV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.DescribeCheckersV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.DescribeCheckersV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.DescribeCheckersV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.DescribeCheckersV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_DescribeCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.DescribeCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.DescribeCheckerV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.DescribeCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.DescribeCheckerV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.DescribeCheckerV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.DescribeCheckerV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.DescribeCheckerV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_GetCheckerStatusV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetCheckerStatusV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetCheckerStatusV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetCheckerStatusV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetCheckerStatusV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.GetCheckerStatusV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.GetCheckerStatusV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.GetCheckerStatusV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_checkerV1Service_GetCheckerIssuesV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetCheckerIssuesV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetCheckerIssuesV1Response
		wantErr  bool
	}{
		// TODO: Add test cases.
		// 		{
		// 			"case 1",
		// 			"erda.msp.apm.checker.CheckerV1Service",
		// 			`
		// erda.msp.apm.checker:
		// `,
		// 			args{
		// 				context.TODO(),
		// 				&pb.GetCheckerIssuesV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&pb.GetCheckerIssuesV1Response{
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
			srv := hub.Service(tt.service).(pb.CheckerV1ServiceServer)
			got, err := srv.GetCheckerIssuesV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkerV1Service.GetCheckerIssuesV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("checkerV1Service.GetCheckerIssuesV1() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}
