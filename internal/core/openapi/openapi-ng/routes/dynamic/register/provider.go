// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dynamic

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes"
)

var (
	_ routes.Register = (*provider)(nil)
)

func init() {
	servicehub.Register("openapi-dynamic-register", &servicehub.Spec{
		Services:   []string{"openapi-dynamic-register.client"},
		ConfigFunc: func() interface{} { return new(Config) },
		Creator:    func() servicehub.Provider { return new(provider) },
	})
}

type provider struct {
	Cfg  *Config
	L    logs.Logger
	Etcd *clientv3.Client `autowired:"etcd-client" optional:"true"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Cfg.Prefix = filepath.Clean("/" + p.Cfg.Prefix)
	return nil
}

func (p *provider) Register(apiProxy *routes.APIProxy) error {
	if p.Etcd == nil {
		p.L.Warn("ETCDv3 client is not autowired, no route will be registered")
		return nil
	}
	if err := apiProxy.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	key := fmt.Sprintf("%s/%s %s", p.Cfg.Prefix, apiProxy.Method, apiProxy.Path)
	data, err := json.MarshalIndent(apiProxy, "\t", "  ")
	if err != nil {
		return err
	}
	p.L.Infof("ETCDv3 put, endpoints: %v,\n\tkey: %s\n\tdata: %s\n", p.Etcd.Endpoints(), key, string(data))
	_, err = p.Etcd.Put(ctx, key, string(data))
	return err
}

type Config struct {
	// Prefix is the key's prefix that openapi stored in ETCD.
	// It defaults "/openapi/apis".
	// You should be careful to make it consistent with the key's prefix that the openapi module uses to get from the ETCD
	// if you configure it.
	Prefix string `file:"prefix" default:"/openapi/apis"`
}
