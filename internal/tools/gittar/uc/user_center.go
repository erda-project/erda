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

	"github.com/sirupsen/logrus"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	userpb "github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
)

var userServiceServer userpb.UserServiceServer

func InitializeUcClient(identity userpb.UserServiceServer) {
	userServiceServer = identity
	logrus.Infof("gittar uc client set up")
}

func FindUserById(id string) (*commonpb.UserInfo, error) {
	ctx := apis.WithInternalClientContext(context.Background(), discover.SvcGittar)
	userResp, err := userServiceServer.GetUser(ctx, &userpb.GetUserRequest{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	user := userResp.Data
	return &commonpb.UserInfo{
		Id:     user.Id,
		Name:   user.Name,
		Nick:   user.Nick,
		Avatar: user.Avatar,
		Phone:  user.Phone,
		Email:  user.Email,
	}, nil
}
