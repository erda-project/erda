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

// GetHostRule depracated
func (s *registerCenterService) GetHostRule(ctx context.Context, req *pb.CetHostRuleRequest) (*pb.CetHostRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.GetHostRule(req.ProjectID, req.Env, req.AppID, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.GetHostRule", err)
	}
	return &pb.CetHostRuleResponse{
		Data: rule,
	}, nil
}

// CreateHostRule depracated
func (s *registerCenterService) CreateHostRule(ctx context.Context, req *pb.CreateHostRuleRequest) (*pb.CreateHostRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.CreateHostRoute(req.ProjectID, req.Env, req.AppID, namespace, req.Rules)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.CreateHostRoute", err)
	}
	return &pb.CreateHostRuleResponse{
		Data: rule,
	}, nil
}

// DeleteHostRule depracated
func (s *registerCenterService) DeleteHostRule(ctx context.Context, req *pb.DeleteHostRuleRequest) (*pb.DeleteHostRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.DeleteHostRoute(req.ProjectID, req.Env, req.AppID, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.DeleteHostRoute", err)
	}
	return &pb.DeleteHostRuleResponse{
		Data: rule,
	}, nil
}
