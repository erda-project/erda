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

package steve

import (
	"context"
	"strings"

	"github.com/rancher/apiserver/pkg/store/apiroot"
	"github.com/rancher/apiserver/pkg/subscribe"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/accesscontrol"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/steve/pkg/client"
	"github.com/rancher/steve/pkg/resources/apigroups"
	"github.com/rancher/steve/pkg/resources/formatters"
	"github.com/rancher/steve/pkg/schema"
	"github.com/rancher/steve/pkg/stores/proxy"
	"github.com/rancher/wrangler/pkg/slice"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"

	"github.com/erda-project/erda/modules/cmp/cache"
	"github.com/erda-project/erda/modules/cmp/conf"
	fm "github.com/erda-project/erda/modules/cmp/steve/formatter"
	cmpproxy "github.com/erda-project/erda/modules/cmp/steve/proxy"
)

func DefaultSchemas(baseSchema *types.APISchemas) {
	subscribe.Register(baseSchema)
	apiroot.Register(baseSchema, []string{"v1"}, "proxy:/apis")
}

func DefaultSchemaTemplates(ctx context.Context, cf *client.Factory,
	discovery discovery.DiscoveryInterface, asl accesscontrol.AccessSetLookup, k8sInterface kubernetes.Interface) []schema.Template {
	cache, _ := cache.New(conf.CacheSize(), conf.CacheSegSize())
	nodeFormatter := fm.NewNodeFormatter(ctx, cache, k8sInterface)
	return []schema.Template{
		DefaultTemplate(cf, asl, cache),
		apigroups.Template(discovery),
		{
			ID:        "configmap",
			Formatter: formatters.DropHelmData,
		},
		{
			ID:        "secret",
			Formatter: formatters.DropHelmData,
		},
		{
			ID:        "pod",
			Formatter: formatters.Pod,
		},
		{
			ID:        "node",
			Formatter: nodeFormatter.Formatter,
		},
	}
}

func DefaultTemplate(clientGetter proxy.ClientGetter, asl accesscontrol.AccessSetLookup, cache *cache.Cache) schema.Template {
	return schema.Template{
		Store:     cmpproxy.NewProxyStore(clientGetter, asl, cache),
		Formatter: formatter(),
	}
}

func selfLink(gvr k8sschema.GroupVersionResource, meta metav1.Object) (prefix string) {
	buf := &strings.Builder{}
	if gvr.Group == "" {
		buf.WriteString("/api/v1/")
	} else {
		buf.WriteString("/apis/")
		buf.WriteString(gvr.Group)
		buf.WriteString("/")
		buf.WriteString(gvr.Version)
		buf.WriteString("/")
	}
	if meta.GetNamespace() != "" {
		buf.WriteString("namespaces/")
		buf.WriteString(meta.GetNamespace())
		buf.WriteString("/")
	}
	buf.WriteString(gvr.Resource)
	buf.WriteString("/")
	buf.WriteString(meta.GetName())
	return buf.String()
}

func formatter() types.Formatter {
	return func(request *types.APIRequest, resource *types.RawResource) {
		if resource.Schema == nil {
			return
		}

		gvr := attributes.GVR(resource.Schema)
		if gvr.Version == "" {
			return
		}

		meta, err := meta.Accessor(resource.APIObject.Object)
		if err != nil {
			return
		}
		selfLink := selfLink(gvr, meta)

		u := request.URLBuilder.RelativeToRoot(selfLink)
		resource.Links["view"] = u

		if _, ok := resource.Links["update"]; !ok && slice.ContainsString(resource.Schema.CollectionMethods, "PUT") {
			resource.Links["update"] = u
		}
	}
}
