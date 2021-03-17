package bundle

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// GetOrg get org by id from cmdb.
func (b *Bundle) GetOrg(idOrName interface{}) (*apistructs.OrgDTO, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.OrgFetchResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/orgs/%v", idOrName)).Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}

// ListOrgs 分页查询企业
func (b *Bundle) ListOrgs(req *apistructs.OrgSearchRequest) (*apistructs.PagingOrgDTO, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.OrgSearchResponse
	r, err := hc.Get(host).Path("/api/orgs").
		Header(httputil.InternalHeader, "bundle").
		Param("q", req.Q).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return nil, toAPIError(r.StatusCode(), resp.Error)
	}
	return &resp.Data, nil
}

// GetOrgByDomain 通过域名获取企业
func (b *Bundle) GetOrgByDomain(domain string, userID string) (*apistructs.OrgDTO, error) {
	// TODO: userID should be deprecated
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.OrgGetByDomainResponse
	r, err := hc.Get(host).Path("/api/orgs/actions/get-by-domain").
		Header(httputil.InternalHeader, "bundle").
		Header("User-ID", userID). // TODO: for compatibility
		Param("domain", domain).
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return nil, toAPIError(r.StatusCode(), resp.Error)
	}
	return resp.Data, nil
}
