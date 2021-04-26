package bundle

import (
	"fmt"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
	"strconv"
)

func (b *Bundle) GetWorkbenchData(userID string, req apistructs.WorkbenchRequest) (*apistructs.WorkbenchResponse, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.WorkbenchResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/workbench/actions/list")).
		Header(httputil.UserHeader, userID).
		Param("pageNo", strconv.FormatInt(int64(req.PageNo), 10)).
		Param("pageSize", strconv.FormatInt(int64(req.PageSize), 10)).
		Param("issueSize", strconv.FormatInt(int64(req.IssueSize), 10)).
		Param("orgID", strconv.FormatInt(int64(req.OrgID), 10)).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrGetWorkBenchData.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp, nil
}

func (b *Bundle) GetIssuesForWorkbench(req apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.IssuePagingResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/workbench/issues/list")).
		Header(httputil.UserHeader, req.UserID).
		Params(req.UrlQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrGetWorkBenchData.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp, nil
}