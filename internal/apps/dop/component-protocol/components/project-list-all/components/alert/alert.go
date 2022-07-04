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

package alert

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/types"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/pkg/common/apis"
)

type ComponentAction struct {
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	orgSvc := ctx.Value(types.OrgService).(org.ClientInterface)

	orgResp, err := orgSvc.GetOrg(apis.WithInternalClientContext(ctx, "dop"), &orgpb.GetOrgRequest{IdOrName: sdk.Identity.OrgID})
	if err != nil {
		return err
	}

	org := orgResp.Data
	if org.BlockoutConfig.BlockDev || org.BlockoutConfig.BlockProd || org.BlockoutConfig.BlockStage || org.BlockoutConfig.BlockTest {
		return json.Unmarshal([]byte(`{ "visible": true, "message": "`+cputil.I18n(ctx, "blockMessage")+`", "type": "warning" }`), &c.Props)
	}
	return json.Unmarshal([]byte(`{ "visible": false }`), &c.Props)
}

func init() {
	base.InitProviderWithCreator("project-list-all", "alert", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
