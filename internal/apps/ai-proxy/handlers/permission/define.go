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

package permission

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/pkg/audit"
)

type MethodPermission struct {
	Method                 interface{}
	OnlyAdmin              bool
	OnlyAk                 bool
	AdminOrAk              bool
	NoAuth                 bool
	CheckButNotSetClientId bool
}

func CheckPermissions(perms ...*MethodPermission) transport.ServiceOption {
	methods := make(map[string]*MethodPermission)
	for _, perm := range perms {
		methodName := audit.GetMethodName(perm.Method)
		methods[methodName] = perm
	}
	return transport.WithInterceptors(
		checkOneMethodPermission(methods),
		checkAndSetClientId(methods),
	)
}

var checkOneMethodPermission = func(methods map[string]*MethodPermission) interceptor.Interceptor {
	return func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			fullMethodName := transport.GetFullMethodName(ctx)
			info := transport.ContextServiceInfo(ctx)
			perm := methods[info.Method()]
			if perm == nil {
				logrus.Errorf("[reject] permission undefined, method: %s", fullMethodName)
				return nil, handlers.ErrNoPermission
			}
			if perm.NoAuth {
				logrus.Infof("[pass] no auth permission, method: %s", fullMethodName)
				return h(ctx, req)
			}
			if perm.OnlyAdmin {
				if !auth.IsAdmin(ctx) {
					return nil, handlers.ErrNoAdminPermission
				}
				logrus.Infof("[pass] only admin permission, method: %s", fullMethodName)
				return h(ctx, req)
			}
			if perm.OnlyAk {
				if !auth.IsClient(ctx) {
					return nil, handlers.ErrNoPermission
				}
				logrus.Infof("[pass] only ak permission, method: %s", info.Method())
				return h(ctx, req)
			}
			if perm.AdminOrAk {
				if !auth.Valid(ctx) {
					return nil, handlers.ErrNoPermission
				}
				logrus.Infof("[pass] admin or ak permission, method: %s", info.Method())
				return h(ctx, req)
			}
			logrus.Errorf("[reject] should not be here, method: %s", fullMethodName)
			return nil, handlers.ErrNoPermission
		}
	}
}

var checkAndSetClientId = func(methods map[string]*MethodPermission) interceptor.Interceptor {
	return func(h interceptor.Handler) interceptor.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			fullMethodName := transport.GetFullMethodName(ctx)
			info := transport.ContextServiceInfo(ctx)
			perm := methods[info.Method()]
			if perm == nil {
				logrus.Errorf("[reject] permission undefined, method: %s", fullMethodName)
				return nil, handlers.ErrNoPermission
			}
			clientId, ok := ctxhelper.GetClientId(ctx)
			if !ok || clientId == "" {
				return h(ctx, req)
			}
			if err := akutil.AutoCheckAndSetClientId(clientId, req, perm.CheckButNotSetClientId); err != nil {
				return nil, err
			}
			return h(ctx, req)
		}
	}
}
