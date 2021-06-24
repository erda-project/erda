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

package nacos

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
	"github.com/erda-project/erda/pkg/netportal"
)

// Adapter .
type Adapter struct {
	ClusterName string
	Addr        string
}

// NewAdapter .
func NewAdapter(clusterName, addr string) *Adapter {
	return &Adapter{
		ClusterName: clusterName,
		Addr:        addr,
	}
}

// GetDubboInterfaceList .
func (a *Adapter) GetDubboInterfaceList(namespace string) ([]*pb.Interface, error) {
	count, err := a.countInterface(namespace, "")
	if err != nil {
		return nil, err
	}
	result, err := a.getInterfaceDetail(namespace, "", 1, count)
	if err != nil {
		return nil, err
	}

	var keys []string
	values := make(map[string]*pb.Interface)
	for _, item := range result {
		if len(item.getSide()) <= 0 {
			continue
		}
		iname := item.getInterfaceName()
		if exist, ok := values[iname]; ok {
			mergeInterface(exist, item.ToInterface())
		} else {
			values[iname] = item.ToInterface()
		}
	}
	list := make([]*pb.Interface, 0, len(result))
	for _, key := range keys {
		list = append(list, values[key])
	}
	return list, nil
}

func mergeInterface(a *pb.Interface, b *pb.Interface) *pb.Interface {
	if a == nil {
		return b
	}
	if b != nil && a.Interfacename == b.Interfacename {
		providers := make(map[string]struct{})
		for _, p := range a.Providerlist {
			providers[p] = struct{}{}
		}
		for _, p := range b.Providerlist {
			if _, ok := providers[p]; !ok {
				a.Providerlist = append(a.Providerlist, p)
			}
		}
		for k, v := range b.Providermap {
			if _, ok := a.Providermap[k]; !ok {
				a.Providermap[k] = v
			}
		}
		consumers := make(map[string]struct{})
		for _, p := range a.Consumerlist {
			consumers[p] = struct{}{}
		}
		for _, p := range b.Consumerlist {
			if _, ok := providers[p]; !ok {
				a.Providerlist = append(a.Providerlist, p)
			}
		}
		for k, v := range b.Consumermap {
			if _, ok := a.Consumermap[k]; !ok {
				a.Consumermap[k] = v
			}
		}
	}
	return a
}

func (a *Adapter) GetHTTPInterfaceList(namespace string) ([]*pb.HTTPService, error) {
	count, err := a.countInterface(namespace, "")
	if err != nil {
		return nil, err
	}
	result, err := a.getInterfaceDetail(namespace, "", 1, count)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.HTTPService, 0, len(result))
	for _, item := range result {
		if len(item.getSide()) > 0 {
			continue
		}
		list = append(list, item.ToHTTPService())
	}
	return list, nil
}

func (a *Adapter) EnableHTTPService(namespace string, service *pb.EnableHTTPService) (*pb.EnableHTTPService, error) {
	count, err := a.countInterface(namespace, service.ServiceName)
	if err != nil {
		return nil, err
	}
	results, err := a.getInterfaceDetail(namespace, service.ServiceName, 1, count)
	if err != nil {
		return nil, err
	}
	var result *ServiceSearchResult
	for _, r := range results {
		if !strings.EqualFold(r.ServiceName, service.ServiceName) {
			continue
		}
		for _, ip := range r.getIPs() {
			if ip == service.Ip {
				result = r
				break
			}
		}
	}
	if result == nil {
		return nil, nil
	}
	h := result.getHostByIP(service.Ip)
	params := url.Values{}
	params.Set("serviceName", service.ServiceName)
	params.Set("clusterName", "DEFAULT")
	params.Set("ip", service.Ip)
	params.Set("port", service.Port)
	params.Set("ephemeral", "true")
	params.Set("weight", strconv.FormatInt(h.Weight, 10))
	params.Set("enabled", strconv.FormatBool(service.Online))
	params.Set("namespaceId", namespace)
	byts, _ := json.Marshal(h.MetaData)
	params.Set("metadata", string(byts))
	resp, err := a.doRequest(http.MethodPut, "/nacos/v1/ns/instance", params, nil)
	if err != nil {
		return nil, err
	}
	byts, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		if strings.EqualFold(string(byts), "ok") {
			return service, nil
		}
	}
	return nil, fmt.Errorf("unexpect status=%d, body=%q", resp.StatusCode, string(byts))
}

func (a *Adapter) countInterface(namespace, keyword string) (int64, error) {
	params := url.Values{}
	params.Set("hasIpCount", "false")
	params.Set("withInstances", "false")
	params.Set("pageNo", "1")
	params.Set("pageSize", "10")
	params.Set("serviceNameParam", keyword)
	params.Set("namespaceId", namespace)
	resp, err := a.doRequest(http.MethodGet, "/nacos/v1/ns/catalog/services", params, nil)
	if err != nil {
		return 0, err
	}
	var body struct {
		Count int64 `json:"count"`
	}
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		err = json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return 0, err
		}
	}
	return body.Count, nil
}

func (a *Adapter) getInterfaceDetail(namespace, keyword string, pageNo, pageSize int64) ([]*ServiceSearchResult, error) {
	params := url.Values{}
	params.Set("hasIpCount", "false")
	params.Set("withInstances", "true")
	params.Set("pageNo", strconv.FormatInt(pageNo, 10))
	params.Set("pageSize", strconv.FormatInt(pageSize, 10))
	params.Set("serviceNameParam", keyword)
	params.Set("namespaceId", namespace)
	resp, err := a.doRequest(http.MethodGet, "/nacos/v1/ns/catalog/services", params, nil)
	if err != nil {
		return nil, err
	}
	var list []*ServiceSearchResult
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		err = json.NewDecoder(resp.Body).Decode(&list)
		if err != nil {
			return nil, err
		}
	}
	filtered := make([]*ServiceSearchResult, 0, len(list))
	for _, item := range list {
		if len(item.ClusterMap) <= 0 || item.ClusterMap["DEFAULT"] == nil || len(item.ClusterMap["DEFAULT"].Hosts) <= 0 {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (a *Adapter) doRequest(method, path string, params url.Values, body io.Reader) (*http.Response, error) {
	req, err := a.newRequest(method, path, params, body)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *Adapter) newRequest(method, path string, params url.Values, body io.Reader) (*http.Request, error) {
	var ustr string
	if len(params) > 0 {
		ustr = fmt.Sprintf("%s%s?%s", a.Addr, path, params.Encode())
	} else {
		ustr = fmt.Sprintf("%s%s", a.Addr, path)
	}
	if len(a.ClusterName) > 0 {
		return netportal.NewNetportalRequest(a.ClusterName, method, ustr, body)
	}
	return http.NewRequest(method, ustr, body)
}
