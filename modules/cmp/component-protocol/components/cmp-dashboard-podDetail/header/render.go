package Header

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda/apistructs"
)

func (header *Header) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	var (
		resp data.Object
		err  error
		req  apistructs.SteveRequest
	)
	header.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	header.SDK = cputil.SDK(ctx)
	req.Namespace = header.SDK.InParams["namespace"].(string)
	req.ClusterName = header.SDK.InParams["clusterName"].(string)
	req.OrgID = header.SDK.Identity.OrgID
	req.UserID = header.SDK.Identity.UserID
	req.Type = apistructs.K8SPod
	req.Name = header.SDK.InParams["podName"].(string)

	resp, err = bdl.Bdl.GetSteveResource(&req)
	if err != nil {
		return err
	}
	(*gs)["pod"] = resp
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "header", func() servicehub.Provider {
		return &Header{Type: "RowContainer"}
	})
}
