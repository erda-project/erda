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
	"bytes"
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/erda-project/erda-proto-go/core/user/pb"
)

func (p *provider) UserPaging(ctx context.Context, req *pb.UserPagingRequest) (*pb.UserPagingResponse, error) {
	conditions := map[string]any{
		"sort": "createdAt_DESC",
	}
	if !isEmptyTrim(req.Name) {
		conditions["username"] = req.Name
	}
	if !isEmptyTrim(req.Email) {
		conditions["email"] = req.Email
	}
	if !isEmptyTrim(req.Nick) {
		conditions["nickname"] = req.Nick
	}
	if !isEmptyTrim(req.Phone) {
		conditions["mobile"] = req.Phone
	}
	if req.Locked {
		conditions["locked"] = req.Locked
	}

	users, err := p.pagingQuery(req.PageNo, req.PageSize, conditions, true)
	if err != nil {
		return nil, err
	}

	userList := make([]*pb.ManagedUser, 0, len(users.Data))
	for _, datum := range users.Data {
		pbUser, err := managedUserMapper(datum)
		if err != nil {
			return nil, err
		}
		userList = append(userList, pbUser)
	}

	return &pb.UserPagingResponse{
		Total: int64(users.Total),
		List:  userList,
	}, nil
}

func (p *provider) UserListLoginMethod(ctx context.Context, req *pb.UserListLoginMethodRequest) (*pb.UserListLoginMethodResponse, error) {
	// TODO: support more login method
	return &pb.UserListLoginMethodResponse{
		Data: []*pb.UserLoginMethod{
			{
				DisplayName: "DEFAULT",
				Value:       "",
			},
		},
	}, nil
}

func (p *provider) UserBatchFreeze(ctx context.Context, req *pb.UserBatchFreezeRequest) (*pb.UserBatchFreezeResponse, error) {
	// Not support freeze user now, ignore
	return nil, errors.New("not support freeze user")
}

func (p *provider) UserBatchUnfreeze(ctx context.Context, req *pb.UserBatchUnFreezeRequest) (*pb.UserBatchUnFreezeResponse, error) {
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(20)

	for _, userID := range req.UserIDs {
		g.Go(func() error {
			return p.userUnlock(userID)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &pb.UserBatchUnFreezeResponse{}, nil
}

func (p *provider) UserBatchUpdateLoginMethod(ctx context.Context, req *pb.UserBatchUpdateLoginMethodRequest) (*pb.UserBatchUpdateLoginMethodResponse, error) {
	// Not support update login method now, ignore
	return &pb.UserBatchUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserCreate(_ context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/iam/api/v1/admin/user/create"
		resp bytes.Buffer
	)

	for _, user := range req.Users {
		r, err := client.Post(p.Cfg.Host).Path(path).
			JSONBody(&UserCreate{
				UserName:          user.Name,
				NickName:          user.Nick,
				Email:             user.Email,
				Mobile:            user.Phone,
				Password:          user.Password,
				NeedResetPassword: false,
			}).
			Do().Body(&resp)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create user")
		}
		if !r.IsOK() {
			return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), resp.String())
		}
	}

	return &pb.UserCreateResponse{}, nil
}

func (p *provider) UserExport(ctx context.Context, req *pb.UserPagingRequest) (*emptypb.Empty, error) {
	//TODO implement me
	panic("implement me")
}

func (p *provider) UserFreeze(ctx context.Context, req *pb.UserFreezeRequest) (*pb.UserFreezeResponse, error) {
	return nil, errors.New("not support freeze user")
}

func (p *provider) UserUnfreeze(ctx context.Context, req *pb.UserUnfreezeRequest) (*pb.UserUnfreezeResponse, error) {
	if err := p.userUnlock(req.UserID); err != nil {
		return nil, err
	}
	return &pb.UserUnfreezeResponse{}, nil
}

func (p *provider) UserUpdateLoginMethod(ctx context.Context, req *pb.UserUpdateLoginMethodRequest) (*pb.UserUpdateLoginMethodResponse, error) {
	// Not support update login method now, ignore
	return &pb.UserUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserUpdateUserinfo(ctx context.Context, req *pb.UserUpdateInfoRequest) (*pb.UserUpdateInfoResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("must provide user id")
	}

	// can update fields: nickname, email, mobile
	// TODO: clearly params, both existed from frontend: phone, mobile, userId, id...and more
	newVals := make(map[string]any)
	if !isEmptyTrim(req.Nick) {
		newVals["nickname"] = req.Nick
	}
	if !isEmptyTrim(req.Email) {
		newVals["email"] = req.Email
	}
	if !isEmptyTrim(req.Mobile) {
		newVals["mobile"] = req.Mobile
	}

	if err := p.updateProfile(req.UserID, newVals); err != nil {
		return nil, err
	}

	return &pb.UserUpdateInfoResponse{}, nil
}

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

func (p *provider) PwdSecurityConfigGet(ctx context.Context, request *pb.PwdSecurityConfigGetRequest) (*pb.PwdSecurityConfigGetResponse, error) {
	return nil, errors.New("iam not support get password security config direct")
}

func (p *provider) PwdSecurityConfigUpdate(ctx context.Context, request *pb.PwdSecurityConfigUpdateRequest) (*pb.PwdSecurityConfigUpdateResponse, error) {
	return nil, errors.New("iam not support update password security config direct")
}
