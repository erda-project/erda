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

package impl

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-proto-go/core/hepa/hub_info/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	context1 "github.com/erda-project/erda/internal/tools/orchestrator/hepa/context"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/hepautils"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	repositoryService "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

func NewHubInfoService() (pb.HubInfoServiceServer, error) {
	hubInfoService, err := repositoryService.NewGatewayHubInfoServiceImpl()
	if err != nil {
		return nil, err
	}
	return &hubInfoHandler{hubInfoService: hubInfoService}, nil
}

type hubInfoHandler struct {
	hubInfoService repositoryService.GatewayHubInfoService
}

func (h hubInfoHandler) CreateOrGetHubInfo(ctx context.Context, req *pb.CreateHubInfoReq) (*pb.GetHubInfoResp, error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	if len(req.GetDomains()) == 0 {
		return nil, erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "no domain")
	}
	hubInfos, err := h.hubInfoService.SelectByAny(&orm.GatewayHubInfo{
		OrgId:           req.GetOrgID(),
		DiceEnv:         req.GetEnv(),
		DiceClusterName: req.GetAz(),
	})
	if err != nil {
		return nil, err
	}
	if hubInfo, ok := hubExists(hubInfos, req.GetDomains()); ok {
		return &pb.GetHubInfoResp{
			Success: true,
			Data: &pb.GetHubInfoItem{
				Id:      hubInfo.Id,
				OrgID:   hubInfo.OrgId,
				Env:     hubInfo.DiceEnv,
				Az:      hubInfo.DiceClusterName,
				Domains: strings.Split(hubInfo.BindDomain, ","),
			},
		}, nil
	}

	var hubInfo = &orm.GatewayHubInfo{
		OrgId:           req.GetOrgID(),
		DiceEnv:         req.GetEnv(),
		DiceClusterName: req.GetAz(),
		BindDomain:      hepautils.SortJoinDomains(req.GetDomains()),
	}
	if err = h.hubInfoService.Insert(hubInfo); err != nil {
		return nil, err
	}
	return &pb.GetHubInfoResp{
		Success: true,
		Data: &pb.GetHubInfoItem{
			Id:        hubInfo.Id,
			OrgID:     hubInfo.OrgId,
			ProjectID: "",
			Env:       hubInfo.DiceEnv,
			Az:        hubInfo.DiceClusterName,
			Domains:   strings.Split(hubInfo.BindDomain, ","),
		},
	}, nil
}

func (h hubInfoHandler) GetHubInfo(ctx context.Context, req *pb.GetHubInfoReq) (*pb.GetHubInfoResp, error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	hubInfos, err := h.hubInfoService.SelectByAny(&orm.GatewayHubInfo{
		OrgId:           req.GetOrgID(),
		DiceEnv:         req.GetEnv(),
		DiceClusterName: req.GetAz(),
	})
	if err != nil {
		return nil, err
	}
	if len(hubInfos) == 0 {
		return &pb.GetHubInfoResp{Success: false}, nil
	}
	for _, hubInfo := range hubInfos {
		domains := strings.Split(hubInfo.BindDomain, ",")
		for _, domain := range domains {
			if req.GetOneOfDomains() == domain {
				return &pb.GetHubInfoResp{
					Success: true,
					Data: &pb.GetHubInfoItem{
						Id:        hubInfo.Id,
						OrgID:     hubInfo.OrgId,
						ProjectID: "",
						Env:       hubInfo.DiceEnv,
						Az:        hubInfo.DiceClusterName,
						Domains:   domains,
					},
				}, nil
			}
		}
	}

	return &pb.GetHubInfoResp{Success: false}, nil
}

func hubExists(hubInfos []orm.GatewayHubInfo, bindDomain []string) (*orm.GatewayHubInfo, bool) {
	if len(hubInfos) == 0 {
		return nil, false
	}
	for _, bindDomain := range bindDomain {
		for _, hubInfo := range hubInfos {
			hubInfoDomains := strings.Split(hubInfo.BindDomain, ",")
			for _, hubDomain := range hubInfoDomains {
				if bindDomain == hubDomain {
					return &hubInfo, true
				}
			}
		}
	}
	return nil, false
}
