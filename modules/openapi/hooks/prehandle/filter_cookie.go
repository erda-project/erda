// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package prehandle

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"github.com/go-redis/redis"

	"github.com/erda-project/erda/modules/openapi/auth"
	"github.com/erda-project/erda/modules/openapi/conf"
)

var filterCookieLock sync.Once
var rediscli *redis.Client

// filter session cookie which is exist in redis, put it at req.context
func FilterCookie(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	filterCookieLock.Do(func() {
		rediscli = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    conf.RedisMasterName(),
			SentinelAddrs: strings.Split(conf.RedisSentinelAddrs(), ","),
			Password:      conf.RedisPwd(),
		})
	})
	cs := req.Cookies()
	sessions := []*http.Cookie{}
	for _, c := range cs {
		if c.Name == conf.SessionCookieName() {
			sessions = append(sessions, c)
		}
	}
	if len(sessions) >= 1 {
		for _, session := range sessions {
			if conf.OryEnabled() {
				// TODO
				*req = *(req.WithContext(context.WithValue(req.Context(), "session", session.Value)))
				return
			}
			if _, err := rediscli.Get(auth.MkSessionKey(session.Value)).Result(); err == redis.Nil {
				continue
			} else if err != nil {
				continue
			}
			*req = *(req.WithContext(context.WithValue(req.Context(), "session", session.Value)))
			return
		}
	}
}
