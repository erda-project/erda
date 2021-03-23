package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

// CreateMBox 创建站内信记录
func (b *Bundle) CreateMBox(request *apistructs.CreateMBoxRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var getResp apistructs.CreateMBoxResponse
	resp, err := hc.Post(host).Path("/api/mboxs").JSONBody(request).
		Do().JSON(&getResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return toAPIError(resp.StatusCode(), getResp.Error)
	}
	return nil
}
