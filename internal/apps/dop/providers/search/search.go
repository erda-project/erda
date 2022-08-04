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

package search

import (
	"context"

	"github.com/erda-project/erda-proto-go/dop/search/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/providers/issue/core/query"
	"github.com/erda-project/erda/internal/apps/dop/providers/search/handlers"
	"github.com/erda-project/erda/internal/apps/dop/providers/search/handlers/issue"
	"github.com/erda-project/erda/internal/apps/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type Interface interface {
	pb.SearchServiceServer
}

type ServiceImpl struct {
	bdl   *bundle.Bundle
	query query.Interface
}

func (p *ServiceImpl) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	req.IdentityInfo = apis.GetIdentityInfo(ctx)
	req.IdentityInfo.OrgID = apis.GetOrgID(ctx)
	orgID, err := apis.GetIntOrgID(ctx)
	if err != nil {
		return nil, err
	}
	req.OrgID = uint64(orgID)
	if err := p.checkRequest(req); err != nil {
		return nil, err
	}
	h := handlers.NewBaseSearch(issue.NewIssueHandler(
		issue.WithBundle(p.bdl),
		issue.WithQuery(p.query),
	))
	h.BeginSearch(ctx, req)
	res, err := h.GetResult()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (p *ServiceImpl) checkRequest(req *pb.SearchRequest) error {
	if req.IdentityInfo == nil ||
		req.IdentityInfo.OrgID == "" ||
		req.IdentityInfo.UserID == "" {
		return apierrors.ErrCheckPermission.NotLogin()
	}
	if req.Query == "" {
		return apierrors.ErrGlobalSearch.MissingParameter("query")
	}
	return nil
}
