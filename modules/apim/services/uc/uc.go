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

// 获取 uc 中的用户信息

package uc

import (
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/apim/conf"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
	"github.com/erda-project/erda/pkg/ucauth"
)

const (
	UserAPI = "/api/open/v1/users"
)

var (
	once      sync.Once
	tokenAuth *ucauth.UCTokenAuth
	client    *httpclient.HTTPClient
)

type User struct {
	ID       uint64  `json:"user_id"`
	UserName string  `json:"username"`
	Nick     string  `json:"nickname"`
	Email    *string `json:"email"`
}

func GetUsers(userIDs []string) (map[string]*User, error) {
	once.Do(func() {
		var err error
		tokenAuth, err = ucauth.NewUCTokenAuth(discover.UC(), conf.UCClientID(), conf.UCClientSecret())
		if err != nil {
			logrus.Fatal("failed to NewUCTokenAuth", err)
		}
		client = httpclient.New(httpclient.WithDialerKeepAlive(time.Second * 10))
	})

	var (
		err   error
		token ucauth.OAuthToken
	)
	if token, err = tokenAuth.GetServerToken(false); err != nil {
		return nil, err
	}

	parts := make([]string, len(userIDs))
	for i, ele := range userIDs {
		parts[i] = strutil.Concat("user_id:", ele)
	}
	query := strutil.Join(parts, " OR ")
	var b []*User
	resp, err := client.Get(discover.UC()).
		Path(UserAPI).
		Param("query", query).
		Header("Authorization", strutil.Concat("Bearer ", token.AccessToken)).
		Do().
		JSON(&b)
	if err != nil {
		return nil, err
	}
	if !resp.IsOK() {
		return nil, errors.Errorf("failed to get users, status code: %d", resp.StatusCode())
	}

	users := make(map[string]*User, len(b))
	for _, ele := range b {
		users[strconv.FormatUint(ele.ID, 10)] = ele
	}

	return users, nil
}
