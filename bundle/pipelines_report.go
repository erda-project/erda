package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) GetPipelineReportSet(pipelineID uint64, types []string) (*apistructs.PipelineReportSet, error) {
	host, err := b.urls.Pipeline()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var pipelineResp apistructs.PipelineReportSetGetResponse
	req := hc.Get(host).Path(fmt.Sprintf("/api/pipeline-reportsets/%d", pipelineID))
	if types != nil {
		for _, v := range types {
			req.Param("type", v)
		}
	}
	httpResp, err := req.Header(httputil.InternalHeader, "bundle").
		Do().JSON(&pipelineResp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}

	if !httpResp.IsOK() || !pipelineResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), pipelineResp.Error)
	}
	return pipelineResp.Data, nil
}
