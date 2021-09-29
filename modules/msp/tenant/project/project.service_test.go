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

package project

import (
	"bou.ke/monkey"
	context "context"
	reflect "reflect"
	testing "testing"

	"github.com/erda-project/erda-infra/providers/i18n"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/common/apis"

	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/tenant/db"
)

func Test_projectService_GetProjects(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetProjectsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetProjectsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetProjectsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetProjectsResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.GetProjects(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.GetProjectList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.GetProjectList() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_GetProjectsTenantsIDs(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetProjectsTenantsIDsRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetProjectsTenantsIDsResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetProjectsTenantsIDsRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetProjectsTenantsIDsResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.GetProjectsTenantsIDs(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.GetProjectsTenantsIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.GetProjectsTenantsIDs() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_CreateProject(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.CreateProjectRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.CreateProjectResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.CreateProjectRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.CreateProjectResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.CreateProject(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.CreateProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.CreateProject() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_UpdateProject(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.UpdateProjectRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.UpdateProjectResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.UpdateProjectRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.UpdateProjectResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.UpdateProject(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.UpdateProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.UpdateProject() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_GetProject(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetProjectRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetProjectResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetProjectRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetProjectResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.GetProject(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.GetProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.GetProject() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_GetProjectOverview(t *testing.T) {
	type args struct {
		ctx context.Context
		req *pb.GetProjectOverviewRequest
	}
	tests := []struct {
		name     string
		service  string
		config   string
		args     args
		wantResp *pb.GetProjectOverviewResponse
		wantErr  bool
	}{
		// TODO: Add test cases.
		//		{
		//			"case 1",
		//			"erda.msp.tenant.project.ProjectService",
		//			`
		//erda.msp.tenant.project:
		//`,
		//			args{
		//				context.TODO(),
		//				&pb.GetProjectOverviewRequest{
		//					// TODO: setup fields
		//				},
		//			},
		//			&pb.GetProjectOverviewResponse{
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
			srv := hub.Service(tt.service).(pb.ProjectServiceServer)
			got, err := srv.GetProjectOverview(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("projectService.GetProjectOverview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("projectService.GetProjectOverview() = %v, want %v", got, tt.wantResp)
			}
		})
	}
}

func Test_projectService_DeleteProject(t *testing.T) {
	init, _, _ := db.MockInit(db.MYSQL)
	projectDB := db.MSPProjectDB{DB: init}
	tenantDB := db.MSPTenantDB{DB: init}
	monitorDB := monitor.MonitorDB{DB: init}

	type fields struct {
		p            *provider
		MSPProjectDB *db.MSPProjectDB
		MSPTenantDB  *db.MSPTenantDB
		MonitorDB    *monitor.MonitorDB
	}
	type args struct {
		ctx context.Context
		req *pb.DeleteProjectRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.DeleteProjectResponse
		wantErr bool
	}{
		{
			name:    "case1",
			fields:  fields{p: nil, MSPProjectDB: &projectDB, MSPTenantDB: &tenantDB, MonitorDB: &monitorDB},
			args:    args{ctx: nil, req: &pb.DeleteProjectRequest{ProjectId: "1"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &projectService{
				p:            tt.fields.p,
				MSPProjectDB: tt.fields.MSPProjectDB,
				MSPTenantDB:  tt.fields.MSPTenantDB,
				MonitorDB:    tt.fields.MonitorDB,
			}
			got, err := s.DeleteProject(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteProject() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_projectService_GetProjectList(t *testing.T) {
	type fields struct {
		p            *provider
		MSPProjectDB *db.MSPProjectDB
		MSPTenantDB  *db.MSPTenantDB
		MonitorDB    *monitor.MonitorDB
	}
	type args struct {
		ctx        context.Context
		projectIDs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Projects
		wantErr bool
	}{
		{"case1", fields{}, args{ctx: nil, projectIDs: []string{"1", "2"}}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s *projectService
			monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetHistoryProjects", func(s *projectService, ctx context.Context, projectIDs []string, projects Projects) ([]apistructs.MicroServiceProjectResponseData, error) {
				var data []apistructs.MicroServiceProjectResponseData
				project := apistructs.MicroServiceProjectResponseData{
					ProjectID:    "1",
					ProjectName:  "test project 1",
					LogoURL:      "http://localhost:8080",
					Envs:         []string{},
					TenantGroups: []string{},
					Workspaces:   map[string]string{},
				}
				project2 := apistructs.MicroServiceProjectResponseData{
					ProjectID:    "2",
					ProjectName:  "test project 2",
					LogoURL:      "http://localhost:8080",
					Envs:         []string{},
					TenantGroups: []string{},
					Workspaces:   map[string]string{},
				}

				data = append(data, project)
				data = append(data, project2)
				return data, nil
			})

			monkey.PatchInstanceMethod(reflect.TypeOf(s), "GetProjectInfo", func(s *projectService, lang i18n.LanguageCodes, id string) (*pb.Project, error) {
				project := &pb.Project{
					Id:           id,
					Name:         "test project " + id,
					Type:         "MSP",
					Relationship: []*pb.TenantRelationship{},
					IsDeleted:    false,
					DisplayName:  "test project " + id,
					DisplayType:  "MSP",
				}
				return project, nil
			})
			monkey.Patch(apis.Language, func(ctx context.Context) i18n.LanguageCodes {
				return nil
			})
			monkey.PatchInstanceMethod(reflect.TypeOf(s), "CovertHistoryProjectToMSPProject", func(s *projectService, ctx context.Context, project apistructs.MicroServiceProjectResponseData) *pb.Project {
				pbProject := pb.Project{}
				pbProject.Id = project.ProjectID
				pbProject.Name = project.ProjectName
				pbProject.DisplayName = project.ProjectName
				pbProject.CreateTime = project.CreateTime.UnixNano()
				pbProject.Type = tenantpb.Type_DOP.String()

				var rss []*pb.TenantRelationship
				for i, env := range project.Envs {
					if env == "" {
						continue
					}
					rs := pb.TenantRelationship{}
					rs.Workspace = env
					rs.TenantID = project.TenantGroups[i]
					rss = append(rss, &rs)
				}
				pbProject.Relationship = rss
				return &pbProject
			})

			got, err := s.GetProjectList(tt.args.ctx, tt.args.projectIDs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProjectList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProjectList() got = %v, want %v", got, tt.want)
			}
		})
	}
}
