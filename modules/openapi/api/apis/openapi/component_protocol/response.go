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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/hooks/posthandle"
)

func modifyProxyResponse(resp *http.Response) error {
	needInjectUserInfo, err := wrapErdaStyleResponse(resp)
	if err != nil {
		logrus.Errorf("failed to wrap erda style response when modify proxied response of component-protocol: %v", err)
		return err
	}
	if !needInjectUserInfo {
		return nil
	}
	if err := posthandle.InjectUserInfo(resp, true); err != nil {
		logrus.Errorf("failed to inject userinfo when modify proxied response of component-protocol: %v", err)
		return err
	}
	return nil
}

// response .
type response struct {
	Success bool                     `json:"success,omitempty"`
	Data    interface{}              `json:"data,omitempty"`
	UserIDs []string                 `json:"userIDs,omitempty"`
	Err     apistructs.ErrorResponse `json:"err,omitempty"`
}

// wrapErdaStyleResponse return needInjectUserInfo to reduce json unmarshal cost.
func wrapErdaStyleResponse(resp *http.Response) (needInjectUserInfo bool, err error) {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	var bodyjson map[string]interface{}
	if err := json.Unmarshal(content, &bodyjson); err != nil {
		// no op if not json response
		resp.Body = ioutil.NopCloser(bytes.NewReader(content))
		return false, nil
	}
	// construct erda style response
	var erdaResp response
	if resp.StatusCode/100 != 2 {
		// {"code":500,"err":"default protocol not exist, scenario: demo1"}
		erdaResp = response{
			Success: false,
			Data:    nil,
			UserIDs: nil,
			Err: apistructs.ErrorResponse{
				Code: strutil.String(bodyjson["code"]),
				Msg:  strutil.String(bodyjson["err"]),
				Ctx:  nil,
			},
		}
	} else {
		erdaResp = response{
			Success: true,
			Data:    bodyjson,
			UserIDs: nil,
			Err:     apistructs.ErrorResponse{},
		}
	}
	// update to response body
	newbody, err := json.Marshal(&erdaResp)
	if err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewReader(content))
		return false, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewReader(newbody))
	resp.Header["Content-Length"] = []string{fmt.Sprint(len(newbody))}

	// whether need inject user info
	_, haveUserIDs := bodyjson["userIDs"]
	return haveUserIDs, nil
}
