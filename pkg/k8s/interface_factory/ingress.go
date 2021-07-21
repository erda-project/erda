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

package interface_factory

import (
	"errors"

	v1 "k8s.io/api/apps/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/client-go/kubernetes"

	erdaextensionsv1beta1 "github.com/erda-project/erda/pkg/k8s/typed/extensions/v1beta1"
	erdav1 "github.com/erda-project/erda/pkg/k8s/typed/networking/v1"
	erdav1beta1 "github.com/erda-project/erda/pkg/k8s/typed/networking/v1beta1"
	"github.com/erda-project/erda/pkg/k8s/union_interface"
)

const IngressKind = "Ingress"

func CreateIngressesHelper(client *kubernetes.Clientset) (union_interface.IngressesHelper, error) {
	exist, err := IsResourceExist(client, IngressKind, extensionsv1beta1.SchemeGroupVersion.String())
	if err != nil {
		return nil, err
	}
	if exist {
		return erdaextensionsv1beta1.NewIngressHelper(client.ExtensionsV1beta1()), nil
	}
	exist, err = IsResourceExist(client, IngressKind, v1beta1.SchemeGroupVersion.String())
	if err != nil {
		return nil, err
	}
	if exist {
		return erdav1beta1.NewIngressHelper(client.NetworkingV1beta1()), nil
	}
	exist, err = IsResourceExist(client, IngressKind, v1.SchemeGroupVersion.String())
	if err != nil {
		return nil, err
	}
	if exist {
		return erdav1.NewIngressHelper(client.NetworkingV1()), nil
	}

	return nil, errors.New("there is no ingress kind")
}
