package bundle

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

func (b *Bundle) ListTestPlanCaseRel(testCaseIDs []uint64) ([]apistructs.TestPlanCaseRel, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	urlQueryStrings := make(map[string][]string)
	for _, tcID := range testCaseIDs {
		urlQueryStrings["id"] = append(urlQueryStrings["id"], fmt.Sprintf("%d", tcID))
	}

	var listResp apistructs.TestPlanCaseRelListResponse
	resp, err := hc.Get(host).Path("/api/testplans/testcase-relations/actions/internal-list").
		Header(httputil.InternalHeader, "bundle").
		Params(urlQueryStrings).
		Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return listResp.Data, nil
}
