package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetCloudAccount(accountID uint64) (*apistructs.CloudAccountAllInfo, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var accountResp apistructs.CloudAccountGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/cloud-accounts/%d", accountID)).Header(httputil.InternalHeader, "bundle").Do().JSON(&accountResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !accountResp.Success {
		return nil, toAPIError(resp.StatusCode(), accountResp.Error)
	}

	return &accountResp.Data, nil
}
