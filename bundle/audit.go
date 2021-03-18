package bundle

import (
	"bytes"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// CreateAuditEvent 创建审计事件
func (b *Bundle) CreateAuditEvent(audits *apistructs.AuditCreateRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Post(host).Path("/api/audits/actions/create").
		Header(httputil.InternalHeader, "bundle").JSONBody(&audits).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to create Audit, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}

// BatchCreateAuditEvent 批量创建审计事件
func (b *Bundle) BatchCreateAuditEvent(audits *apistructs.AuditBatchCreateRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Post(host).Path("/api/audits/actions/batch-create").
		Header(httputil.InternalHeader, "bundle").JSONBody(&audits).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to create Audit, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}
	return nil
}
