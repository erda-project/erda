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
	context "context"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/erda-project/erda-infra/providers/i18n"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/tenant"
	"github.com/erda-project/erda/modules/msp/tenant/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type projectService struct {
	p            *provider
	MSPProjectDB *db.MSPProjectDB
	MSPTenantDB  *db.MSPTenantDB
	MonitorDB    *monitor.MonitorDB
}

func (s *projectService) getDisplayWorkspace(lang i18n.LanguageCodes, workspace string) string {
	if lang == nil {
		return workspace
	}
	switch workspace {
	case tenantpb.Workspace_DEV.String():
		return s.p.I18n.Text(lang, "workspace_dev")
	case tenantpb.Workspace_TEST.String():
		return s.p.I18n.Text(lang, "workspace_test")
	case tenantpb.Workspace_STAGING.String():
		return s.p.I18n.Text(lang, "workspace_staging")
	case tenantpb.Workspace_PROD.String():
		return s.p.I18n.Text(lang, "workspace_prod")
	case tenantpb.Workspace_DEFAULT.String():
		return s.p.I18n.Text(lang, "workspace_default")
	default:
		return ""
	}
}

func (s *projectService) getDisplayType(lang i18n.LanguageCodes, projectType string) string {
	if lang == nil {
		return projectType
	}
	switch projectType {
	case tenantpb.Type_DOP.String():
		return s.p.I18n.Text(lang, "project_type_dop")
	case tenantpb.Type_MSP.String():
		return s.p.I18n.Text(lang, "project_type_msp")
	default:
		return ""
	}
}

type Projects []*pb.Project

func (p Projects) Len() int { return len(p) }

func (p Projects) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p Projects) Less(i, j int) bool { return p[i].CreateTime > p[j].CreateTime }

func (s *projectService) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	projects, err := s.getProjects(ctx, req.ProjectId)
	if err != nil {
		return &pb.GetProjectsResponse{Data: nil}, err
	}
	sort.Sort(projects)
	return &pb.GetProjectsResponse{Data: projects}, nil
}

var workspaces = []string{
	tenantpb.Workspace_DEV.String(),
	tenantpb.Workspace_TEST.String(),
	tenantpb.Workspace_STAGING.String(),
	tenantpb.Workspace_PROD.String(),
	tenantpb.Workspace_DEFAULT.String(),
}

func (s *projectService) getProjects(ctx context.Context, projectIDs []string) (Projects, error) {
	var projects Projects

	// request orch for history project
	orchProjects, err := s.getHistoryProjects(ctx, projectIDs, projects)
	if err != nil {
		return nil, err
	}
	for _, project := range orchProjects {
		pbProject := s.covertHistoryProjectToMSPProject(ctx, project)
		projects = append(projects, pbProject)
	}

	for _, id := range projectIDs {
		project, err := s.getProject(apis.Language(ctx), id)
		if err != nil {
			return nil, err
		}
		if project == nil {
			continue
		}

		for i, p := range projects {
			if p.Id == id {
				projects = append(projects[:i], projects[i+1:]...)
				break
			}
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (s *projectService) getHistoryProjects(ctx context.Context, projectIDs []string, projects Projects) ([]apistructs.MicroServiceProjectResponseData, error) {
	params := url.Values{}
	for _, id := range projectIDs {
		params.Add("projectId", id)
	}
	userID := apis.GetUserID(ctx)
	orgID := apis.GetOrgID(ctx)
	orchProjects, err := s.p.bdl.GetMSProjects(orgID, userID, params)
	if err != nil {
		return nil, err
	}
	return orchProjects, err
}

func (s *projectService) covertHistoryProjectToMSPProject(ctx context.Context, project apistructs.MicroServiceProjectResponseData) *pb.Project {
	pbProject := pb.Project{}
	pbProject.Id = project.ProjectID
	pbProject.Name = project.ProjectName
	pbProject.DisplayName = project.ProjectName
	pbProject.CreateTime = project.CreateTime.UnixNano()
	pbProject.Type = tenantpb.Type_DOP.String()
	pbProject.DisplayType = s.getDisplayType(apis.Language(ctx), tenantpb.Type_DOP.String())

	var rss []*pb.TenantRelationship
	for i, env := range project.Envs {
		if env == "" {
			continue
		}
		rs := pb.TenantRelationship{}
		rs.Workspace = env
		rs.DisplayWorkspace = s.getDisplayWorkspace(apis.Language(ctx), env)
		rs.TenantID = project.TenantGroups[i]
		rss = append(rss, &rs)
	}
	pbProject.Relationship = rss
	return &pbProject
}

func (s *projectService) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	if req.Id == "" || req.Name == "" || req.Type == "" || req.DisplayName == "" {
		return nil, errors.NewMissingParameterError("id,name,displayName or type missing.")
	}
	var project db.MSPProject
	project.Id = req.Id
	project.Name = req.Name
	project.Type = req.Type
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	project.DisplayName = req.DisplayName
	project.IsDeleted = false

	// create msp project
	create, err := s.MSPProjectDB.Create(&project)
	if err != nil {
		return nil, err
	}

	mspTenant := db.MSPTenant{}
	if project.Type == tenantpb.Type_MSP.String() {
		mspTenant.Id = tenant.GenerateTenantID(project.Id, project.Type, tenantpb.Workspace_DEFAULT.String())
		mspTenant.CreatedAt = time.Now()
		mspTenant.UpdatedAt = time.Now()
		mspTenant.IsDeleted = false
		mspTenant.Type = tenantpb.Type_MSP.String()
		mspTenant.RelatedProjectId = project.Id
		mspTenant.RelatedWorkspace = tenantpb.Workspace_DEFAULT.String()
		insertTenant, err := s.MSPTenantDB.InsertTenant(&mspTenant)
		if err != nil {
			s.p.Log.Errorf("create msp tenant (msp project) failed.", err)
			return nil, err
		}
		mspTenant = *insertTenant
	}

	pbProject := s.convertToProject(create)
	if project.Type == tenantpb.Type_MSP.String() {
		var relationships []*pb.TenantRelationship
		rs := pb.TenantRelationship{}
		rs.Workspace = mspTenant.RelatedWorkspace
		rs.TenantID = mspTenant.Id
		rs.DisplayWorkspace = s.getDisplayWorkspace(apis.Language(ctx), rs.Workspace)
		relationships = append(relationships, &rs)
		pbProject.Relationship = relationships
	}
	return &pb.CreateProjectResponse{Data: pbProject}, nil
}

func (s *projectService) UpdateProject(ctx context.Context, req *pb.UpdateProjectRequest) (*pb.UpdateProjectResponse, error) {
	project, err := s.MSPProjectDB.Query(req.Id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return &pb.UpdateProjectResponse{Data: nil}, nil
	}
	project.UpdatedAt = time.Now()
	project.Name = req.Name
	project.DisplayName = req.DisplayName
	project, err = s.MSPProjectDB.Update(project)
	pbProject := s.convertToProject(project)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateProjectResponse{Data: pbProject}, nil
}

func (s *projectService) getProject(lang i18n.LanguageCodes, id string) (*pb.Project, error) {
	projectDB, err := s.MSPProjectDB.Query(id)
	if err != nil {
		return nil, err
	}
	if projectDB == nil {
		return nil, nil
	}
	project := s.convertToProject(projectDB)
	tenants, err := s.MSPTenantDB.QueryTenantByProjectID(id)
	if err != nil {
		return nil, err
	}
	var trs []*pb.TenantRelationship
	for _, workspace := range workspaces {
		for _, mspTenant := range tenants {
			if workspace == mspTenant.RelatedWorkspace {
				var rs pb.TenantRelationship
				rs.TenantID = mspTenant.Id
				rs.Workspace = mspTenant.RelatedWorkspace
				rs.DisplayWorkspace = s.getDisplayWorkspace(lang, rs.Workspace)
				trs = append(trs, &rs)
				break
			}
		}
	}

	project.Relationship = trs
	project.DisplayType = s.getDisplayType(lang, project.Type)
	return project, nil
}

func (s *projectService) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	project, err := s.getProject(apis.Language(ctx), req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project == nil {
		// request orch for history project
		params := url.Values{}
		params.Add("projectId", req.ProjectID)
		userID := apis.GetUserID(ctx)
		orgID := apis.GetOrgID(ctx)
		orchProjects, err := s.p.bdl.GetMSProjects(orgID, userID, params)
		if err != nil {
			return nil, err
		}
		historyMicroserviceProject := orchProjects[0]
		mspProject := s.covertHistoryProjectToMSPProject(ctx, historyMicroserviceProject)
		project = mspProject
	}

	return &pb.GetProjectResponse{Data: project}, nil
}

func (s *projectService) DeleteProject(ctx context.Context, req *pb.DeleteProjectRequest) (*pb.DeleteProjectResponse, error) {
	_, err := s.MSPProjectDB.Delete(req.ProjectId)
	if err != nil {
		return nil, err
	}
	tenants, err := s.MSPTenantDB.QueryTenantByProjectID(req.ProjectId)
	if err != nil {
		return nil, err
	}
	for _, mspTenant := range tenants {
		_, err := s.MSPTenantDB.DeleteTenantByTenantID(tenant.GenerateTenantID(mspTenant.RelatedProjectId, mspTenant.Type, mspTenant.RelatedWorkspace))
		if err != nil {
			return nil, err
		}
	}
	return &pb.DeleteProjectResponse{Data: nil}, nil
}

func (s *projectService) GetProjectOverview(ctx context.Context, req *pb.GetProjectOverviewRequest) (*pb.GetProjectOverviewResponse, error) {
	projects, err := s.getProjects(ctx, req.ProjectId)
	predata := pb.ProjectOverviewList{}
	var data []*pb.ProjectOverview
	pv := pb.ProjectOverview{}

	if err != nil {
		return &pb.GetProjectOverviewResponse{Data: nil}, err
	}
	projectCount := int64(projects.Len())
	workspaceCount := int64(0)
	for _, project := range projects {
		workspaceCount += int64(len(project.Relationship))
	}
	pv.ProjectCount = projectCount
	pv.WorkspaceCount = workspaceCount
	data = append(data, &pv)
	predata.Data = data
	return &pb.GetProjectOverviewResponse{Data: &predata}, nil
}

func (s *projectService) GetProjectsTenantsIDs(ctx context.Context, req *pb.GetProjectsTenantsIDsRequest) (*pb.GetProjectsTenantsIDsResponse, error) {
	var ids []string
	idmap := map[string]string{}
	for _, idstr := range req.ProjectId {
		id, err := strconv.ParseInt(idstr, 10, 64)
		if err != nil {
			return nil, err
		}
		monitors, err := s.MonitorDB.GetMonitorByProjectId(id)
		for _, m := range monitors {
			ids = append(ids, m.TerminusKey)
			idmap[m.TerminusKey] = m.TerminusKey
		}
	}

	for _, id := range req.ProjectId {
		tenants, err := s.MSPTenantDB.QueryTenantByProjectID(id)
		if err != nil {
			return nil, err
		}

		for _, mspTenant := range tenants {
			if idmap[mspTenant.Id] != mspTenant.Id {
				ids = append(ids, mspTenant.Id)
			}
		}
	}

	return &pb.GetProjectsTenantsIDsResponse{Data: ids}, nil
}

func (s *projectService) convertToProject(project *db.MSPProject) *pb.Project {
	return &pb.Project{
		Id:          project.Id,
		Name:        project.Name,
		DisplayName: project.DisplayName,
		Type:        project.Type,
		CreateTime:  project.CreatedAt.UnixNano(),
		UpdateTime:  project.UpdatedAt.UnixNano(),
		IsDeleted:   project.IsDeleted,
	}
}
