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

package issueFilter

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func (i *ComponentFilter) GenComponentState(c *apistructs.Component) error {
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

func (f *ComponentFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	// init filter
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	if err := f.GenComponentState(c); err != nil {
		return err
	}

	// operation
	switch event.Operation.String() {
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		if err := f.InitDefaultOperation(f.State); err != nil {
			return err
		}
	case f.Operations[OperationKeyFilter].Key.String():
		// use rendering
	case f.Operations[OperationKeyCreatorSelectMe].Key.String():
		f.State.FrontendConditionValues.CreatorIDs = []string{f.CtxBdl.Identity.UserID}
	case f.Operations[OperationKeyAssigneeSelectMe].Key.String():
		f.State.FrontendConditionValues.AssigneeIDs = []string{f.CtxBdl.Identity.UserID}
	case f.Operations[OperationKeyOwnerSelectMe].Key.String():
		f.State.FrontendConditionValues.OwnerIDs = []string{f.CtxBdl.Identity.UserID}
	}

	if err := f.PostSetState(); err != nil {
		return err
	}

	if err := f.SetToProtocolComponent(c); err != nil {
		return err
	}

	return nil
}

func (f *ComponentFilter) PostSetState() error {
	var err error

	// url query
	f.State.Base64UrlQueryParams, err = f.generateUrlQueryParams()
	if err != nil {
		return err
	}

	// condition props
	f.State.FrontendConditionProps, err = f.SetStateConditionProps()
	if err != nil {
		return err
	}

	// condition values

	// issuePagingRequest
	f.State.IssuePagingRequest, err = f.generateIssuePagingRequest()
	if err != nil {
		return err
	}

	return nil
}

func (f *ComponentFilter) generateUrlQueryParams() (string, error) {
	fb, err := json.Marshal(f.State.FrontendConditionValues)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(fb)
	return base64Str, nil
}

func (f *ComponentFilter) SetToProtocolComponent(c *apistructs.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}

func (f *ComponentFilter) InitDefaultOperation(state State) error {
	f.Props = filter.Props{Delay: 300}
	f.Operations = GetAllOperations()
	f.State.FrontendConditionProps = generateFrontendConditionProps(f.InParams.FrontendFixedIssueType, state)

	// 初始化时从 url query params 中获取已经存在的过滤参数
	if f.InParams.FrontendUrlQuery != "" {
		b, err := base64.StdEncoding.DecodeString(f.InParams.FrontendUrlQuery)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &f.State.FrontendConditionValues); err != nil {
			return err
		}
	} else {
		f.State.FrontendConditionValues.StateBelongs = map[string][]apistructs.IssueStateBelong{
			"TASK":        {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
			"REQUIREMENT": {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
			"BUG":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved},
			"ALL":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved},
		}[f.InParams.FrontendFixedIssueType]
	}

	return nil
}
