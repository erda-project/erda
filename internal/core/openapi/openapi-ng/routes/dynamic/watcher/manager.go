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

package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/erda-project/erda/internal/core/openapi/openapi-ng/routes"
	httpapi "github.com/erda-project/erda/pkg/common/httpapi"
	"github.com/erda-project/erda/pkg/discover"
)

type serviceInfo struct {
	Service string `json:"service"`
	URL     string `json:"url"`
}

func (p *provider) listServices() interface{} {
	var list []*serviceInfo
	services := discover.Services()
	for _, service := range services {
		url, err := p.Discover.ServiceURL("http", service)
		if err == nil {
			list = append(list, &serviceInfo{
				Service: service,
				URL:     url,
			})
		}
	}
	return httpapi.Success(list)
}

func (p *provider) listAPIProxies() interface{} {
	list, err := p.getAPIProxies()
	if err != nil {
		return httpapi.Errors.Internal(err)
	}
	return httpapi.Success(list)
}

func (p *provider) setAPIProxy(body routes.APIProxy) interface{} {
	body.Method = strings.TrimSpace(body.Method)
	body.Path = formatPath(strings.TrimSpace(body.Path))
	body.BackendPath = formatPath(strings.TrimSpace(body.BackendPath))
	err := body.Validate()
	if err != nil {
		return httpapi.Errors.InvalidParameter(err)
	}
	err = p.saveAPIProxy(&body, clientv3.NoLease)
	if err != nil {
		return httpapi.Errors.Internal(err)
	}
	return httpapi.Success("OK")
}

func (p *provider) setAPIProxyWithKeepAlive(req *http.Request, body struct {
	List []*routes.APIProxy `json:"list"`
}) interface{} {
	if len(body.List) <= 0 {
		return httpapi.Errors.InvalidParameter("proxy list is empty")
	}
	for _, proxy := range body.List {
		proxy.Method = strings.TrimSpace(proxy.Method)
		proxy.Path = formatPath(strings.TrimSpace(proxy.Path))
		proxy.BackendPath = formatPath(strings.TrimSpace(proxy.BackendPath))
		if err := proxy.Validate(); err != nil {
			return httpapi.Errors.InvalidParameter(err)
		}
	}
	ctx := req.Context()
	leaseID, err := p.grantLeaseID(ctx)
	if err != nil {
		return httpapi.Errors.Internal(err)
	}
	for _, proxy := range body.List {
		if err := p.saveAPIProxy(proxy, leaseID); err != nil {
			return httpapi.Errors.Internal(err)
		}
	}
	err = p.keepAliveLease(ctx, leaseID)
	if err != nil {
		return httpapi.Errors.Internal(err)
	}
	return httpapi.Success("OK")
}

func (p *provider) saveAPIProxy(a *routes.APIProxy, leaseID clientv3.LeaseID) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.EtcdRequestTimeout)
	defer cancel()
	key := fmt.Sprintf("%s/%s %s", p.Cfg.Prefix, a.Method, a.Path)
	byts, err := json.Marshal(a)
	if err != nil {
		return err
	}
	var opts []clientv3.OpOption
	if leaseID != clientv3.NoLease {
		opts = append(opts, clientv3.WithLease(leaseID))
	}
	_, err = p.Etcd.Put(ctx, key, string(byts), opts...)
	return err
}

func (p *provider) removeAPIProxy(body routes.APIProxy) interface{} {
	body.Method = strings.TrimSpace(body.Method)
	body.Path = formatPath(strings.TrimSpace(body.Path))
	err := p.deleteAPIProxy(body.Method, body.Path)
	if err != nil {
		return httpapi.Errors.Internal(err)
	}
	return httpapi.Success("OK")
}

func (p *provider) deleteAPIProxy(method, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.EtcdRequestTimeout)
	defer cancel()
	key := fmt.Sprintf("%s/%s %s", p.Cfg.Prefix, method, path)
	_, err := p.Etcd.Delete(ctx, key)
	return err
}

func (p *provider) getAPIProxies() (list []*routes.APIProxy, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.EtcdRequestTimeout)
	defer cancel()
	resp, err := p.Etcd.Get(ctx, p.Cfg.Prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		idx := strings.Index(key, " ")
		if idx < 0 {
			p.Log.Errorf("invalid api key %q format in etcd", key)
			continue
		}
		api := &routes.APIProxy{
			Method: strings.TrimSpace(key[:idx]),
			Path:   strings.TrimSpace(key[idx+1:]),
		}
		err := json.Unmarshal(kv.Value, api)
		if err != nil {
			p.Log.Errorf("invalid api (%s) format in etcd: %v", key, err)
			continue
		}
		if err := api.Validate(); err != nil {
			p.Log.Errorf("invalid api (%s): %v", key, err)
			continue
		}
		list = append(list, api)
	}
	return list, nil
}

func (p *provider) grantLeaseID(ctx context.Context) (clientv3.LeaseID, error) {
	resp, err := p.Etcd.Grant(ctx, int64(p.Cfg.KeepAliveTTL.Seconds()))
	if err != nil {
		return clientv3.NoLease, err
	}
	return resp.ID, nil
}

func (p *provider) revokeLeaseID(leaseID clientv3.LeaseID) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Cfg.EtcdRequestTimeout)
	defer cancel()
	_, err := p.Etcd.Revoke(ctx, leaseID)
	if err != nil {
		p.Log.Error(err)
	}
}

func (p *provider) keepAliveLease(ctx context.Context, leaseID clientv3.LeaseID) error {
	defer p.revokeLeaseID(leaseID)
	keepAlive, err := p.Etcd.KeepAlive(ctx, leaseID)
	if err != nil {
		return err
	}
	for {
		select {
		case _, ok := <-keepAlive:
			// eat messages until keep alive channel closes
			if !ok {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func formatPath(path string) string {
	return "/" + strings.TrimLeft(path, "/")
}
