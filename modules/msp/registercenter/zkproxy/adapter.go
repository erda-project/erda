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

package zkproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/erda-project/erda-proto-go/msp/registercenter/pb"
	"github.com/erda-project/erda/pkg/http/httputil"
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

func (a *Adapter) GetAllInterfaceList(projectID, env, namespace string) ([]*pb.Interface, error) {
	req, err := a.newRequestN(http.MethodGet, httputil.JoinPathR("listinterface", projectID, env), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.Interface, 0)
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		err = json.NewDecoder(resp.Body).Decode(&list)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (a *Adapter) GetRouteRule(interfaceName, projectID, env, namespace string) (*pb.RequestRule, error) {
	resp, err := a.doRequetN(http.MethodGet, httputil.JoinPathR("listinterface", "route", interfaceName, projectID, env), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	return getRequestRuleFromResponse(resp)
}

func (a *Adapter) CreateRouteRule(interfaceName, projectID, env, namespace string, rule *pb.RequestRule) (*pb.RequestRule, error) {
	byts, _ := json.Marshal(rule)
	resp, err := a.doRequetN(http.MethodPost, httputil.JoinPathR("listinterface", "route", interfaceName, projectID, env), nil, bytes.NewReader(byts), namespace)
	if err != nil {
		return nil, err
	}
	return getRequestRuleFromResponse(resp)
}

func (a *Adapter) DeleteRouteRule(interfaceName, projectID, env, namespace string) (*pb.RequestRule, error) {
	resp, err := a.doRequetN(http.MethodDelete, httputil.JoinPathR("listinterface", "route", interfaceName, projectID, env), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	return getRequestRuleFromResponse(resp)
}

func getRequestRuleFromResponse(resp *http.Response) (*pb.RequestRule, error) {
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body = struct {
			Success bool            `json:"success"`
			Data    *pb.RequestRule `json:"data"`
			Err     interface{}     `json:"err"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("failed request: %v", body.Err)
		}
		return body.Data, nil
	}
	// return nil, fmt.Errorf("unexpect status=%d", resp.StatusCode)
	return nil, nil
}

func (a *Adapter) GetHostRule(projectID, env, appID, namespace string) (*pb.HostRules, error) {
	resp, err := a.doRequetN(http.MethodGet, httputil.JoinPathR("rule", "branch", projectID, env, appID), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	return getHostRulesFromResponse(resp)
}

func (a *Adapter) CreateHostRoute(projectID, env, appID, namespace string, rules *pb.HostRules) (*pb.HostRules, error) {
	byts, _ := json.Marshal(rules)
	resp, err := a.doRequetN(http.MethodPost, httputil.JoinPathR("rule", "branch", projectID, env, appID), nil, bytes.NewReader(byts), namespace)
	if err != nil {
		return nil, err
	}
	return getHostRulesFromResponse(resp)
}

func (a *Adapter) DeleteHostRoute(projectID, env, appID, namespace string) (*pb.HostRules, error) {
	resp, err := a.doRequetN(http.MethodDelete, httputil.JoinPathR("rule", "branch", projectID, env, appID), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	return getHostRulesFromResponse(resp)
}

func getHostRulesFromResponse(resp *http.Response) (*pb.HostRules, error) {
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body = struct {
			Success bool          `json:"success"`
			Data    *pb.HostRules `json:"data"`
			Err     interface{}   `json:"err"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("failed request: %v", body.Err)
		}
		return body.Data, nil
	}
	// return nil, fmt.Errorf("unexpect status=%d", resp.StatusCode)
	return nil, nil
}

func (a *Adapter) GetHostRuntimeRule(host, projectID, env, namespace string) (*pb.HostRuntimeRules, error) {
	resp, err := a.doRequetN(http.MethodGet, httputil.JoinPathR("rule", "host", projectID, env, host), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	return getHostRuntimeRulesFromResponse(resp)
}

func (a *Adapter) CreateHostRuntimeRule(projectID, env, host, namespace string, rules *pb.HostRuntimeRules) (*pb.HostRuntimeRules, error) {
	byts, _ := json.Marshal(rules)
	resp, err := a.doRequetN(http.MethodPost, httputil.JoinPathR("rule", "host", projectID, env, host), nil, bytes.NewReader(byts), namespace)
	if err != nil {
		return nil, err
	}
	return getHostRuntimeRulesFromResponse(resp)
}

func getHostRuntimeRulesFromResponse(resp *http.Response) (*pb.HostRuntimeRules, error) {
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body = struct {
			Success bool                 `json:"success"`
			Data    *pb.HostRuntimeRules `json:"data"`
			Err     interface{}          `json:"err"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("failed request: %v", body.Err)
		}
		return body.Data, nil
	}
	// return nil, fmt.Errorf("unexpect status=%d", resp.StatusCode)
	return nil, nil
}

func (a *Adapter) GetAllHostRuntimeRules(projectID, env, appID, namespace string) (*pb.HostRuntimeInterfaces, error) {
	resp, err := a.doRequetN(http.MethodGet, httputil.JoinPathR("listhostinterface", "timestamp", projectID, env, appID), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body = struct {
			Success bool                      `json:"success"`
			Data    *pb.HostRuntimeInterfaces `json:"data"`
			Err     interface{}               `json:"err"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("failed request: %v", body.Err)
		}
		return body.Data, nil
	}
	// return nil, fmt.Errorf("unexpect status=%d", resp.StatusCode)
	return nil, nil
}

func (a *Adapter) GetDubboInterfaceTime(interfaceName, projectID, env, namespace string) (*pb.DubboInterfaceTime, error) {
	resp, err := a.doRequetN(http.MethodGet, httputil.JoinPathR("interface", "timestamp", interfaceName, projectID, env), nil, nil, namespace)
	if err != nil {
		return nil, err
	}
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusMultipleChoices {
		var body = struct {
			Success bool                   `json:"success"`
			Data    *pb.DubboInterfaceTime `json:"data"`
			Err     interface{}            `json:"err"`
		}{}
		err := json.NewDecoder(resp.Body).Decode(&body)
		if err != nil {
			return nil, err
		}
		if !body.Success {
			return nil, fmt.Errorf("failed request: %v", body.Err)
		}
		return body.Data, nil
	}
	// return nil, fmt.Errorf("unexpect status=%d", resp.StatusCode)
	return nil, nil
}

func (a *Adapter) doRequetN(method, path string, params url.Values, body io.Reader, namespace string) (*http.Response, error) {
	req, err := a.newRequestN(method, path, params, body, namespace)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *Adapter) newRequestN(method, path string, params url.Values, body io.Reader, namespace string) (*http.Request, error) {
	req, err := a.newRequest(method, path, params, body)
	if err == nil {
		req.Header.Set("namespace", namespace)
	}
	return req, err
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
