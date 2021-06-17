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

package registercenter

import (
	"context"
	"strings"

	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
	"github.com/erda-project/erda/modules/msp/registercenter/zkproxy"
	"github.com/erda-project/erda/pkg/common/errors"
)

// GetHostRuntimeRule depracated
func (s *registerCenterService) GetHostRuntimeRule(ctx context.Context, req *pb.GetHostRuntimeRuleRequest) (*pb.GetHostRuntimeRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.GetHostRuntimeRule(req.ProjectID, req.Env, req.Host, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.GetHostRuntimeRule", err)
	}
	return &pb.GetHostRuntimeRuleResponse{Data: rule}, nil
}

// CreateHostRuntimeRule depracated
func (s *registerCenterService) CreateHostRuntimeRule(ctx context.Context, req *pb.CreateHostRuntimeRuleRequest) (*pb.CreateHostRuntimeRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.CreateHostRuntimeRule(req.ProjectID, req.Env, req.Host, namespace, req.Rules)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.CreateHostRuntimeRule", err)
	}
	return &pb.CreateHostRuntimeRuleResponse{Data: rule}, nil
}

// GetAllHostRuntimeRules depracated
func (s *registerCenterService) GetAllHostRuntimeRules(ctx context.Context, req *pb.GetAllHostRuntimeRulesRequest) (*pb.GetAllHostRuntimeRulesResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rules, err := adp.GetAllHostRuntimeRules(req.ProjectID, req.Env, req.AppID, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.GetAllHostRuntimeRules", err)
	}
	return &pb.GetAllHostRuntimeRulesResponse{Data: rules}, nil
}
