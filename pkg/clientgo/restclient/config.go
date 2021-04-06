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
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
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
