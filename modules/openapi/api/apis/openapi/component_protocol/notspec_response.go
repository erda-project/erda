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

package component_protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/cp/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/hooks/posthandle"
)

const (
	proxyErrorCode = "Proxy Error"
)

func modifyProxyResponse(proxyConfig types.ProxyConfig) func(*http.Response) error {
	return func(resp *http.Response) error {
		logrus.Infof("[DEBUG] start wrap erda style resp at %s", time.Now().Format(time.StampNano))
		if err := wrapErdaStyleResponse(proxyConfig, resp); err != nil {
			logrus.Errorf("failed to wrap erda style response when modify proxied response of component-protocol: %v", err)
			return err
		}
		logrus.Infof("[DEBUG] end wrap erda style resp at %s", time.Now().Format(time.StampNano))
		return nil
	}
}

// response .
type response struct {
	Success  bool                           `json:"success,omitempty"`
	Data     interface{}                    `json:"data,omitempty"`
	Err      apistructs.ErrorResponse       `json:"err,omitempty"`
	UserIDs  []string                       `json:"userIDs,omitempty"`
	UserInfo map[string]apistructs.UserInfo `json:"userInfo,omitempty"`
}

type cpErrResponse struct {
	Code int    `json:"code,omitempty"`
	Err  string `json:"err,omitempty"`
}

// wrapErdaStyleResponse wrap response by erda response.
func wrapErdaStyleResponse(proxyConfig types.ProxyConfig, resp *http.Response) (wErr error) {
	if resp.Header.Get("nowrap") == "true" {
		logrus.Info("resp has nowrap header, skip wrap erda style response")
		return
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			resp.Body = ioutil.NopCloser(bytes.NewReader(content))
			wErr = fmt.Errorf("err: %v, responseBody: %s", r, string(content))
		}
		resp.Header.Set("Content-Type", "application/json")
	}()

	// construct erda style response
	var erdaResp response
	if resp.StatusCode/100 != 2 {
		var cpErrResp cpErrResponse
		if err := json.Unmarshal(content, &cpErrResp); err != nil {
			panic(err)
		}
		erdaResp = response{
			Success: false,
			Err: apistructs.ErrorResponse{
				Code: proxyErrorCode + ": " + strutil.String(cpErrResp.Code),
				Msg:  cpErrResp.Err,
			},
		}
	} else {
		var renderResp pb.RenderResponse
		if err := json.Unmarshal(content, &renderResp); err != nil {
			panic(err)
		}
		erdaResp = response{
			Success: true,
			Data:    &renderResp,
			Err:     apistructs.ErrorResponse{},
		}
		// whether need inject user info
		if renderResp.Protocol != nil && len(renderResp.Protocol.GlobalState) > 0 {
			userIDsValue, ok := renderResp.Protocol.GlobalState[cptype.GlobalInnerKeyUserIDs.String()]
			if ok {
				var userIDs []string
				if err := cputil.ObjJSONTransfer(userIDsValue, &userIDs); err != nil {
					panic(err)
				}
				userIDs = strutil.DedupSlice(userIDs, true)
				userInfos, err := posthandle.GetUsers(userIDs, true)
				if err != nil {
					return err
				}
				// inject to response body
				erdaResp.UserIDs = userIDs
				erdaResp.UserInfo = userInfos
			}
		}
	}

	// update response body
	newErdaBody, err := json.Marshal(erdaResp)
	if err != nil {
		panic(err)
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(newErdaBody))
	resp.Header.Set("Content-Length", fmt.Sprint(len(newErdaBody)))

	return nil
}
