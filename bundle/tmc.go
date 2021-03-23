package bundle

import (
	"fmt"
	"net/url"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
)

// PutRuntimeService 部署runtime之后，orchestrator需要将服务域名信息通过此接口提交给hepa
func (b *Bundle) PutRuntimeService(req *apistructs.RuntimeServiceRequest) error {
	host, err := b.urls.Hepa()
	if err != nil {
		return err
	}
	hc := b.hc

	var resp apistructs.Header
	r, err := hc.Put(host).
		Path("/api/gateway/runtime-services").
		JSONBody(req).
		Do().
		JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return toAPIError(r.StatusCode(), resp.Error)
	}
	return nil
}

// DeleteRuntimeService 删除runtime前，调用hepa
func (b *Bundle) DeleteRuntimeService(runtimeID uint64) error {
	host, err := b.urls.Hepa()
	if err != nil {
		return err
	}
	hc := b.hc

	var resp apistructs.Header
	r, err := hc.Delete(host).
		Path(fmt.Sprintf("/api/gateway/runtime-services/%d", runtimeID)).
		Do().
		JSON(&resp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return toAPIError(r.StatusCode(), resp.Error)
	}
	return nil
}

// GetTenantGroupDetails .
func (b *Bundle) GetTenantGroupDetails(tenantGroup string) (*apistructs.TenantGroupDetails, error) {
	host, err := b.urls.TMC()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.TenantGroupDetailsResponse
	r, err := hc.Get(host).
		Path("/api/tmc/tenants/group/details").Param("tenantGroup", tenantGroup).
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

// GetMonitorStatusMetricDetails .
func (b *Bundle) GetMonitorStatusMetricDetails(metricID string) (*apistructs.MonitorStatusMetricDetails, error) {
	host, err := b.urls.TMC()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.MonitorStatusMetricDetailsResponse
	r, err := hc.Get(host).
		Path(fmt.Sprintf("/api/tmc/status/metrics/%s/details", url.QueryEscape(metricID))).
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
