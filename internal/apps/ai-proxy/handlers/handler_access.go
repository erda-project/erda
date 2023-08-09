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

package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/providers/dao"
)

var (
	_ pb.AccessServer = (*AccessHandler)(nil)
)

type AccessHandler struct {
	Log logs.Logger
	Dao dao.DAO
}

func (h *AccessHandler) Access(ctx context.Context, req *pb.AccessReq) (*pb.AccessResponse, error) {
	switch platform := req.GetPlatform(); strings.ToLower(platform) {
	case "erda":
		switch scope := strings.ToLower(req.GetScope()); scope {
		case "org", "organization":
			return h.AccessInOrg(ctx, &pb.OrgAccessReq{OrgId: req.GetOrgId(), UserId: req.GetUserId()})
		default:
			return nil, HTTPError(errors.Errorf("invalid scope %s in the platform %s", scope, platform), http.StatusBadRequest)
		}
	default:
		return nil, HTTPError(errors.Errorf("invalid platform %s", platform), http.StatusBadRequest)
	}
}

func (h *AccessHandler) AccessInOrg(ctx context.Context, req *pb.OrgAccessReq) (*pb.AccessResponse, error) {
	// TODO: hard code here
	switch req.GetOrgName() {
	case "erda", "terminus":
		return &pb.AccessResponse{Access: true}, nil
	default:
		return &pb.AccessResponse{Access: false}, nil
	}
}
