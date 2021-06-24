package manager

import (
	"net/http"
	"strconv"

	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/pkg/errors"
)

type USERID string

// Invalid return UserID is valid or invalid
func (uid USERID) Invalid() bool {
	return string(uid) == ""
}

func GetOrgID(req *http.Request) (uint64, error) {
	// get organization id
	orgIDStr := req.URL.Query().Get("orgId")
	if orgIDStr == "" {
		orgIDStr = req.Header.Get(httputil.OrgHeader)
		if orgIDStr == "" {
			return 0, errors.Errorf("invalid param, orgId is empty")
		}
	}
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return 0, errors.Errorf("invalid param, orgId is invalid")
	}
	return orgID, nil
}
