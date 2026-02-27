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
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/pkg/transport/interceptor"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/auth/akutil"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/handlers"
	"github.com/erda-project/erda/internal/pkg/audit"
)

var (
	noNeedAuthMethodsMu sync.RWMutex
	noNeedAuthMethods   = map[string]struct{}{}
)

type MethodPermission struct {
	Method            interface{}
	OnlyAdmin         bool
	OnlyClient        bool
	AdminOrClient     bool // no client-token
	LoggedIn          bool // just logged in
	NoNeedAuth        bool
	SkipSetClientInfo bool
}

func CheckPermissions(perms ...*MethodPermission) transport.ServiceOption {
	methods := make(map[string]*MethodPermission)
	for _, perm := range perms {
		methodName := audit.GetMethodName(perm.Method)
		methods[methodName] = perm
		if perm.NoNeedAuth {
			noNeedAuthMethodsMu.Lock()
			noNeedAuthMethods[methodName] = struct{}{}
			noNeedAuthMethodsMu.Unlock()
		}
	}
	return transport.WithInterceptors(
		checkOneMethodPermission(methods),
		checkAndSetClientId(methods),
	)
}

func IsNoNeedAuthMethod(ctx context.Context) bool {
	info := transport.ContextServiceInfo(ctx)
	if info == nil {
		return false
	}
	return IsNoNeedAuthMethodName(info.Method())
}

func IsNoNeedAuthMethodName(methodName string) bool {
	noNeedAuthMethodsMu.RLock()
	_, ok := noNeedAuthMethods[methodName]
	noNeedAuthMethodsMu.RUnlock()
	return ok
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
			if perm.NoNeedAuth {
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
			if perm.OnlyClient {
				if !auth.IsClient(ctx) {
					return nil, handlers.ErrNoPermission
				}
				logrus.Infof("[pass] only ak permission, method: %s", fullMethodName)
				return h(ctx, req)
			}
			if perm.AdminOrClient {
				if !auth.IsAdminOrClient(ctx) {
					return nil, handlers.ErrNoPermission
				}
				logrus.Infof("[pass] admin or ak permission, method: %s", fullMethodName)
				return h(ctx, req)
			}
			if perm.LoggedIn {
				if !auth.IsLoggedIn(ctx) {
					return nil, handlers.ErrNotAuthorized
				}
				logrus.Infof("[pass] logged-in permission, method: %s", fullMethodName)
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
			clientToken, ok := ctxhelper.GetClientToken(ctx)
			var clientTokenId string
			if ok {
				clientTokenId = clientToken.Id
			}
			if err := akutil.AutoCheckAndSetClientInfo(clientId, clientTokenId, req, perm.SkipSetClientInfo); err != nil {
				return nil, err
			}
			return h(ctx, req)
		}
	}
}
