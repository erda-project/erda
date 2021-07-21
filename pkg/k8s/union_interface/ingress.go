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

package union_interface

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

type MatchType int

const (
	PrefixMatch MatchType = 0
	ExactMatch  MatchType = 1
)

type IngressRoute struct {
	Domain    string
	Path      string
	MatchType MatchType
}

type IngressBackend struct {
	ServiceName string
	ServicePort int
}

type IngressMaterial struct {
	Name      string
	Namespace string
	Routes    []IngressRoute
	Backend   IngressBackend
	NeedTLS   bool
}

type IngressesHelper interface {
	Ingresses(namespace string) IngressInterface
	NewIngress(IngressMaterial) interface{}
	IngressAnnotationBatchSet(ingress interface{}, kvs map[string]string) error
	IngressAnnotationSet(ingress interface{}, key, value string) error
	IngressAnnotationBatchGet(ingress interface{}) (map[string]string, error)
	IngressAnnotationGet(ingress interface{}, key string) (string, error)
	IngressAnnotationClear(ingress interface{}, key string) error
}

type IngressInterface interface {
	Create(ctx context.Context, ingress interface{}, opts v1.CreateOptions) (interface{}, error)
	Update(ctx context.Context, ingress interface{}, opts v1.UpdateOptions) (interface{}, error)
	UpdateStatus(ctx context.Context, ingress interface{}, opts v1.UpdateOptions) (interface{}, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (interface{}, error)
	List(ctx context.Context, opts v1.ListOptions) (interface{}, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result interface{}, err error)
	Apply(ctx context.Context, ingress interface{}, opts v1.ApplyOptions) (result interface{}, err error)
	ApplyStatus(ctx context.Context, ingress interface{}, opts v1.ApplyOptions) (result interface{}, err error)
}
