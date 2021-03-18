package bundle

import (
	"fmt"
	"io"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// CollectLogs 收集日志
func (b *Bundle) CollectLogs(source string, body io.Reader) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc
	resp, err := hc.Post(host).Path(strutil.Concat("/collect/logs/", source)).
		Header("Internal-Client", "bundle").
		RawBody(body).
		Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("failed to call monitor status %d", resp.StatusCode()))
	}
	return nil
}

// CollectMetrics 收集指标
func (b *Bundle) CollectMetrics(metrics *apistructs.Metrics) error {
	host, err := b.urls.Collector()
	if err != nil {
		return err
	}
	hc := b.hc
	resp, err := hc.Post(host).Path("/collect/metrics").
		Header("Internal-Client", "bundle").JSONBody(&metrics).Do().DiscardBody()
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(fmt.Errorf("failed to call monitor status %d", resp.StatusCode()))
	}
	return nil
}
