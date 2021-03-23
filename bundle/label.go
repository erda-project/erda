package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// GetLabel 通过id获取label
func (b *Bundle) GetLabel(id uint64) (*apistructs.ProjectLabel, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var labelResp apistructs.ProjectLabelGetByIDResponseData
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/labels/%d", id)).Header(httputil.InternalHeader, "bundle").
		Do().JSON(&labelResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !labelResp.Success {
		return nil, toAPIError(resp.StatusCode(), labelResp.Error)
	}

	return &labelResp.Data, nil
}
