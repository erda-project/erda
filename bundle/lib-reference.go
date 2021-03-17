package bundle

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// PublisherItemRefered 根据发布内容 id 查看是否被库应用引用
func (b *Bundle) PublisherItemRefered(libID uint64) (uint64, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var libReferenceListResp apistructs.LibReferenceListResponse
	resp, err := hc.Get(host).Path("/api/lib-references").
		Param("libID", strconv.FormatUint(libID, 10)).
		Header("Accept", "application/json").
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&libReferenceListResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !libReferenceListResp.Success {
		return 0, toAPIError(resp.StatusCode(), libReferenceListResp.Error)
	}

	return libReferenceListResp.Data.Total, nil
}
