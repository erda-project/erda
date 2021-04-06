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

package v1beta1

import (
	secv1beta1_api "istio.io/client-go/pkg/apis/security/v1beta1"
	secv1beta1 "istio.io/client-go/pkg/clientset/versioned/typed/security/v1beta1"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewSecurityClient(addr string) (*secv1beta1.SecurityV1beta1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &secv1beta1_api.SchemeGroupVersion
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
	return secv1beta1.New(client), nil
}
