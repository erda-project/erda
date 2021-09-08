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

package issueFilter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
	"github.com/erda-project/erda/pkg/strutil"
)

func (i *ComponentFilter) GenComponentState(c *cptype.Component) error {
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

type SaveMeta struct {
	Name string `json:"name"`
}

type DeleteMeta struct {
	ID string `json:"id"`
}

func getMeta(ori map[string]interface{}, dst interface{}) error {
	m := ori["meta"]
	if m == nil {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func (f *ComponentFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	// init filter
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	if err := f.GenComponentState(c); err != nil {
		return err
	}

	if err := f.initFilterBms(); err != nil {
		return err
	}

	// operation
	switch event.Operation.String() {
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		if err := f.InitDefaultOperation(ctx, f.State); err != nil {
			return err
		}
	case f.Operations[OperationKeyFilter].Key.String():
		if f.State.FrontendChangedKey == string(PropConditionKeyFilterID) {
			f.FlushOptsFromBm = f.State.FrontendConditionValues.FilterID
		}
		// use rendering later `PostSetState`
	case f.Operations[OperationKeyCreatorSelectMe].Key.String():
		f.State.FrontendConditionValues.CreatorIDs = []string{f.sdk.Identity.UserID}
	case f.Operations[OperationKeyAssigneeSelectMe].Key.String():
		f.State.FrontendConditionValues.AssigneeIDs = []string{f.sdk.Identity.UserID}
	case f.Operations[OperationKeyOwnerSelectMe].Key.String():
		f.State.FrontendConditionValues.OwnerIDs = []string{f.sdk.Identity.UserID}
	case f.Operations[OperationKeySaveFilter].Key.String():
		if len(f.Bms) >= conf.MaxIssueFilterBm() {
			// wont go to here
			return fmt.Errorf("issue filter bookmarks execced limit: %d", conf.MaxIssueFilterBm())
		}
		var meta SaveMeta
		if err := getMeta(event.OperationData, &meta); err != nil {
			return err
		}
		pageKey := f.issueFilterBmSvc.GenPageKey(f.InParams.FrontendFixedIteration, f.InParams.FrontendFixedIssueType)
		filterID, err := f.issueFilterBmSvc.Create(&dao.IssueFilterBookmark{
			Name:         meta.Name,
			UserID:       f.sdk.Identity.UserID,
			ProjectID:    strutil.String(f.InParams.ProjectID),
			PageKey:      pageKey,
			FilterEntity: f.InParams.FrontendUrlQuery,
		})
		if err != nil {
			return err
		}
		f.State.FrontendConditionValues.FilterID = filterID
		// re-init bookmarks
		if err := f.initFilterBms(); err != nil {
			return err
		}
	case f.Operations[OperationKeyDeleteFilter].Key.String():
		var meta DeleteMeta
		if err := getMeta(event.OperationData, &meta); err != nil {
			return err
		}
		if err := f.issueFilterBmSvc.Delete(meta.ID); err != nil {
			return err
		}
		// re-init bookmarks
		if err := f.initFilterBms(); err != nil {
			return err
		}
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

	if f.FlushOptsFromBm != "" {
		err = f.flushOptsByFilterID(f.FlushOptsFromBm)
		if err != nil {
			return err
		}
	}

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
	filterID := f.State.FrontendConditionValues.FilterID
	f.State.FrontendConditionValues.FilterID = "" // remove filterID, it's not filter entity
	defer func() {
		f.State.FrontendConditionValues.FilterID = filterID // restore
	}()

	fb, err := json.Marshal(f.State.FrontendConditionValues)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(fb)
	return base64Str, nil
}

func (f *ComponentFilter) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}

func (f *ComponentFilter) InitDefaultOperation(ctx context.Context, state State) error {
	f.Props = filter.Props{Delay: 2000}
	f.Operations = GetAllOperations()
	f.State.FrontendConditionProps = f.generateFrontendConditionProps(ctx, f.InParams.FrontendFixedIssueType, state)

	stateBelongs := map[string][]apistructs.IssueStateBelong{
		"TASK":        {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
		"REQUIREMENT": {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
		"BUG":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved},
		"ALL":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResloved},
	}[f.InParams.FrontendFixedIssueType]
	types := []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}
	res := make(map[string][]int64)
	res["ALL"] = make([]int64, 0)
	for _, v := range types {
		req := &apistructs.IssueStatesGetRequest{
			ProjectID:    f.InParams.ProjectID,
			StateBelongs: stateBelongs,
			IssueType:    v,
		}
		ids, err := f.issueStateSvc.GetIssueStateIDs(req)
		if err != nil {
			return err
		}
		res[v.String()] = ids
		res["ALL"] = append(res["ALL"], ids...)
	}

	// 初始化时从 url query params 中获取已经存在的过滤参数
	if f.InParams.FrontendUrlQuery != "" {
		filterID := f.determineFilterID(f.InParams.FrontendUrlQuery)
		if err := f.flushOptsByFilter(filterID, f.InParams.FrontendUrlQuery); err != nil {
			return err
		}
	} else {
		f.State.FrontendConditionValues.States = res[f.InParams.FrontendFixedIssueType]
	}

	return nil
}

func (f *ComponentFilter) determineFilterID(filterEntity string) string {
	for _, bm := range f.Bms {
		if bm.FilterEntity == filterEntity {
			return bm.ID
		}
	}
	return ""
}

func (f *ComponentFilter) flushOptsByFilterID(filterID string) error {
	for _, bm := range f.Bms {
		if bm.ID == filterID {
			return f.flushOptsByFilter(bm.ID, bm.FilterEntity)
		}
	}
	return nil
}

func (f *ComponentFilter) flushOptsByFilter(filterID, filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &f.State.FrontendConditionValues); err != nil {
		return err
	}
	f.State.FrontendConditionValues.FilterID = filterID
	return nil
}

func (f *ComponentFilter) initFilterBms() error {
	pageKey := f.issueFilterBmSvc.GenPageKey(f.InParams.FrontendFixedIteration, f.InParams.FrontendFixedIssueType)
	mp, err := f.issueFilterBmSvc.ListMyBms(f.sdk.Identity.UserID, strutil.String(f.InParams.ProjectID))
	if err != nil {
		return err
	}
	f.Bms = mp.GetByPageKey(pageKey)
	return nil
}
