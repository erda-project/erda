package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
)

// PushLog 推日志
func (b *Bundle) PushLog(req *apistructs.LogPushRequest) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc

	resp, err := hc.Post(host, httpclient.RetryErrResp).
		Path("/collect/logs/job").
		JSONBody(req.Lines).
		Header("Content-Type", "application/json").
		Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to pushLog, error is %d", resp.StatusCode()))
	}
	return nil
}
