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

package project

import (
	context "context"
	"net/url"
	"time"

	"github.com/erda-project/erda-infra/providers/i18n"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
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

func (s *projectService) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	var projects []*pb.Project

	// request orch for history project
	params := url.Values{}
	for _, id := range req.ProjectId {
		params.Add("projectId", id)
	}
	userID := apis.GetUserID(ctx)
	orgID := apis.GetOrgID(ctx)
	orchProjects, err := s.p.bdl.GetMSProjects(orgID, userID, params)
	if err != nil {
		return nil, err
	}
	for _, project := range orchProjects {
		pbProject := pb.Project{}
		pbProject.Id = project.ProjectID
		pbProject.Name = project.ProjectName
		pbProject.DisplayName = project.ProjectName
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
		projects = append(projects, &pbProject)
	}

	for _, id := range req.ProjectId {
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
	return &pb.GetProjectsResponse{Data: projects}, nil
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
	for _, mspTenant := range tenants {
		var rs pb.TenantRelationship
		rs.TenantID = mspTenant.Id
		rs.Workspace = mspTenant.RelatedWorkspace
		rs.DisplayWorkspace = s.getDisplayWorkspace(lang, rs.Workspace)
		trs = append(trs, &rs)
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
	return &pb.GetProjectResponse{Data: project}, nil
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
