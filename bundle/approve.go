package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) CreateApprove(req *apistructs.ApproveCreateRequest) (*apistructs.ApproveDTO, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var createResp apistructs.ApproveCreateResponse
	resp, err := hc.Post(host).Path("/api/approves").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(req).Do().JSON(&createResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return nil, toAPIError(resp.StatusCode(), createResp.Error)
	}

	return &createResp.Data, nil
}
