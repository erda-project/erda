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

package apistructs

import (
	"os"

	"github.com/pkg/errors"
)

type ComponentIngressUpdateResponse struct {
	Header
}

type ComponentIngressUpdateRequest struct {
	K8SNamespace  string `json:"k8sNamespace"`
	ComponentName string `json:"componentName"`
	ComponentPort int    `json:"componentPort"`
	// 若为空，则使用当前集群名称
	ClusterName string `json:"clusterName"`
	// 若为空，则使用ComponentName
	IngressName string `json:"ingressName"`
	// 若为空，则清除ingress
	Routes       []IngressRoute `json:"routes"`
	RouteOptions RouteOptions   `json:"routeOptions"`
}

type IngressRoute struct {
	Domain string
	Path   string
}

type RouteOptions struct {
	// 重写转发域名
	RewriteHost *string `json:"rewriteHost"`
	// 重写转发路径
	RewritePath *string `json:"rewritePath"`
	// Path中是否使用了正则
	UseRegex bool `json:"useRegex"`
	// 是否开启TLS，不填时，默认为true
	EnableTLS *bool `json:"enableTls"`
	// 参考: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/
	Annotations map[string]string `json:"annotations"`
	// 参考: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#configuration-snippet
	LocationSnippet *string `json:"locationSnippet"`
}

func (req *ComponentIngressUpdateRequest) CheckValid() error {
	if req.ComponentName == "" {
		return errors.New("component name is empty")
	}
	if req.ComponentPort == 0 {
		return errors.New("component port is empty")
	}
	for _, route := range req.Routes {
		if route.Domain == "" || route.Path == "" {
			return errors.New("invalid route")
		}
	}
	if req.RouteOptions.EnableTLS == nil {
		enabled := true
		req.RouteOptions.EnableTLS = &enabled
	}
	if req.IngressName == "" {
		req.IngressName = req.ComponentName
	}
	if req.ClusterName == "" {
		req.ClusterName = os.Getenv("DICE_CLUSTER_NAME")
	}
	return nil
}
