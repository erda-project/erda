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

package legacy_upstream

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	pb2 "github.com/erda-project/erda-proto-go/core/hepa/hub_info/pb"
	"github.com/erda-project/erda-proto-go/core/hepa/legacy_upstream/pb"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common/vars"
	context1 "github.com/erda-project/erda/internal/tools/orchestrator/hepa/context"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/hub_info"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/legacy_upstream"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

type upstreamService struct {
	hubInfoService hub_info.Interface
}

func (s *upstreamService) Register(ctx context.Context, req *pb.RegisterRequest) (resp *pb.RegisterResponse, err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())

	if req.GetUpstream() == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "invalid request")
		return
	}
	logrus.Infof("Call /api/gateway/register_async with Req: %+v\n", *(req.GetUpstream()))
	if req.GetUpstream().GetRuntimeId() == "" || req.GetUpstream().GetOnlyRuntimePath() {
		return s.register(ctx, req)
	}
	if _, err = s.register(ctx, req); err != nil {
		return nil, err
	}
	req.Upstream.RuntimeId = ""
	return s.register(ctx, req)
}

func (s *upstreamService) register(ctx context.Context, req *pb.RegisterRequest) (resp *pb.RegisterResponse, err error) {
	service := legacy_upstream.Service.Clone(ctx)

	// convert struct and adjust it
	reqDto := dto.FromUpstream(req.Upstream)
	if req.GetUpstream().GetScene() == orm.HubScene {
		if err := s.patchHubInfo(ctx, reqDto); err != nil {
			return nil, err
		}
	}
	if err = reqDto.Init(); err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, err.Error())
		return
	}

	// adjust every api struct
	for i := 0; i < len(reqDto.ApiList); i++ {
		if err = (&reqDto.ApiList[i]).Init(); err != nil {
			err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, err.Error())
			return
		}
	}

	// do upstream register
	result, err := service.UpstreamRegister(ctx, reqDto)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.RegisterResponse{
		Data: result,
	}
	return
}

// AsyncRegister PUT /api/gateway/register_async
func (s *upstreamService) AsyncRegister(ctx context.Context, req *pb.AsyncRegisterRequest) (resp *pb.AsyncRegisterResponse, err error) {
	ctx = context1.WithLoggerIfWithout(ctx, logrus.StandardLogger())
	l := ctx.(*context1.LogContext).Entry()
	l.WithError(nil).Infof("Call /api/gateway/register_async with Req: %+v\n", *(req.GetUpstream()))

	if req.Upstream == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "invalid request")
		return
	}
	if req.GetUpstream().GetRuntimeId() == "" || req.GetUpstream().GetOnlyRuntimePath() {
		return s.asyncRegister(ctx, req)
	}
	go func() {
		l := ctx.(*context1.LogContext).Entry()
		response, err := s.Register(ctx, &pb.RegisterRequest{Upstream: req.GetUpstream()})
		if err != nil {
			l.WithError(err).WithField("registerTag", req.GetUpstream().GetRegisterTag()).Errorln("failed to go s.Register")
			return
		}
		if response == nil || !response.GetData() {
			l.WithField("registerTag", req.GetUpstream().GetRegisterTag()).Errorln("failed to go s.Register")
		}
	}()
	return &pb.AsyncRegisterResponse{Data: true}, nil
}

func (s *upstreamService) asyncRegister(ctx context.Context, req *pb.AsyncRegisterRequest) (resp *pb.AsyncRegisterResponse, err error) {
	service := legacy_upstream.Service.Clone(ctx)

	reqDto := dto.FromUpstream(req.Upstream)
	if req.GetUpstream().GetScene() == orm.HubScene {
		if err := s.patchHubInfo(ctx, reqDto); err != nil {
			return nil, err
		}
	}
	if err = reqDto.Init(); err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, err.Error())
		return
	}
	for i := 0; i < len(reqDto.ApiList); i++ {
		if err = (&reqDto.ApiList[i]).Init(); err != nil {
			err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, err.Error())
			return
		}
	}
	result, err := service.UpstreamRegisterAsync(ctx, reqDto)
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.AsyncRegisterResponse{
		Data: result,
	}
	return
}

func (s *upstreamService) patchHubInfo(ctx context.Context, reqDto *dto.UpstreamRegisterDto) error {
	if len(reqDto.ApiList) == 0 || reqDto.ApiList[0].Domain == "" {
		return errors.New("no api in list")
	}
	hubInfo, err := s.hubInfoService.GetHubInfo(ctx, &pb2.GetHubInfoReq{
		OrgID:        reqDto.OrgId,
		ProjectID:    reqDto.ProjectId,
		Env:          reqDto.Env,
		Az:           reqDto.Az,
		OneOfDomains: strings.Split(reqDto.ApiList[0].Domain, ",")[0],
	})
	if err != nil {
		return err
	}
	if hubInfo == nil || !hubInfo.GetSuccess() {
		return errors.New("hub info not found")
	}
	reqDto.HubInfoID = hubInfo.GetData().GetId()
	return nil
}
