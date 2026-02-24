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

package iam

import (
	"context"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/erda-project/erda-infra/pkg/strutil"
	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/common/apis"
)

// GetUser get user detail info
func (p *provider) GetUser(_ context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := p.getUser(req.UserID, true)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		Data: user,
	}, nil
}

func (p *provider) FindUsers(_ context.Context, req *pb.FindUsersRequest) (*pb.FindUsersResponse, error) {
	if len(req.IDs) == 0 {
		return &pb.FindUsersResponse{}, nil
	}
	sysOpExist := strutil.Exist(req.IDs, common.SystemOperator)
	if sysOpExist {
		req.IDs = strutil.RemoveSlice(req.IDs, common.SystemOperator)
	}

	intIds := make([]int, 0, len(req.IDs))
	for _, id := range req.IDs {
		intId, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		intIds = append(intIds, intId)
	}

	users, err := p.findByIDs(intIds, true)
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

	return &pb.FindUsersResponse{
		Data: pbUsers,
	}, nil
}

// FindUsersByKey find users by key (username, nickname, mobile, email)
func (p *provider) FindUsersByKey(ctx context.Context, req *pb.FindUsersByKeyRequest) (*pb.FindUsersByKeyResponse, error) {
	key := req.Key
	if key == "" {
		return &pb.FindUsersByKeyResponse{}, nil
	}
	g, ctx := errgroup.WithContext(context.Background())

	userMap := make(map[string]*commonpb.UserInfo)
	var mu sync.Mutex

	conditions := []string{"username", "nickname", "mobile", "email"}

	for _, condition := range conditions {
		g.Go(func() error {
			users, err := p.findUsersByQuery(condition, key)
			if err != nil {
				return err
			}

			mu.Lock()
			defer mu.Unlock()
			for _, u := range users {
				pbUser := userMapper(u)
				userMap[pbUser.Id] = pbUser
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	foundUsers := make([]*commonpb.UserInfo, 0, len(userMap))
	for _, u := range userMap {
		foundUsers = append(foundUsers, u)
	}

	return &pb.FindUsersByKeyResponse{
		Data: foundUsers,
	}, nil
}

func (p *provider) UserMe(ctx context.Context, req *pb.UserMeRequest) (*commonpb.UserInfo, error) {
	return p.Me(ctx, req)
}

func (p *provider) Me(ctx context.Context, _ *pb.UserMeRequest) (*commonpb.UserInfo, error) {
	u, err := p.getUser(apis.GetUserID(ctx), true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user")
	}
	return &commonpb.UserInfo{
		Id:     u.Id,
		Name:   u.Name,
		Nick:   u.Nick,
		Avatar: u.Avatar,
		Phone:  u.Phone,
		Email:  u.Email,
	}, nil
}
