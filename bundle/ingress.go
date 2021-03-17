package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// Create or update component ingress
func (b *Bundle) CreateOrUpdateComponentIngress(req apistructs.ComponentIngressUpdateRequest) error {
	host, err := b.urls.Hepa()
	if err != nil {
		return err
	}
	var fetchResp apistructs.ComponentIngressUpdateResponse
	resp, err := b.hc.Put(host).Path("/api/gateway/component-ingress").Header(httputil.InternalHeader, "bundle").JSONBody(req).Do().JSON(&fetchResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return toAPIError(resp.StatusCode(), fetchResp.Error)
	}
	return nil
}
