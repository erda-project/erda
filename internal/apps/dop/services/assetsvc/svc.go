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

// Package asset API 资产
package assetsvc

import (
	"context"
	"strconv"

	"github.com/erda-project/erda-infra/providers/i18n"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/services/branchrule"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

type Service struct {
	trans         i18n.Translator
	branchRuleSvc *branchrule.BranchRule
	bdl           *bundle.Bundle
	org           org.Interface
	userService   userpb.UserServiceServer
}

type Option func(*Service)

func New(opts ...Option) *Service {
	r := &Service{}
	for _, op := range opts {
		op(r)
	}
	return r
}

// WithI18n sets the i18n client
func WithI18n(translator i18n.Translator) Option {
	return func(svc *Service) {
		svc.trans = translator
	}
}

// WithBranchRuleSvc sets the branch rule client
func WithBranchRuleSvc(branchRule *branchrule.BranchRule) Option {
	return func(service *Service) {
		service.branchRuleSvc = branchRule
	}
}

func WithBundle(bdl *bundle.Bundle) Option {
	return func(svc *Service) {
		svc.bdl = bdl
	}
}

func WithOrg(org org.Interface) Option {
	return func(svc *Service) {
		svc.org = org
	}
}

func WithUserService(userService userpb.UserServiceServer) Option {
	return func(svc *Service) {
		svc.userService = userService
	}
}

func (svc *Service) getOrg(ctx context.Context, orgID uint64) (*orgpb.Org, error) {
	orgResp, err := svc.org.GetOrg(apis.WithInternalClientContext(ctx, discover.SvcDOP),
		&orgpb.GetOrgRequest{IdOrName: strconv.FormatUint(orgID, 10)})
	if err != nil {
		return nil, err
	}
	return orgResp.Data, nil
}
