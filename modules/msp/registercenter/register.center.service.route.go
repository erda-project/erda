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

// GetRouteRule depracated
func (s *registerCenterService) GetRouteRule(ctx context.Context, req *pb.GetRouteRuleRequest) (*pb.GetRouteRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.GetRouteRule(req.InterfaceName, req.ProjectID, req.Env, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.GetRouteRule", err)
	}
	return &pb.GetRouteRuleResponse{
		Data: rule,
	}, nil
}

// CreateRouteRule depracated
func (s *registerCenterService) CreateRouteRule(ctx context.Context, req *pb.CreateRouteRuleRequest) (*pb.CreateRouteRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.CreateRouteRule(req.InterfaceName, req.ProjectID, req.Env, namespace, req.Rule)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.CreateRouteRule", err)
	}
	return &pb.CreateRouteRuleResponse{Data: rule}, nil
}

// DeleteRouteRule depracated
func (s *registerCenterService) DeleteRouteRule(ctx context.Context, req *pb.DeleteRouteRuleRequest) (*pb.DeleteRouteRuleResponse, error) {
	namespace := req.TenantID
	if len(namespace) <= 0 {
		namespace = req.ProjectID + "_" + strings.ToLower(req.Env)
	}
	host, err := s.getzkProxyHost(req.ClusterName)
	if err != nil {
		return nil, errors.NewDataBaseError(err)
	}
	adp := zkproxy.NewAdapter(req.ClusterName, host)
	rule, err := adp.DeleteRouteRule(req.InterfaceName, req.ProjectID, req.Env, namespace)
	if err != nil {
		return nil, errors.NewServiceInvokingError("zkproxy.DeleteRouteRule", err)
	}
	return &pb.DeleteRouteRuleResponse{Data: rule}, nil
}
