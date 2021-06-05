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

package config

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/clusterdialer"
)

// ParseManageConfig parse manage config to rest.Config
func ParseManageConfig(clusterName string, c *apistructs.ManageConfig) (*rest.Config, error) {
	if c == nil {
		return nil, fmt.Errorf("empty manage config")
	}

	switch c.Type {
	case apistructs.ManageToken, apistructs.ManageCert:
		return GetRestConfig(c)
	case apistructs.ManageProxy:
		fallthrough
	default:
		return GetDialerRestConfig(clusterName, c)
	}
}

// GetDialerRestConfig get cluster dialer rest.Config
func GetDialerRestConfig(clusterName string, c *apistructs.ManageConfig) (*rest.Config, error) {
	rc, err := GetRestConfig(c)
	if err != nil {
		return nil, err
	}

	rc.TLSClientConfig.NextProtos = []string{"http/1.1"}
	rc.UserAgent = rest.DefaultKubernetesUserAgent() + " cluster " + clusterName
	rc.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
		if ht, ok := rt.(*http.Transport); ok {
			ht.DialContext = clusterdialer.DialContext(clusterName)
		}
		return rt
	}

	return rc, nil
}

// GetRestConfig get rest.Config with manage config
func GetRestConfig(c *apistructs.ManageConfig) (*rest.Config, error) {
	// If not provide api-server address, use in-cluster rest config
	if c.Address == "" {
		return rest.InClusterConfig()
	}

	rc := &rest.Config{
		Host:    c.Address,
		APIPath: "/apis",
		QPS:     1000,
		Burst:   100,
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		TLSClientConfig: rest.TLSClientConfig{},
		UserAgent:       rest.DefaultKubernetesUserAgent(),
		RateLimiter:     flowcontrol.NewTokenBucketRateLimiter(1000, 100),
	}

	// If ca data is empty, the certificate is not validated
	if c.CaData == "" {
		rc.TLSClientConfig.Insecure = true
	} else {
		caBytes, err := base64.StdEncoding.DecodeString(c.CaData)
		if err != nil {
			return nil, err
		}

		rc.TLSClientConfig.CAData = caBytes
	}

	// If token is not empty, use token firstly, else use certificate
	if c.Token != "" {
		rc.BearerToken = c.Token
	} else {
		if c.CertData == "" || c.KeyData == "" {
			return nil, fmt.Errorf("must provide available cert data and key data")
		}

		certBytes, err := base64.StdEncoding.DecodeString(c.CertData)
		if err != nil {
			return nil, err
		}

		keyBytes, err := base64.StdEncoding.DecodeString(c.KeyData)
		if err != nil {
			return nil, err
		}

		rc.TLSClientConfig.CertData = certBytes
		rc.TLSClientConfig.KeyData = keyBytes
	}

	return rc, nil
}
