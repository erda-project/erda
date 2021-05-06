// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package apitestsv2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/customhttp"
	"github.com/erda-project/erda/pkg/httpclientutil"
	"github.com/erda-project/erda/pkg/strutil"
)

const k8sServiceSuffix = ".svc.cluster.local"

func handleCustomNetportalRequest(apiReq *apistructs.APIRequestInfo, netportalOpt *netportalOption) (*http.Request, error) {
	useNetportal := useNetportal(apiReq.URL, netportalOpt)
	if !useNetportal {
		return http.NewRequest(apiReq.Method, apiReq.URL, nil)
	}
	// use netportal
	apiReq.URL = strutil.Concat(netportalOpt.url, "/", httpclientutil.RmProto(apiReq.URL))

	customReq, err := customhttp.NewRequest(apiReq.Method, apiReq.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to handle custom netportal request, err: %v", err)
	}
	for k, values := range customReq.Header {
		for _, v := range values {
			apiReq.Headers.Add(k, v)
		}
	}

	return customReq, nil
}

// useNetportal return true or false to represent use netportal or not.
func useNetportal(inputURL string, netportalOpt *netportalOption) bool {
	// cannot use if no netportal url
	if netportalOpt == nil || netportalOpt.url == "" {
		return false
	}
	// only host have k8s service suffix will use netportal
	r, err := url.ParseRequestURI(inputURL)
	if err != nil {
		logrus.Errorf("failed to parse apitest url, url: %s, err: %v", inputURL, err)
		// if err, not use netportal
		return false
	}
	hostport := r.Host
	ss := strings.SplitN(hostport, ":", 2)
	host := ss[0]
	// doesn't have k8s svc suffix, do not use nerportal
	if !strings.HasSuffix(host, k8sServiceSuffix) {
		return false
	}
	// if parsed k8s namespace is blacklist, do not use netportal
	ns := getK8sNamespace(host)
	inBlacklist := strutil.Exist(netportalOpt.blacklistOfK8sNamespaceAccess, ns)
	if inBlacklist {
		return false
	}
	return true
}

func getK8sNamespace(k8sHost string) string {
	hostWithoutSuf := strings.TrimSuffix(k8sHost, k8sServiceSuffix)
	ss := strings.Split(hostWithoutSuf, ".")
	ns := ss[len(ss)-1]
	return ns
}
