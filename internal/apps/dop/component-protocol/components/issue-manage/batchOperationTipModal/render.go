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

package batchOperationTipModal

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/types"
)

func init() {
	base.InitProviderWithCreator("issue-manage", "batchOperationTipModal", func() servicehub.Provider {
		return &BatchOperationTipModal{}
	})
}
func (bot *BatchOperationTipModal) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	bot.SDK = cputil.SDK(ctx)
	projectid, _ := strconv.ParseUint(bot.SDK.InParams["projectId"].(string), 10, 64)
	bot.ctx = ctx
	var IssueDoshboardBatchSubmit cptype.OperationKey = "batchSubmit"
	err := cputil.ObjJSONTransfer(c.Operations, &bot.Operations)
	if err != nil {
		return err
	}
	switch event.Operation {
	case cptype.InitializeOperation:
		bot.getOperations()
		err := cputil.ObjJSONTransfer(bot.Operations, &c.Operations)
		if err != nil {
			return err
		}
		bot.getProps()
		err = cputil.ObjJSONTransfer(bot.Props, &c.Props)
		if err != nil {
			return err
		}
		return nil
	case cptype.RenderingOperation:
		bot.getOperations()
		err := cputil.ObjJSONTransfer(bot.Operations, &c.Operations)
		if err != nil {
			return err
		}
		bot.State.Visible = false
		operationKey := (*gs)["OperationKey"]
		if operationKey == nil {
			break
		}
		if operationKey == "delete" {
			bot.State.Visible = true
			selectRowKeys := (*gs)["SelectedRowKeys"].([]string)
			bot.State.SelectedRowKeys = selectRowKeys
			ops := Operation{}
			err := cputil.ObjJSONTransfer(bot.Operations["onOk"], &ops)
			if err != nil {
				return err
			}
			ops.Meta.Type = "delete"
			bot.Operations["onOk"] = ops
			bot.Props.Content = bot.getContent(bot.SDK.I18n("Delete following items"), selectRowKeys)
		}
	case IssueDoshboardBatchSubmit:
		bot.State.Visible = false
		ops := Operation{}
		err := cputil.ObjJSONTransfer(bot.Operations["onOk"], &ops)
		if err != nil {
			return err
		}
		if ops.Meta.Type == "delete" {
			bot.ctx = ctx
			cputil.MustObjJSONTransfer(&c.State, &bot.State)
			bot.State.Visible = false
			issueSvc := ctx.Value(types.IssueService).(pb.IssueCoreServiceServer)
			bot.issueSvc = issueSvc
			_, err = bot.DeleteItems(bot.State.SelectedRowKeys, projectid)
			if err != nil {
				return err
			}
			(*gs)["OperationKey"] = ""
			c.State["selectedRowKeys"] = []string{}
		}
	}
	return bot.SetComponent(c)
}

func (bot *BatchOperationTipModal) SetComponent(c *cptype.Component) error {
	err := cputil.ObjJSONTransfer(bot.Props, &c.Props)
	if err != nil {
		return err
	}
	err = cputil.ObjJSONTransfer(bot.State, &c.State)
	if err != nil {
		return err
	}
	err = cputil.ObjJSONTransfer(bot.Operations, &c.Operations)
	if err != nil {
		return err
	}
	return nil
}

func (bot *BatchOperationTipModal) getContent(tip string, nodeIDs []string) string {
	content := fmt.Sprintf("%s\n", tip)
	for _, id := range nodeIDs {
		splits := strings.Split(id, "/")
		content += fmt.Sprintf("%s\n", splits[0])
	}
	return content
}

func (bot *BatchOperationTipModal) DeleteItems(itemIDs []string, projectID uint64) (*pb.BatchDeleteIssueResponse, error) {
	var req = &pb.BatchDeleteIssueRequest{}
	req.Ids = itemIDs
	req.ProjectID = projectID
	k, err := bot.issueSvc.BatchDeleteIssues(bot.ctx, req)
	if err != nil {
		return nil, err
	}
	return k, err
}

func (bot *BatchOperationTipModal) getOperations() {
	bot.Operations = map[string]interface{}{
		"onOk": Operation{
			Key:        "batchSubmit",
			Reload:     true,
			SuccessMsg: bot.SDK.I18n("Items deleted successfully"),
			Meta:       Meta{Type: ""},
		},
	}
}

func (bot *BatchOperationTipModal) getProps() {
	bot.Props = Props{
		Status:  "warning",
		Content: "",
		Title:   bot.SDK.I18n("warning"),
	}
}
