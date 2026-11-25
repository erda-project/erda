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
	"testing"

	"github.com/erda-project/erda-proto-go/core/user/pb"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

type fakeSysAdminChecker struct {
	admin  bool
	err    error
	userID string
}

func (f *fakeSysAdminChecker) IsSysAdmin(userID string) (bool, error) {
	f.userID = userID
	return f.admin, f.err
}

type fakeUserService struct {
	pb.UnimplementedUserServiceServer
	called bool
}

func (f *fakeUserService) UserCreate(context.Context, *pb.UserCreateRequest) (*pb.UserCreateResponse, error) {
	f.called = true
	return &pb.UserCreateResponse{}, nil
}

func TestAuthorizedUserServiceUserCreate(t *testing.T) {
	tests := []struct {
		name            string
		userID          string
		admin           bool
		checkErr        error
		wantCode        string
		wantUserService bool
	}{
		{
			name:            "sysadmin can create user",
			userID:          "1",
			admin:           true,
			wantUserService: true,
		},
		{
			name:     "normal user cannot create user",
			userID:   "2",
			admin:    false,
			wantCode: "AccessDenied",
		},
		{
			name:     "missing user is not logged in",
			userID:   "",
			wantCode: "NotLogin",
		},
		{
			name:     "sysadmin query failure denies access",
			userID:   "3",
			checkErr: errors.New("query failed"),
			wantCode: "InternalError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userService := &fakeUserService{}
			checker := &fakeSysAdminChecker{admin: tt.admin, err: tt.checkErr}
			service := &authorizedUserService{
				Interface:       userService,
				sysAdminChecker: checker,
			}

			ctx := apis.WithUserIDContext(context.Background(), tt.userID)
			_, err := service.UserCreate(ctx, &pb.UserCreateRequest{})
			if tt.wantCode == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			} else {
				apiErr, ok := err.(*errorresp.APIError)
				if !ok {
					t.Fatalf("expected APIError, got %T", err)
				}
				if apiErr.Code() != tt.wantCode {
					t.Fatalf("expected error code %q, got %q", tt.wantCode, apiErr.Code())
				}
			}
			if userService.called != tt.wantUserService {
				t.Fatalf("expected user service called=%v, got %v", tt.wantUserService, userService.called)
			}
			if tt.userID != "" && checker.userID != tt.userID {
				t.Fatalf("expected checked user %q, got %q", tt.userID, checker.userID)
			}
		})
	}
}
