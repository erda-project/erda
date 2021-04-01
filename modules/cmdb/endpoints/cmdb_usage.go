package endpoints

import (
	"net/http"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/cmdb/model"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) getOrgByRequest(r *http.Request) (*model.Org, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return nil, errors.Errorf("missing org id header")
	}
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return nil, errors.Errorf("invalid org id")
	}
	org, err := e.org.Get(orgID)
	if err != nil {
		return nil, err
	}
	return org, nil
}
