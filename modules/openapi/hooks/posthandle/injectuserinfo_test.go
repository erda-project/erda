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

package posthandle

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/pkg/strutil"
)

func TestInjectUserInfo(t *testing.T) {
	testHookUC = func(b *[]User) {
		*b = []User{
			{
				ID:        "1234",
				Name:      "name",
				AvatarURL: "avatar_url",
				Phone:     "123",
			},
			{
				ID:        "2222",
				Name:      "name2",
				AvatarURL: "avatar_url2",
				Phone:     "1232",
			},
			{
				ID:        "333",
				Name:      "name3",
				AvatarURL: "avatar_url3",
				Phone:     "12333",
			},
		}
	}
	r := http.Response{Header: http.Header{}, Body: ioutil.NopCloser(strings.NewReader(`{"userIDs": ["123", "2345"], "data": {"a": "1", "b": 2}}`))}
	assert.Nil(t, InjectUserInfo(&r, false))
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err)
	assert.True(t, strutil.Contains(string(body), "userInfo"))
	assert.True(t, strutil.Contains(string(body), "data"))
}

func TestInjectUserInfo2(t *testing.T) {
	testHookUC = func(b *[]User) {
		*b = []User{
			{
				ID:        "333",
				Name:      "name3",
				AvatarURL: "avatar_url3",
				Phone:     "12333",
			},
		}
	}
	originbody := `{"NOuserIDs":["123","2345"],"data":{"a":"1","b":2}}`
	r := http.Response{Body: ioutil.NopCloser(strings.NewReader(originbody))}
	assert.Nil(t, InjectUserInfo(&r, false))
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err)

	assert.Equal(t, len(originbody), len(string(body)))

}
