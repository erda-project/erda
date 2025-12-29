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
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/internal/core/user/common"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/pointer"
)

func (p *provider) newAuthedClient(refresh *bool) (*httpclient.HTTPClient, error) {
	oauthToken, err := p.OAuthTokenProvider.ExchangeClientCredentials(
		context.Background(), pointer.BoolDeref(refresh, false), nil,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to exchange client credentials token")
	}

	return p.client.BearerTokenAuth(oauthToken.AccessToken), nil
}

func (p *provider) FindUsers(ctx context.Context, req *pb.FindUsersRequest) (*pb.FindUsersResponse, error) {
	if len(req.IDs) == 0 {
		return &pb.FindUsersResponse{}, nil
	}
	intIds := make([]int, 0, len(req.IDs))
	for _, id := range req.IDs {
		intId, err := strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
		intIds = append(intIds, intId)
	}

	//sysOpExist := strutil.Exist(ids, common.SystemOperator)
	//if sysOpExist {
	//	ids = strutil.RemoveSlice(ids, common.SystemOperator)
	//}

	users, err := p.findByIDs(intIds)
	if err != nil {
		return nil, err
	}
	//if sysOpExist {
	//	users = append(users, common.SystemUser)
	//}
	userList := make([]*pb.User, 0, len(users))
	for _, i := range users {
		userList = append(userList, common.ToPbUser(i))
	}
	return &pb.FindUsersResponse{Data: userList}, nil
}

// FindUsersByKey find users by key (username, nickname, mobile, email)
func (p *provider) FindUsersByKey(ctx context.Context, req *pb.FindUsersByKeyRequest) (*pb.FindUsersByKeyResponse, error) {
	key := req.Key
	if key == "" {
		return &pb.FindUsersByKeyResponse{}, nil
	}
	g, ctx := errgroup.WithContext(context.Background())

	userMap := make(map[string]*pb.User)
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
				pbUser := common.ToPbUser(*userMapper(u))
				userMap[pbUser.ID] = pbUser
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	foundUsers := make([]*pb.User, 0, len(userMap))
	for _, u := range userMap {
		foundUsers = append(foundUsers, u)
	}

	return &pb.FindUsersByKeyResponse{
		Data: foundUsers,
	}, nil
}

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
			JSONBody(&common.IAMUserCreate{
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

func (p *provider) UserUpdateUserinfo(ctx context.Context, req *pb.UserUpdateInfoRequset) (*pb.UserUpdateInfoResponse, error) {
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
func (p *provider) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	user, err := p.getUser(req.UserID)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserResponse{
		Data: common.ToPbUser(*user),
	}, nil
}

func (p *provider) PwdSecurityConfigGet(ctx context.Context, request *pb.PwdSecurityConfigGetRequest) (*pb.PwdSecurityConfigGetResponse, error) {
	return nil, errors.New("iam not support get password security config direct")
}

func (p *provider) PwdSecurityConfigUpdate(ctx context.Context, request *pb.PwdSecurityConfigUpdateRequest) (*pb.PwdSecurityConfigUpdateResponse, error) {
	return nil, errors.New("iam not support update password security config direct")
}

func (p *provider) updateProfile(userId string, newVal map[string]any) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/iam/api/v1/admin/user/%s/update"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(fmt.Sprintf(path, userId)).
		JSONBody(newVal).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to update user profile")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp common.IAMResponse[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Data {
		return errors.New("failed to update user profile")
	}

	return nil
}

func (p *provider) userUnlock(userId string) error {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return err
	}

	var (
		path = "/iam/api/v1/admin/user/%s/unlock"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).Path(fmt.Sprintf(path, userId)).
		JSONBody(map[string]any{}).
		Do().Body(&body)
	if err != nil {
		return errors.Wrapf(err, "failed to create user")
	}
	if !r.IsOK() {
		return errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp common.IAMResponse[bool]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return err
	}

	if !resp.Data {
		return errors.New("failed to unlock user")
	}

	return nil
}

func (p *provider) pagingQuery(no, size int64, conditions map[string]any, plainText bool) (*common.IAMPagingData[[]*common.IAMUserDto], error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	conditions["no"] = no
	conditions["size"] = size

	var (
		path = "/iam/api/v1/admin/user/paging"
		body bytes.Buffer
	)

	if plainText {
		path = "/iam/api/v1/admin/user/plaintext/paging"
	}

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(conditions).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by query")
	}

	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, body: %s", path, r.StatusCode(), body.String())
	}

	var resp common.IAMResponse[*common.IAMPagingData[[]*common.IAMUserDto]]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (p *provider) findUsersByQuery(fieldName, key string) ([]*common.IAMUserDto, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/iam/api/v1/admin/user/list"
		resp common.IAMResponse[[]*common.IAMUserDto]
	)

	r, err := client.Post(p.Cfg.Host).Path(path).
		JSONBody(map[string]string{
			fieldName: key,
		}).
		Do().JSON(&resp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by query")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d", path, r.StatusCode())
	}

	return resp.Data, nil
}

func (p *provider) findByIDs(ids []int) ([]common.User, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = "/iam/api/v1/admin/user/find-by-ids"
		body bytes.Buffer
	)

	r, err := client.Post(p.Cfg.Host).
		Path(path).JSONBody(map[string][]int{
		"userIds": ids,
	}).Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find users by ids")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var resp common.IAMResponse[[]common.IAMUserDto]
	if err := json.NewDecoder(&body).Decode(&resp); err != nil {
		return nil, err
	}

	userList := make([]common.User, len(resp.Data))
	for i, user := range resp.Data {
		userList[i] = *userMapper(&user)
	}

	return userList, nil
}

func (p *provider) getUser(userID string) (*common.User, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	var (
		path = fmt.Sprintf("/iam/api/v1/admin/user/%s/find", userID)
		resp common.IAMResponse[common.IAMUserDto]
	)

	r, err := client.Get(p.Cfg.Host).
		Path(path).
		Do().JSON(&resp)
	if err != nil {
		return nil, err
	}

	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d", path, r.StatusCode())
	}
	if resp.Data.ID == 0 {
		return nil, errors.Errorf("failed to find user %s", userID)
	}

	return userMapper(&resp.Data), nil
}
