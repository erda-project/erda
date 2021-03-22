package netportal

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/customhttp"
)

var bdl *bundle.Bundle

func init() {
	bdl = bundle.New(bundle.WithScheduler())
}

// NewNetportalRequest 在NewRequest基础上增加ClusterName参数，完成netportal代理请求
func NewNetportalRequest(clusterName, method, url string, body io.Reader) (*http.Request, error) {
	info, err := bdl.QueryClusterInfo(clusterName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	netportalUrl := info.Get(apistructs.NETPORTAL_URL)
	if netportalUrl != "" {
		url = strings.TrimPrefix(url, "http://")
		url = strings.TrimPrefix(url, "https://")
		url = fmt.Sprintf("%s/%s", netportalUrl, url)
	}
	return customhttp.NewRequest(method, url, body)
}
