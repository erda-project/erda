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

	cmpproxy "github.com/erda-project/erda/modules/cmp/steve/proxy"
)

func DefaultSchemas(baseSchema *types.APISchemas) {
	subscribe.Register(baseSchema)
	apiroot.Register(baseSchema, []string{"v1"}, "proxy:/apis")
}

func DefaultSchemaTemplates(cf *client.Factory,
	discovery discovery.DiscoveryInterface, asl accesscontrol.AccessSetLookup) []schema.Template {
	return []schema.Template{
		DefaultTemplate(cf, asl),
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
	}
}

func DefaultTemplate(clientGetter proxy.ClientGetter, asl accesscontrol.AccessSetLookup) schema.Template {
	return schema.Template{
		Store:     cmpproxy.NewProxyStore(clientGetter, asl),
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
