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

package user

import (
	"context"
	"errors"

	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	apierrors "github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/common/apis"
)

type sysAdminChecker interface {
	IsSysAdmin(userID string) (bool, error)
}

type authorizedUserService struct {
	common.Interface
	sysAdminChecker sysAdminChecker
}

func newAuthorizedUserService(service common.Interface, mysqlSvc mysql.Interface) common.Interface {
	var checker sysAdminChecker
	if mysqlSvc != nil {
		checker = &dao.DBClient{DB: mysqlSvc.DB()}
	}
	return &authorizedUserService{
		Interface:       service,
		sysAdminChecker: checker,
	}
}

func (s *authorizedUserService) UserCreate(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, apierrors.ErrInvoke.NotLogin()
	}
	if s.sysAdminChecker == nil {
		return nil, apierrors.ErrInvoke.InternalError(errors.New("sysadmin checker is unavailable"))
	}

	admin, err := s.sysAdminChecker.IsSysAdmin(userID)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !admin {
		return nil, apierrors.ErrInvoke.AccessDenied()
	}

	return s.Interface.UserCreate(ctx, req)
}
