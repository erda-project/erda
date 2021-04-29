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

package ucauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/pkg/httpclient"
)

type OryKratosSession struct {
	ID       string            `json:"id"`
	Active   bool              `json:"active"`
	Identity OryKratosIdentity `json:"identity"`
}

type OryKratosIdentity struct {
	ID       USERID                  `json:"id"`
	SchemaID string                  `json:"schema_id"`
	Traits   OryKratosIdentityTraits `json:"traits"`
}

type OryKratosIdentityTraits struct {
	Email string                      `json:"email"`
	Name  OryKratosIdentityTraitsName `json:"name"`
}

type OryKratosIdentityTraitsName struct {
	First string `json:"first"`
	Last  string `json:"last"`
}

func nameConversion(name OryKratosIdentityTraitsName) string {
	// TODO: eastern name vs western name
	return name.Last + name.First
}

func identityToUser(i OryKratosIdentity) User {
	return User{
		ID:    string(i.ID),
		Nick:  nameConversion(i.Traits.Name),
		Email: i.Traits.Email,
	}
}

func identityToUserInfo(i OryKratosIdentity) UserInfo {
	return userToUserInfo(identityToUser(i))
}

func userToUserInfo(u User) UserInfo {
	return UserInfo{
		ID:        USERID(u.ID),
		Email:     u.Email,
		Phone:     u.Phone,
		AvatarUrl: u.AvatarURL,
		UserName:  u.Name,
		NickName:  u.Nick,
		Enabled:   true,
	}
}

func userToUserInPaging(u User) userInPaging {
	return userInPaging{
		Id:       u.ID,
		Avatar:   u.AvatarURL,
		Username: u.Name,
		Nickname: u.Nick,
		Mobile:   u.Phone,
		Email:    u.Email,
		Enabled:  true,
	}
}

func whoami(kratosPublicAddr string, sessionID string) (UserInfo, error) {
	var buf bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(kratosPublicAddr).
		Cookie(&http.Cookie{
			Name:  "ory_kratos_session",
			Value: sessionID,
		}).
		Path("/sessions/whoami").
		Do().Body(&buf)
	if err != nil {
		return UserInfo{}, err
	}
	if !r.IsOK() {
		return UserInfo{}, fmt.Errorf("bad session")
	}
	var i OryKratosSession
	if err := json.Unmarshal(buf.Bytes(), &i); err != nil {
		return UserInfo{}, err
	}
	//return r.ResponseHeader("X-Kratos-Authenticated-Identity-Id"), nil
	return identityToUserInfo(i.Identity), nil
}

func getUserByID(kratosPrivateAddr string, userID string) (*User, error) {
	var buf bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(kratosPrivateAddr).
		Path("/identities/" + userID).
		Do().Body(&buf)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("bad session")
	}
	var i OryKratosIdentity
	if err := json.Unmarshal(buf.Bytes(), &i); err != nil {
		return nil, err
	}
	u := identityToUser(i)
	return &u, nil
}

func getUserByIDs(kratosPrivateAddr string, userIDs []string) ([]User, error) {
	var users []User
	for _, id := range userIDs {
		u, err := getUserByID(kratosPrivateAddr, id)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, nil
}

func getUserByKey(kratosPrivateAddr string, key string) ([]User, error) {
	p := 1
	size := 1000
	cnt := 0
	var users []User
	for {
		ul, err := getUserPage(kratosPrivateAddr, p, size)
		if err != nil {
			return nil, err
		}
		for _, u := range ul {
			if strings.Contains(u.Email, key) {
				users = append(users, u)
				cnt++
			}
		}
		if cnt >= 10 {
			return users, nil
		}
		p++
		if p > 100 {
			return users, nil
		}
	}
	return nil, nil
}

func getUserPage(kratosPrivateAddr string, page, perPage int) ([]User, error) {
	var buf bytes.Buffer
	r, err := httpclient.New(httpclient.WithCompleteRedirect()).
		Get(kratosPrivateAddr).
		Path("/identities").
		Param("page", fmt.Sprintf("%d", page)).
		Param("per_page", fmt.Sprintf("%d", perPage)).
		Do().Body(&buf)
	if err != nil {
		return nil, err
	}
	if !r.IsOK() {
		return nil, fmt.Errorf("bad session")
	}
	var i []OryKratosIdentity
	if err := json.Unmarshal(buf.Bytes(), &i); err != nil {
		return nil, err
	}
	var users []User
	for _, u := range i {
		users = append(users, identityToUser(u))
	}
	return users, nil
}
