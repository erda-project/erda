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

package Header

import (
	"context"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (header *Header) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	var (
		resp data.Object
		err  error
		req  apistructs.SteveRequest
	)
	header.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)

	req.Namespace = header.SDK.InParams["namespace"].(string)
	req.ClusterName = header.SDK.InParams["clusterName"].(string)
	req.OrgID = header.SDK.Identity.OrgID
	req.UserID = header.SDK.Identity.UserID
	req.Type = apistructs.K8SNode
	req.Name = header.SDK.InParams["nodeId"].(string)

	resp, err = header.CtxBdl.GetSteveResource(&req)
	if err != nil {
		return err
	}
	(*gs)["node"] = resp
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "header", func() servicehub.Provider {
		return &Header{Type: "RowContainer"}
	})
}
