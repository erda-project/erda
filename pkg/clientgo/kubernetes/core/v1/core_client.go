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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewCoreClient creates a new CoreV1Client for the given addr.
func NewCoreClient(addr string) (*corev1client.CoreV1Client, error) {
	config := restclient.GetDefaultConfig("/api")
	config.GroupVersion = &corev1.SchemeGroupVersion
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return corev1client.New(client), nil
}
