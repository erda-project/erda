package user

import (
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// GetOrgID 从 http request 的 header 中读取 org id.
func GetOrgID(r *http.Request) (uint64, error) {
	v := r.Header.Get("ORG-ID")

	orgID, err := strconv.ParseUint(v, 10, 64)
	if err == nil {
		return orgID, nil
	}

	return 0, errors.Errorf("invalid org id")
}
