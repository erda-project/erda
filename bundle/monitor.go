package bundle

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
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
