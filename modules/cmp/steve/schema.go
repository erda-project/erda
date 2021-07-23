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

package steve

import (
	"context"
	"strings"

	"github.com/rancher/apiserver/pkg/store/apiroot"
	"github.com/rancher/apiserver/pkg/subscribe"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/steve/pkg/attributes"
	"github.com/rancher/steve/pkg/client"
	"github.com/rancher/steve/pkg/resources/apigroups"
	"github.com/rancher/steve/pkg/resources/cluster"
	"github.com/rancher/steve/pkg/resources/formatters"
	"github.com/rancher/steve/pkg/resources/userpreferences"
	"github.com/rancher/steve/pkg/schema"
	steveschema "github.com/rancher/steve/pkg/schema"
	"github.com/rancher/steve/pkg/stores/proxy"
	"github.com/rancher/wrangler/pkg/slice"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"

	cmpproxy "github.com/erda-project/erda/modules/cmp/steve/proxy"
)

func DefaultSchemas(ctx context.Context, baseSchema *types.APISchemas, cg proxy.ClientGetter,
	schemaFactory steveschema.Factory) error {
	subscribe.Register(baseSchema)
	apiroot.Register(baseSchema, []string{"v1"}, "proxy:/apis")
	cluster.Register(ctx, baseSchema, cg, schemaFactory)
	userpreferences.Register(baseSchema)
	return nil
}

func DefaultSchemaTemplates(cf *client.Factory,
	baseSchemas *types.APISchemas,
	discovery discovery.DiscoveryInterface) []schema.Template {
	return []schema.Template{
		DefaultTemplate(cf),
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
			ID: "management.cattle.io.cluster",
			Customize: func(apiSchema *types.APISchema) {
				cluster.AddApply(baseSchemas, apiSchema)
			},
		},
	}
}

func DefaultTemplate(clientGetter proxy.ClientGetter) schema.Template {
	return schema.Template{
		Store:     cmpproxy.NewProxyStore(clientGetter),
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
