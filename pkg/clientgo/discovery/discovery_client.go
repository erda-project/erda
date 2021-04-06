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

package discovery

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/erda-project/erda/pkg/clientgo/restclient"
)

func NewDiscoveryClient(addr string) (*discovery.DiscoveryClient, error) {
	config := restclient.GetDefaultConfig("")
	config.APIPath = ""
	config.GroupVersion = nil
	if config.Timeout == 0 {
		config.Timeout = 32 * time.Second
	}
	codec := runtime.NoopEncoder{Decoder: scheme.Codecs.UniversalDecoder()}
	config.NegotiatedSerializer = serializer.NegotiatedSerializerWrapper(runtime.SerializerInfo{Serializer: codec})
	client, err := restclient.NewInetRESTClient(addr, config)
	if err != nil {
		return nil, err
	}
	return discovery.NewDiscoveryClient(client), nil
}
