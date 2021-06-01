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

package restclient

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/clusterdialer"
)

func GetDefaultConfig(apiPath string) *rest.Config {
	if apiPath == "" {
		apiPath = "/apis"
	}
	return &rest.Config{
		APIPath:     apiPath,
		QPS:         1000,
		Burst:       100,
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(1000, 100),
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		UserAgent: rest.DefaultKubernetesUserAgent(),
	}
}

func GetDialerRestConfig(clusterName string, config *apistructs.ManageConfig) (*rest.Config, error) {
	rc, err := GetRestConfig(config)
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

func GetRestConfig(config *apistructs.ManageConfig) (*rest.Config, error) {
	if config.Address == "" {
		return nil, errors.New("address is empty")
	}

	rc := &rest.Config{
		Host:    config.Address,
		APIPath: "/apis",
		QPS:     1000,
		Burst:   100,
		ContentConfig: rest.ContentConfig{
			NegotiatedSerializer: scheme.Codecs.WithoutConversion(),
		},
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: config.Insecure,
		},
		UserAgent:   rest.DefaultKubernetesUserAgent(),
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(1000, 100),
	}

	if !config.Insecure {
		if config.CaData == "" {
			return nil, fmt.Errorf("the ca data is empty which insecure is disabled")
		}

		caBytes, err := base64.StdEncoding.DecodeString(config.CaData)
		if err != nil {
			return nil, err
		}

		rc.TLSClientConfig.CAData = caBytes
	}

	switch config.Type {
	case apistructs.ManageToken, apistructs.ManageProxy:
		if config.Token == "" {
			return nil, fmt.Errorf("token/proxy manage type must provide available token")
		}
		rc.BearerToken = config.Token
	case apistructs.ManageCert:
		if config.CertData == "" || config.KeyData == "" {
			return nil, fmt.Errorf("cert manage tpye must provide available cert data and key data")
		}
		certBytes, err := base64.StdEncoding.DecodeString(config.CertData)
		if err != nil {
			return nil, err
		}

		keyBytes, err := base64.StdEncoding.DecodeString(config.KeyData)
		if err != nil {
			return nil, err
		}

		rc.TLSClientConfig.CertData = certBytes
		rc.TLSClientConfig.KeyData = keyBytes
	case apistructs.ManageInet:
		fallthrough
	default:
		return nil, fmt.Errorf("type not support, use clientgo.New(addr) instead")
	}

	return rc, nil
}
