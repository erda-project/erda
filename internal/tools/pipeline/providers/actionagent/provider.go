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

package actionagent

import (
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/jsonstore/etcd"
)

type config struct{}

type provider struct {
	Cfg *config
	Log logs.Logger

	bdl             *bundle.Bundle
	accessibleCache jsonstore.JsonStore
	etcdctl         *etcd.Store
}

func (p *provider) Init(ctx servicehub.Context) error {
	js, err := jsonstore.New()
	if err != nil {
		return fmt.Errorf("failed to init jsonstore, err: %v", err)
	}
	p.accessibleCache = js
	etcdClient, err := etcd.New()
	if err != nil {
		return fmt.Errorf("failed to init etcd client, err: %v", err)
	}
	p.etcdctl = etcdClient
	p.bdl = bundle.New(bundle.WithCMDB(), bundle.WithCMP())
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("actionagent", &servicehub.Spec{
		Services:     []string{"actionagent"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "pipeline action agent",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
