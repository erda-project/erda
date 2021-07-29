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
	"time"

	"github.com/erda-project/erda-infra/providers/i18n"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/modules/msp/tenant"
	"github.com/erda-project/erda/modules/msp/tenant/db"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/common/errors"
)

type projectService struct {
	p            *provider
	MSPProjectDB *db.MSPProjectDB
	MSPTenantDB  *db.MSPTenantDB
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

func (s *projectService) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	var projects []*pb.Project
	for _, id := range req.ProjectId {
		projectDB, err := s.MSPProjectDB.Query(id)
		if err != nil {
			return nil, err
		}
		if projectDB == nil {
			continue
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
			rs.DisplayWorkspace = s.getDisplayWorkspace(apis.Language(ctx), rs.Workspace)
			trs = append(trs, &rs)
		}

		project.Relationship = trs
		projects = append(projects, project)
	}
	return &pb.GetProjectsResponse{Data: projects}, nil
}

func (s *projectService) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	if req.Id <= 0 || req.Name == "" || req.Type == "" || req.DisplayName == "" {
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
		mspTenant.CreateTime = time.Now()
		mspTenant.UpdateTime = time.Now()
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
