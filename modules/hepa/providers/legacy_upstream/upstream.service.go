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
	context "context"

	"github.com/pkg/errors"

	pb "github.com/erda-project/erda-proto-go/core/hepa/legacy_upstream/pb"
	"github.com/erda-project/erda/modules/hepa/common/vars"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/modules/hepa/services/legacy_upstream"
	erdaErr "github.com/erda-project/erda/pkg/common/errors"
)

type upstreamService struct {
	p *provider
}

func (s *upstreamService) Register(ctx context.Context, req *pb.RegisterRequest) (resp *pb.RegisterResponse, err error) {
	service := legacy_upstream.Service.Clone(ctx)
	if req.Upstream == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "invalid request")
		return
	}
	result, err := service.UpstreamRegister(dto.FromUpstream(req.Upstream))
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.RegisterResponse{
		Data: result,
	}
	return

}
func (s *upstreamService) AsyncRegister(ctx context.Context, req *pb.AsyncRegisterRequest) (resp *pb.AsyncRegisterResponse, err error) {
	service := legacy_upstream.Service.Clone(ctx)
	if req.Upstream == nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, "invalid request")
		return
	}
	result, err := service.UpstreamRegisterAsync(dto.FromUpstream(req.Upstream))
	if err != nil {
		err = erdaErr.NewInvalidParameterError(vars.TODO_PARAM, errors.Cause(err).Error())
		return
	}
	resp = &pb.AsyncRegisterResponse{
		Data: result,
	}
	return

}
