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
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/common/pb"
)

func init() {
	servicehub.Register("openapi-dynamic-register", &servicehub.Spec{
		Services:   []string{"openapi-dynamic-register.client"},
		ConfigFunc: func() interface{} { return new(Config) },
		Creator:    func() servicehub.Provider { return new(provider) },
	})
}

type Register interface {
	Register(route *Route) error
}

type provider struct {
	Cfg  *Config
	Etcd *clientv3.Client `autowired:"etcd-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Cfg.Prefix = filepath.Clean("/" + p.Cfg.Prefix)
	return nil
}

func (p *provider) Register(route *Route) error {
	if err := route.Validate(); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	key := fmt.Sprintf("%s/%s %s", p.Cfg.Prefix, route.Method, route.Path)
	data, err := json.Marshal(route)
	if err != nil {
		return err
	}
	log.Printf("ETCDv3 put, endpoints: %v, key: %s, data: %s\n", p.Etcd.Endpoints(), key, string(data))
	_, err = p.Etcd.Put(ctx, key, string(data))
	return err
}

type Route struct {
	Method      string      `json:"method" yaml:"method"`
	Path        string      `json:"path" yaml:"path"`
	ServiceURL  string      `json:"service_url" yaml:"service_url"`
	BackendPath string      `json:"backend_path" yaml:"backend_path"`
	Auth        *pb.APIAuth `json:"auth" yaml:"auth"`
}

func (r *Route) Validate() error {
	r.Method = strings.TrimSpace(r.Method)
	r.Method = strings.ToUpper(r.Method)
	if r.Method == "" {
		r.Method = http.MethodGet
	}
	r.Path = strings.TrimSpace(r.Path)
	r.Path = "/" + strings.TrimLeft(r.Path, "/")
	r.BackendPath = strings.TrimSpace(r.BackendPath)
	r.BackendPath = "/" + strings.TrimLeft(r.BackendPath, "/")

	u, err := url.Parse(r.ServiceURL)
	if err != nil {
		return errors.Wrap(err, "failed to parse route.ServiceURL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.Errorf("scheme of route.ServiceURL must be http or https, got %s", u.Scheme)
	}
	if len(u.Host) == 0 {
		return errors.New("host of route.ServiceURL must not be empty")
	}

	if r.Auth == nil {
		r.Auth = new(pb.APIAuth)
	}

	return nil
}

func (r *Route) Get(key string) any {
	// for "path", "method"
	field := reflect.ValueOf(r).Elem().FieldByName(key)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}

	// for "CheckLogin", "CheckToken" ...
	if r.Auth == nil {
		r.Auth = new(pb.APIAuth)
	}
	field = reflect.ValueOf(r.Auth).Elem().FieldByName(key)
	if field.IsValid() || field.CanInterface() {
		return field.Interface()
	}

	return nil
}

type Config struct {
	Prefix string `file:"prefix" default:"/openapi/apis"`
}
