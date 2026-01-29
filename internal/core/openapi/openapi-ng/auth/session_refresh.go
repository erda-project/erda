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

package auth

type sessionRefreshCtxKey struct{}

//func WithSessionRefresh(ctx context.Context, refresh *common.SessionRefresh) context.Context {
//	if refresh == nil {
//		return ctx
//	}
//	return context.WithValue(ctx, sessionRefreshCtxKey{}, refresh)
//}
//
//func GetSessionRefresh(ctx context.Context) *common.SessionRefresh {
//	if ctx == nil {
//		return nil
//	}
//	v := ctx.Value(sessionRefreshCtxKey{})
//	if v == nil {
//		return nil
//	}
//	if refresh, ok := v.(*common.SessionRefresh); ok {
//		return refresh
//	}
//	return nil
//}
