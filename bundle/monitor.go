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

package bundle

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetMonitorAlertByID .
func (b *Bundle) GetMonitorAlertByID(id int64) (*apistructs.Alert, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetMonitorAlertResponse
	resp, err := hc.Get(host).Path("/api/alerts/"+strconv.FormatInt(id, 10)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

// GetMonitorAlertByScope .
func (b *Bundle) GetMonitorAlertByScope(scope, scopeID string) (*apistructs.Alert, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var response struct {
		apistructs.Header
		Data struct {
			List  []*apistructs.Alert `json:"list"`
			Total int64               `json:"total"`
		} `json:"data"`
	}
	url.QueryEscape(scope)
	resp, err := hc.Get(host).Path(
		fmt.Sprintf("/api/alerts?scope=%s&scopeID=%s&pageSize=1&pageNo=1",
			url.QueryEscape(scope), url.QueryEscape(scopeID))).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	if len(response.Data.List) <= 0 {
		return nil, nil
	}
	return response.Data.List[0], nil
}

// GetMonitorCustomAlertByID .
func (b *Bundle) GetMonitorCustomAlertByID(id int64) (*apistructs.Alert, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetMonitorAlertResponse
	resp, err := hc.Get(host).Path("/api/customize/alerts/"+strconv.FormatInt(id, 10)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

// GetMonitorCustomAlertByScope .
func (b *Bundle) GetMonitorCustomAlertByScope(scope, scopeID string) (*apistructs.Alert, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var response struct {
		apistructs.Header
		Data struct {
			List  []*apistructs.Alert `json:"list"`
			Total int64               `json:"total"`
		} `json:"data"`
	}
	url.QueryEscape(scope)
	resp, err := hc.Get(host).Path(
		fmt.Sprintf("/api/customize/alerts?scope=%s&scopeID=%s&pageSize=1&pageNo=1",
			url.QueryEscape(scope), url.QueryEscape(scopeID))).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&response)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !response.Success {
		return nil, toAPIError(resp.StatusCode(), response.Error)
	}
	if len(response.Data.List) <= 0 {
		return nil, nil
	}
	return response.Data.List[0], nil
}

// GetMonitorReportTasksByID .
func (b *Bundle) GetMonitorReportTasksByID(id int64) (*apistructs.ReportTask, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetMonitorReportTaskResponse
	resp, err := hc.Get(host).Path("/api/org/report/tasks/"+strconv.FormatInt(id, 10)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

func (b *Bundle) RegisterConfig(desc string, configList []apistructs.MonitorConfig) error {
	host, err := b.urls.Monitor()
	if err != nil {
		return err
	}
	hc := b.hc

	var response struct {
		apistructs.Header
	}

	resp, err := hc.Put(host).
		Path(fmt.Sprintf("/api/config/register?desc=%s", desc)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(configList).
		Do().
		JSON(&response)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() || !response.Success {
		return toAPIError(resp.StatusCode(), response.Error)
	}

	return nil
}
