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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/kubernetes/scheme"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

// NewBatchClient creates a new BatchV1Client for the given config.
func NewBatchClient(addr string) (*batchv1client.BatchV1Client, error) {
	var (
		client rest.Interface
		err    error
		config *rest.Config
	)
	if addr != "" {
		config = restclient.GetDefaultConfig("")
		config.GroupVersion = &batchv1.SchemeGroupVersion
		client, err = restclient.NewInetRESTClient(addr, config)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		config.GroupVersion = &batchv1.SchemeGroupVersion
		config.APIPath = "/apis"
		config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
		rest.SetKubernetesDefaults(config)
		client, err = rest.RESTClientFor(config)
	}
	if err != nil {
		return nil, err
	}
	return batchv1client.New(client), nil
}
