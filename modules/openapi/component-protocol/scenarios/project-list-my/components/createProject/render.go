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

package createProject

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentProjectCreation) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentProjectCreation) GenComponentState(c *apistructs.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	return nil
}

func (i *ComponentProjectCreation) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return err
	}
	if err := i.GenComponentState(c); err != nil {
		return err
	}

	i.Props = Props{
		Visible: false,
		Text:    "创建项目",
		Type:    "primary",
	}

	i.Operations = map[string]interface{}{
		"click": Operation{
			Key:    "createProject",
			Reload: false,
			Command: Command{
				Key:     "goto",
				Target:  "createProject",
				JumpOut: true,
			},
		},
	}

	orgID, err := strconv.Atoi(i.ctxBdl.Identity.OrgID)
	if err != nil {
		return err
	}

	if i.State.IsEmpty {
		permission, err := i.ctxBdl.Bdl.CheckPermission(&apistructs.PermissionCheckRequest{
			UserID:   i.ctxBdl.Identity.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.ProjectResource,
			Action:   apistructs.CreateAction,
		})
		if err != nil {
			return err
		}
		if permission.Access {
			i.Props.Visible = true
		}
	}

	c.Operations = i.Operations
	c.Props = i.Props
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentProjectCreation{}
}
