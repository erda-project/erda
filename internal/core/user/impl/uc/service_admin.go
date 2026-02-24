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
	"bytes"
	"context"
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cast"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

var ucLoginMethodI18nMap = map[string]map[string]string{
	"username": {"en-US": "username", "zh-CN": "账密登录", "marks": "external"},
	"sso":      {"en-US": "sso", "zh-CN": "单点登录", "marks": "internal"},
	"email":    {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
	"mobile":   {"en-US": "default", "zh-CN": "默认登录方式", "marks": ""},
}

func (p *provider) UserListLoginMethod(ctx context.Context, req *pb.UserListLoginMethodRequest) (*pb.UserListLoginMethodResponse, error) {
	res, err := p.handleListLoginMethod()
	if err != nil {
		return nil, err
	}

	locale := apis.GetLang(ctx)
	if locale == "" {
		locale = "zh-CN"
	}

	deDup := make(map[string]struct{})
	var methods []*pb.UserLoginMethod
	for _, v := range res.RegistryType {
		tmp := getLoginTypeByUC(v)
		if tmp == nil {
			continue
		}
		if _, ok := deDup[tmp["marks"]]; ok {
			continue
		}
		methods = append(methods, &pb.UserLoginMethod{
			DisplayName: tmp[locale],
			Value:       tmp["marks"],
		})
		deDup[tmp["marks"]] = struct{}{}
	}

	return &pb.UserListLoginMethodResponse{Data: methods}, nil
}

func (p *provider) PwdSecurityConfigGet(_ context.Context, _ *pb.PwdSecurityConfigGetRequest) (*pb.PwdSecurityConfigGetResponse, error) {
	config, err := p.handleGetPwdSecurityConfig()
	if err != nil {
		return nil, err
	}

	return &pb.PwdSecurityConfigGetResponse{
		Data: &pb.PwdSecurityConfig{
			CaptchaChallengeNumber:   config.CaptchaChallengeNumber,
			ContinuousPwdErrorNumber: config.ContinuousPwdErrorNumber,
			MaxPwdErrorNumber:        config.MaxPwdErrorNumber,
			ResetPassWordPeriod:      config.ResetPassWordPeriod,
		},
	}, nil
}

func (p *provider) PwdSecurityConfigUpdate(ctx context.Context, req *pb.PwdSecurityConfigUpdateRequest) (*pb.PwdSecurityConfigUpdateResponse, error) {
	if err := p.handleUpdatePwdSecurityConfig(&PwdSecurityConfig{
		CaptchaChallengeNumber:   req.CaptchaChallengeNumber,
		ContinuousPwdErrorNumber: req.ContinuousPwdErrorNumber,
		MaxPwdErrorNumber:        req.MaxPwdErrorNumber,
		ResetPassWordPeriod:      req.ResetPassWordPeriod,
	}); err != nil {
		return nil, err
	}

	return &pb.PwdSecurityConfigUpdateResponse{}, nil
}

func (p *provider) UserBatchFreeze(ctx context.Context, req *pb.UserBatchFreezeRequest) (*pb.UserBatchFreezeResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleFreezeUser(id, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchFreezeResponse{}, nil
}

func (p *provider) UserBatchUnfreeze(ctx context.Context, req *pb.UserBatchUnFreezeRequest) (*pb.UserBatchUnFreezeResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		if err := p.handleUnfreezeUser(id, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUnFreezeResponse{}, nil
}

func (p *provider) UserBatchUpdateLoginMethod(ctx context.Context, req *pb.UserBatchUpdateLoginMethodRequest) (*pb.UserBatchUpdateLoginMethodResponse, error) {
	operatorID := apis.GetUserID(ctx)
	for _, id := range req.UserIDs {
		updateReq := &UpdateLoginMethodRequest{
			ID:     id,
			Source: req.Source,
		}
		if err := p.handleUpdateLoginMethod(updateReq, operatorID); err != nil {
			return nil, err
		}
	}
	return &pb.UserBatchUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserCreate(ctx context.Context, req *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	operatorID := apis.GetUserID(ctx)
	if err := p.handleCreateUsers(req, operatorID); err != nil {
		return nil, err
	}
	return &pb.UserCreateResponse{}, nil
}

func (p *provider) UserExport(ctx context.Context, req *pb.UserPagingRequest) (*pb.UserExportResponse, error) {
	var (
		users  []*pb.ManagedUser
		total  int64
		pageNo = req.PageNo
	)

	if pageNo == 0 {
		pageNo = 100
	}

	for {
		data, err := p.UserPaging(ctx, req)
		if err != nil {
			return nil, err
		}

		if total == 0 {
			total = data.Total
		}

		users = append(users, data.List...)
		if int64(len(users)) >= total {
			break
		}
		req.PageNo++
	}

	locale := apis.GetLang(ctx)
	if locale == "" {
		locale = "zh-CN"
	}

	loginMethodMap, err := p.getLoginMethodMap(locale)
	if err != nil {
		return nil, err
	}

	return &pb.UserExportResponse{
		Total:        total,
		List:         users,
		LoginMethods: loginMethodMap,
	}, nil
}

func (p *provider) UserFreeze(ctx context.Context, req *pb.UserFreezeRequest) (*pb.UserFreezeResponse, error) {
	if err := p.handleFreezeUser(req.UserID, apis.GetUserID(ctx)); err != nil {
		return nil, err
	}
	return &pb.UserFreezeResponse{}, nil
}

func (p *provider) UserUnfreeze(ctx context.Context, req *pb.UserUnfreezeRequest) (*pb.UserUnfreezeResponse, error) {
	if err := p.handleUnfreezeUser(req.UserID, apis.GetUserID(ctx)); err != nil {
		return nil, err
	}
	return &pb.UserUnfreezeResponse{}, nil
}

func (p *provider) UserUpdateLoginMethod(ctx context.Context, req *pb.UserUpdateLoginMethodRequest) (*pb.UserUpdateLoginMethodResponse, error) {
	id := req.ID
	if id == "" {
		id = req.UserID
	}
	operatorID := apis.GetUserID(ctx)
	updateReq := &UpdateLoginMethodRequest{
		ID:     id,
		Source: req.Source,
	}
	if err := p.handleUpdateLoginMethod(updateReq, operatorID); err != nil {
		return nil, err
	}
	return &pb.UserUpdateLoginMethodResponse{}, nil
}

func (p *provider) UserUpdateUserinfo(ctx context.Context, req *pb.UserUpdateInfoRequest) (*pb.UserUpdateInfoResponse, error) {
	operatorID := apis.GetUserID(ctx)
	if req.UserID == "" {
		return nil, errors.New("user id is empty")
	}

	updateInfoReq := &UpdateUserInfoRequest{
		ID: req.UserID,
	}
	if req.Nick != "" {
		updateInfoReq.Nick = req.Nick
	}
	if req.Name != "" {
		updateInfoReq.UserName = req.Name
	}
	if req.Mobile != "" {
		updateInfoReq.Mobile = req.Mobile
	}
	if req.Email != "" {
		updateInfoReq.Email = req.Email
	}

	if err := p.handleUpdateUserInfo(updateInfoReq, operatorID); err != nil {
		return nil, err
	}
	return &pb.UserUpdateInfoResponse{}, nil
}

func getLoginTypeByUC(key string) map[string]string {
	if v, ok := ucLoginMethodI18nMap[key]; ok {
		return v
	}
	return nil
}

func convertCreateUserItem(item *pb.UserCreateItem) CreateUserItem {
	return CreateUserItem{
		Username: item.Name,
		Nickname: item.Nick,
		Mobile:   item.Phone,
		Email:    item.Email,
		Password: item.Password,
	}
}

func (p *provider) getLoginMethodMap(locale string) (map[string]string, error) {
	res, err := p.handleListLoginMethod()
	if err != nil {
		return nil, err
	}

	valueDisplayNameMap := make(map[string]string)
	deDupMap := make(map[string]struct{})
	for _, v := range res.RegistryType {
		tmp := getLoginTypeByUC(v)
		if tmp == nil {
			continue
		}
		if _, ok := deDupMap[tmp["marks"]]; ok {
			continue
		}
		valueDisplayNameMap[tmp["marks"]] = tmp[locale]
		deDupMap[tmp["marks"]] = struct{}{}
	}

	return valueDisplayNameMap, nil
}

func (p *provider) UserPaging(_ context.Context, req *pb.UserPagingRequest) (*pb.UserPagingResponse, error) {
	client, err := p.newAuthedClient(nil)
	if err != nil {
		return nil, err
	}

	conditions := url.Values{}
	if req.Name != "" {
		conditions.Add("username", req.Name)
	}
	if req.Nick != "" {
		conditions.Add("nickname", req.Nick)
	}
	if req.Phone != "" {
		conditions.Add("mobile", req.Phone)
	}
	if req.Email != "" {
		conditions.Add("email", req.Email)
	}
	if req.Locked {
		conditions.Add("locked", strconv.Itoa(1))
	}
	if req.Source != "" {
		conditions.Add("source", req.Source)
	}
	if req.PageNo > 0 {
		conditions.Add("pageNo", cast.ToString(req.PageNo))
	}
	if req.PageSize > 0 {
		conditions.Add("pageSize", cast.ToString(req.PageSize))
	}

	var (
		path = "/api/user/admin/paging"
		body bytes.Buffer
	)

	r, err := client.Get(p.Cfg.Host).Path(path).
		Params(conditions).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to paging user")
	}
	if !r.IsOK() {
		return nil, errors.Errorf("failed to call %s, status code: %d, resp: %s", path, r.StatusCode(), body.String())
	}

	var users Response[*UserPaging]
	if err := json.NewDecoder(&body).Decode(&users); err != nil {
		return nil, err
	}

	pbUsers := lo.Map(users.Result.Data, func(datum *UserDto, _ int) *pb.ManagedUser {
		return managedUserMapper(datum)
	})

	return &pb.UserPagingResponse{
		Total: users.Result.Total,
		List:  pbUsers,
	}, nil
}
