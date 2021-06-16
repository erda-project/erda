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

package nacos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/msp/configcenter/pb"
	"github.com/erda-project/erda/pkg/netportal"
)

// SearchMode .
type SearchMode string

var (
	// SearchModeBlur .
	SearchModeBlur SearchMode = "BLUR"
	// SearchModeAccurate .
	SearchModeAccurate SearchMode = "ACCURATE"
)

// Adapter .
type Adapter struct {
	ClusterName string
	Addr        string
	User        string
	Password    string
}

// NewAdapter .
func NewAdapter(clusterName, addr, user, password string) *Adapter {
	return &Adapter{
		ClusterName: clusterName,
		Addr:        addr,
		User:        user,
		Password:    password,
	}
}

// SearchResponse .
type SearchResponse struct {
	Total       int64         `json:"json:"totalCount"`
	Pages       int64         `json:"json:"pagesAvailable"`
	ConfigItems []*ConfigItem `json:"json:"pageItems"`
}

// ConfigItem .
type ConfigItem struct {
	DataID  string
	Group   string
	Content string
}

// ToConfigCenterGroups .
func (s *SearchResponse) ToConfigCenterGroups() *pb.Groups {
	return &pb.Groups{}
}

// SearchConfig .
func (a *Adapter) SearchConfig(mode SearchMode, tenantName, groupName, dataID string, page, pageSize int) (*SearchResponse, error) {
	auth, err := a.loginNacos()
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Set("search", strings.ToLower(string(mode)))
	params.Set("dataId", dataID)
	params.Set("group", groupName)
	params.Set("tenant", tenantName)
	params.Set("pageNo", strconv.Itoa(page))
	params.Set("pageSize", strconv.Itoa(pageSize))
	req, err := a.newRequest(http.MethodGet, "/nacos/v1/cs/configs", params, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	var body SearchResponse
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
	}
	return &body, nil
}

// SaveConfig .
func (a *Adapter) SaveConfig(tenantName, groupName, dataID, content string) error {
	auth, err := a.loginNacos()
	if err != nil {
		return err
	}
	params := url.Values{}
	params.Set("dataId", dataID)
	params.Set("group", groupName)
	params.Set("tenant", tenantName)
	params.Set("content", content)
	req, err := a.newRequest(http.MethodGet, "/nacos/v1/cs/configs", params, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices < resp.StatusCode {
		return fmt.Errorf("nacos response status error: %d %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// loginNacos .
func (a *Adapter) loginNacos() (string, error) {
	params := url.Values{}
	params.Set("username", a.User)
	params.Set("password", a.Password)
	req, err := a.newRequest(http.MethodPost, "/nacos/v1/auth/login", params, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body struct {
			AccessToken string `json:"accessToken"`
			Data        string `json:"data"`
		}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return "", err
		}
		if len(body.AccessToken) > 0 {
			return body.AccessToken, nil
		} else if len(body.Data) > 0 {
			return body.Data, nil
		}
	}
	return "", nil
}

func (a *Adapter) newRequest(method, path string, params url.Values, body io.Reader) (*http.Request, error) {
	var ustr string
	if len(params) > 0 {
		ustr = fmt.Sprintf("%s%s?%s", a.Addr, path, params.Encode())
	} else {
		ustr = fmt.Sprintf("%s%s", a.Addr, path)
	}
	if len(a.ClusterName) > 0 {
		return netportal.NewNetportalRequest(a.ClusterName, method, ustr, body)
	}
	return http.NewRequest(method, ustr, body)
}
