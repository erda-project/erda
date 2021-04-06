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

package v1alpha3

import (
	netv1alpha3_api "istio.io/client-go/pkg/apis/networking/v1alpha3"
	netv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewNetworkingClient creates a new NetworkingV1alpha3 for the given addr.
func NewNetworkingClient(addr string) (*netv1alpha3.NetworkingV1alpha3Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &netv1alpha3_api.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return netv1alpha3.New(client), nil
}
