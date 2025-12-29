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

package prehandle

//var filterCookieLock sync.Once
//var rediscli *redis.Client
//var credStore common.CredentialStore
//
//// filter session cookie which is exist in redis, put it at req.context
//func FilterCookie(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
//	filterCookieLock.Do(func() {
//		rediscli = redis.NewFailoverClient(&redis.FailoverOptions{
//			MasterName:    conf.RedisMasterName(),
//			SentinelAddrs: strings.Split(conf.RedisSentinelAddrs(), ","),
//			Password:      conf.RedisPwd(),
//		})
//		credStore = ucstore.New(&ucstore.Config{
//			Redis:      rediscli,
//			CookieName: conf.SessionCookieName(),
//			Expire:     0, // only need Load
//		})
//	})
//	if credStore != nil {
//		if cred, err := credStore.Load(ctx, req); err == nil && cred != nil && cred.SessionID != "" {
//			*req = *(req.WithContext(context.WithValue(req.Context(), "session", cred.SessionID)))
//			return
//		}
//	}
//}
