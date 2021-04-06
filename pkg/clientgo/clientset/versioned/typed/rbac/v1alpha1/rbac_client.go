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

package v1alpha1

import (
	rbacv1alpha1_api "istio.io/client-go/pkg/apis/rbac/v1alpha1"
	rbacv1alpha1 "istio.io/client-go/pkg/clientset/versioned/typed/rbac/v1alpha1"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewRBACClient(addr string) (*rbacv1alpha1.RbacV1alpha1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &rbacv1alpha1_api.SchemeGroupVersion
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
	return rbacv1alpha1.New(client), nil
}
