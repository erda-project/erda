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

package content

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

type Content struct {
	ctxBdl protocol.ContextBundle
	Type   string `json:"type"`
	State  State  `json:"state"`
	Props  Props  `json:"props"`
}

type Props struct {
	Visible bool `json:"visible"`
}

type State struct {
	ProsNum int `json:"prosNum"`
}

func (t *Content) getWorkbenchData() (*apistructs.WorkbenchResponse, error) {
	orgID, err := strconv.ParseUint(t.ctxBdl.Identity.OrgID, 10, 64)
	if err != nil {
		return nil, err
	}
	req := apistructs.WorkbenchRequest{
		OrgID:     orgID,
		PageSize:  1,
		PageNo:    1,
		IssueSize: 1,
	}
	res, err := t.ctxBdl.Bdl.GetWorkbenchData(t.ctxBdl.Identity.UserID, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (this *Content) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	this.ctxBdl = bdl
	return nil
}

func (this *Content) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := this.SetCtxBundle(ctx); err != nil {
		return err
	}
	this.Type = "Container"
	this.Props.Visible = true
	return nil
}

func RenderCreator() protocol.CompRender {
	return &Content{}
}
