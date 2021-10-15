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

package apis

import (
	"context"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/servicehub"
	checkerpb "github.com/erda-project/erda-proto-go/msp/apm/checker/pb"
	"github.com/erda-project/erda/modules/msp/apm/checker/storage/db"
)

func Test_checkerV1Service_CreateCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *checkerpb.CreateCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.CreateCheckerV1Response
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
		// 				&checkerpb.CreateCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.CreateCheckerV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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
		req *checkerpb.UpdateCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.UpdateCheckerV1Response
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
		// 				&checkerpb.UpdateCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.UpdateCheckerV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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
		req *checkerpb.DeleteCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.DeleteCheckerV1Response
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
		// 				&checkerpb.DeleteCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.DeleteCheckerV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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

func Test_checkerV1Service_DescribeCheckerV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *checkerpb.DescribeCheckerV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.DescribeCheckerV1Response
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
		// 				&checkerpb.DescribeCheckerV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.DescribeCheckerV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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
		req *checkerpb.GetCheckerStatusV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.GetCheckerStatusV1Response
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
		// 				&checkerpb.GetCheckerStatusV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.GetCheckerStatusV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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
		req *checkerpb.GetCheckerIssuesV1Request
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *checkerpb.GetCheckerIssuesV1Response
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
		// 				&checkerpb.GetCheckerIssuesV1Request{
		// 					// TODO: setup fields
		// 				},
		// 			},
		// 			&checkerpb.GetCheckerIssuesV1Response{
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
			srv := hub.Service(tt.service).(checkerpb.CheckerV1ServiceServer)
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

func Test_checkerV1Service_DescribeCheckersV1(t *testing.T) {
	type args struct {
		ctx context.Context
		req *checkerpb.DescribeCheckersV1Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"case1", args{ctx: nil, req: &checkerpb.DescribeCheckersV1Request{ProjectID: -1, Env: "TEST"}}, true},
		{"case2", args{ctx: nil, req: &checkerpb.DescribeCheckersV1Request{ProjectID: 0, Env: "TEST"}}, false},
		{"case3", args{ctx: nil, req: &checkerpb.DescribeCheckersV1Request{ProjectID: 1, Env: "TEST"}}, false},
		{"case4", args{ctx: nil, req: &checkerpb.DescribeCheckersV1Request{ProjectID: -2, Env: "TEST"}}, true},
		{"case5", args{ctx: nil, req: &checkerpb.DescribeCheckersV1Request{ProjectID: 2, Env: "TEST"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var prodb *db.ProjectDB
			monkey.PatchInstanceMethod(reflect.TypeOf(prodb), "GetByProjectID", func(prodb *db.ProjectDB, projectID int64) (*db.Project, error) {
				if projectID == -1 {
					return nil, errors.New("no project")
				}
				if projectID == 0 || projectID == 2 {
					return nil, nil
				}
				return &db.Project{ID: projectID}, nil
			})
			//// ListByProjectIDAndEnv
			var metricdb *db.MetricDB
			monkey.PatchInstanceMethod(reflect.TypeOf(metricdb), "ListByProjectIDAndEnv", func(metricdb *db.MetricDB, projectID int64, env string) ([]*db.Metric, error) {
				var ms []*db.Metric
				if projectID == -1 {
					return nil, errors.New("no project")
				}
				if projectID == -2 {
					return nil, errors.New("no project")
				}
				if projectID == -3 && env == "ERROR" {
					return nil, errors.New("no project")
				}
				for i := 0; i < 2; i++ {
					ms = append(ms, &db.Metric{
						ID:         0,
						ProjectID:  0,
						ServiceID:  0,
						Name:       "",
						URL:        "",
						Mode:       "HTTP",
						Status:     0,
						Env:        "TEST",
						Config:     "",
						CreateTime: time.Now(),
						UpdateTime: time.Now(),
						IsDeleted:  "N",
					})
				}
				return ms, nil
			})
			var cv1s *checkerV1Service
			monkey.PatchInstanceMethod(reflect.TypeOf(cv1s), "QueryCheckersLatencySummaryByProject", func(cv1s *checkerV1Service, projectID int64, metrics map[int64]*checkerpb.DescribeItemV1) error {
				if projectID == -2 || projectID == 2 {
					return errors.New("no project")
				}
				return nil
			})

			s := &checkerV1Service{}
			_, err := s.DescribeCheckersV1(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DescribeCheckersV1() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
