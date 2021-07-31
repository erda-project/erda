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

package tenant

import (
	context "context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	pb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda/modules/msp/instance/db/monitor"
	"github.com/erda-project/erda/modules/msp/tenant/db"
	"github.com/erda-project/erda/pkg/common/errors"
)

type tenantService struct {
	p           *provider
	MSPTenantDB *db.MSPTenantDB
	MonitorDB   *monitor.MonitorDB
}

func GenerateTenantID(projectID string, tenantType, workspace string) string {
	md5H := md5.New()
	hStr := fmt.Sprintf("%v-%s-%s", projectID, tenantType, workspace)
	md5H.Write([]byte(hStr))
	return hex.EncodeToString(md5H.Sum(nil))
}

func (s *tenantService) GetTenantID(projectID int64, workspace, tenantGroup, tenantType string) (string, error) {
	item, err := s.MonitorDB.GetMonitorByProjectIdAndWorkspace(projectID, workspace)
	if err != nil {
		return "", err
	}
	if item != nil {
		return tenantGroup, nil
	}

	tenant, err := s.MSPTenantDB.QueryTenantByProjectIDAndWorkspace(projectID, workspace)
	if err != nil {
		return "", err
	}
	if tenant == nil {
		return tenant.Id, nil
	}
	return "", nil
}

func (s *tenantService) CreateTenant(ctx context.Context, req *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error) {
	if req.ProjectID == "" {
		return nil, errors.NewMissingParameterError("projectId")
	}
	if req.TenantType == "" {
		return nil, errors.NewMissingParameterError("tenantType")
	}
	if len(req.Workspaces) <= 0 {
		return nil, errors.NewMissingParameterError("workspaces")
	}

	var tenants []*pb.Tenant
	for _, workspace := range req.Workspaces {
		tenantID := GenerateTenantID(req.ProjectID, req.TenantType, workspace)

		queryTenant, err := s.MSPTenantDB.QueryTenant(tenantID)
		if err != nil {
			return nil, err
		}
		if queryTenant != nil {
			tenants = append(tenants, s.covertToTenant(queryTenant))
			continue
		}

		var tenant db.MSPTenant
		tenant.Id = tenantID
		tenant.Type = req.TenantType
		tenant.RelatedWorkspace = workspace
		tenant.RelatedProjectId = req.ProjectID
		tenant.UpdatedAt = time.Now()
		tenant.CreatedAt = time.Now()
		tenant.IsDeleted = false
		result, err := s.MSPTenantDB.InsertTenant(&tenant)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, s.covertToTenant(result))
	}
	return &pb.CreateTenantResponse{Data: tenants}, nil
}

func (s *tenantService) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantResponse, error) {
	if req.ProjectID == "" {
		return nil, errors.NewMissingParameterError("projectId")
	}
	if req.TenantType == "" {
		return nil, errors.NewMissingParameterError("tenantType")
	}
	if req.Workspace == "" {
		return nil, errors.NewMissingParameterError("workspace")
	}
	tenant, err := s.MSPTenantDB.QueryTenant(GenerateTenantID(req.ProjectID, req.TenantType, req.Workspace))
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, errors.NewInternalServerErrorMessage("tenant not exist.")
	}
	return &pb.GetTenantResponse{Data: s.covertToTenant(tenant)}, nil
}

func (s *tenantService) covertToTenant(tenant *db.MSPTenant) *pb.Tenant {
	return &pb.Tenant{
		Id:         tenant.Id,
		Type:       tenant.Type,
		Workspace:  tenant.RelatedWorkspace,
		ProjectID:  tenant.RelatedProjectId,
		CreateTime: tenant.CreatedAt.UnixNano(),
		UpdateTime: tenant.UpdatedAt.UnixNano(),
		IsDeleted:  tenant.IsDeleted,
	}
}
