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

package v1alpha2

import (
	configv1alpha2_api "istio.io/client-go/pkg/apis/config/v1alpha2"
	configv1alpha2 "istio.io/client-go/pkg/clientset/versioned/typed/config/v1alpha2"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewConfigClient(addr string) (*configv1alpha2.ConfigV1alpha2Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &configv1alpha2_api.SchemeGroupVersion
	var client rest.Interface
	var err error
	if addr != "" {
		client, err = restclient.NewInetRESTClient(addr, config)
	} else {
		client, err = rest.RESTClientFor(config)
	}
	if err != nil {
		return nil, err
	}
	return configv1alpha2.New(client), nil
}
