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
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	policyv1beta1client "k8s.io/client-go/kubernetes/typed/policy/v1beta1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewPolicyClient creates a new PolicyV1beta1 for the given addr.
func NewPolicyClient(addr string) (*policyv1beta1client.PolicyV1beta1Client, error) {
	config := restclient.GetDefaultConfig("")
	config.GroupVersion = &policyv1beta1.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return policyv1beta1client.New(client), nil
}
