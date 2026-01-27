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

package uc

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/strutil"
)

func (p *provider) FindUsers(_ context.Context, req *pb.FindUsersRequest) (*pb.FindUsersResponse, error) {
	ids := req.IDs
	if len(ids) == 0 {
		return &pb.FindUsersResponse{}, nil
	}
	sysOpExist := strutil.Exist(ids, common.SystemOperator)
	if sysOpExist {
		ids = strutil.RemoveSlice(ids, common.SystemOperator)
	}

	parts := make([]string, 0)
	for _, id := range ids {
		parts = append(parts, "user_id:"+id)
	}
	query := strings.Join(parts, " OR ")
	users, err := p.handleQueryUsers(query)
	if err != nil {
		return nil, err
	}

	if sysOpExist {
		users = append(users, common.SystemUser)
	}

	pbUsers := make([]*commonpb.UserInfo, 0, len(users))
	if req.KeepOrder {
		userMap := lo.KeyBy(users, func(u *commonpb.UserInfo) string {
			return u.Id
		})
		for _, id := range req.IDs {
			if user, exists := userMap[id]; exists {
				pbUsers = append(pbUsers, user)
			}
		}
	} else {
		for _, i := range users {
			pbUsers = append(pbUsers, i)
		}
	}

	return &pb.FindUsersResponse{Data: pbUsers}, nil
}

func (p *provider) FindUsersByKey(ctx context.Context, req *pb.FindUsersByKeyRequest) (*pb.FindUsersByKeyResponse, error) {
	key := req.Key
	if key == "" {
		return &pb.FindUsersByKeyResponse{}, nil
	}
	query := fmt.Sprintf("username:%s OR nickname:%s OR phone_number:%s OR email:%s", key, key, key, key)
	users, err := p.handleQueryUsers(query)
	if err != nil {
		return nil, err
	}
	pbUsers := make([]*commonpb.UserInfo, 0, len(users))
	for _, i := range users {
		pbUsers = append(pbUsers, i)
	}
	return &pb.FindUsersByKeyResponse{Data: pbUsers}, nil
}

func (p *provider) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := p.handleGetUser(req.UserID)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserResponse{
		Data: user,
	}, nil
}

func (p *provider) UserMe(ctx context.Context, r *pb.UserMeRequest) (*commonpb.UserInfo, error) {
	userID := apis.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("must provide user id")
	}

	user, err := p.handleGetUser(userID)
	if err != nil {
		return nil, err
	}
	return &commonpb.UserInfo{
		Id:     user.Id,
		Name:   user.Name,
		Nick:   user.Nick,
		Avatar: user.Avatar,
		Phone:  user.Phone,
		Email:  user.Email,
	}, nil
}

func (p *provider) Me(ctx context.Context, r *pb.UserMeRequest) (*commonpb.UserInfo, error) {
	return p.UserMe(ctx, r)
}

func (p *provider) UserEventWebhook(_ context.Context, _ *pb.UserEventWebhookRequest) (*pb.UserEventWebhookResponse, error) {
	// uc don't need to impl
	return &pb.UserEventWebhookResponse{}, nil
}
