// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	credentialpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cluster-dialer/config"
)

const (
	SkipAllKey = "all"
)

type Option func(authorizer *Authorizer)

type Authorizer struct {
	Credential credentialpb.AccessKeyServiceServer
	cfg        *config.Config
}

func New(opts ...Option) *Authorizer {
	a := Authorizer{}

	for _, opt := range opts {
		opt(&a)
	}

	return &a
}

// WithCredentialClient with core-services credential client
func WithCredentialClient(credential credentialpb.AccessKeyServiceServer) Option {
	return func(a *Authorizer) {
		a.Credential = credential
	}
}

// WithConfig with dialer config
func WithConfig(cfg *config.Config) Option {
	return func(a *Authorizer) {
		a.cfg = cfg
	}
}

// Authorizer auth before connection create
func (a *Authorizer) Authorizer(req *http.Request) (string, bool, error) {
	// inner proxy not need auth
	if req.URL.Path == "/clusterdialer" {
		return "proxy", true, nil
	}

	clusterKey := req.Header.Get("X-Erda-Cluster-Key")
	authInfo := req.Header.Get("Authorization")

	// Check header param
	if clusterKey == "" || authInfo == "" {
		return "", false, fmt.Errorf("cluster key or auth info is empty")
	}

	if !a.CheckNeedAuth(clusterKey) {
		// doesn't need auth, skip
		logrus.Infof("cluster %s in white list, skip auth", clusterKey)
		return clusterKey, true, nil
	}

	logrus.Debugf("get auth info: %s", authInfo)

	// Doesn't need cache now
	akSkResp, err := a.Credential.QueryAccessKeys(context.Background(), &credentialpb.QueryAccessKeysRequest{
		AccessKey: authInfo,
		Scope:     apistructs.CMPClusterScope,
		ScopeId:   clusterKey,
		Status:    credentialpb.StatusEnum_ACTIVATE,
	})
	if err != nil {
		logrus.Errorf("get key pair error: %v", err)
		return "", false, fmt.Errorf("server internal error")
	}

	if akSkResp.Total == 0 {
		return "", false, fmt.Errorf("auth failed, access key: %s", authInfo)
	} else if akSkResp.Total > 1 {
		logrus.Errorf("duplidate key, please check or rest, access key: %s", authInfo)
		return "", false, fmt.Errorf("auth server error, duplidate data")
	}

	logrus.Infof("auth success, resp info: %+v", akSkResp)

	return clusterKey, true, nil
}

// CheckNeedAuth check auth
func (a *Authorizer) CheckNeedAuth(clusterKey string) bool {
	// TODO: get auth or not result from cmp/cluster-manager
	wlStr := a.cfg.AuthWhitelist

	// Match all clusters
	if wlStr == SkipAllKey {
		return false
	} else if wlStr == "" {
		return true
	}

	wl := strings.Split(strings.Replace(wlStr, " ", "", -1), ",")

	sort.Strings(wl)
	i := sort.SearchStrings(wl, clusterKey)
	if i < len(wl) && wl[i] == clusterKey {
		return false
	}

	return true
}
