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

package releaseFilter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/bundle/apierrors"
	cmpTypes "github.com/erda-project/erda/modules/cmp/component-protocol/types"
)

func init() {
	base.InitProviderWithCreator("release-manage", "releaseFilter", func() servicehub.Provider {
		return &ComponentReleaseFilter{}
	})
}

func (f *ComponentReleaseFilter) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	f.InitComponent(ctx)
	if err := f.GenComponentState(component); err != nil {
		return errors.Errorf("failed to gen release filter component state, %v", err)
	}

	if event.Operation == cptype.InitializeOperation {
		if err := f.DecodeURLQuery(); err != nil {
			return errors.Errorf("failed to decode url query for release filter component, %v", err)
		}
		releaseID, _ := f.sdk.InParams["appReleaseIDs"].(string)
		if releaseID != "" {
			f.State.Values.ReleaseID = releaseID
			delete(f.sdk.InParams, "appReleasesIDs")
		}
	}

	if err := f.RenderFilter(); err != nil {
		return err
	}
	if err := f.EncodeURLQuery(); err != nil {
		return errors.Errorf("failed to encode url query for release filter component, %v", err)
	}
	f.Transfer(component)
	return nil
}

func (f *ComponentReleaseFilter) InitComponent(ctx context.Context) {
	sdk := cputil.SDK(ctx)
	f.sdk = sdk
	bdl := ctx.Value(cmpTypes.GlobalCtxKeyBundle).(*bundle.Bundle)
	f.bdl = bdl
}

func (f *ComponentReleaseFilter) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	f.State = state
	return nil
}

func (f *ComponentReleaseFilter) DecodeURLQuery() error {
	queryData, ok := f.sdk.InParams["releaseFilter__urlQuery"].(string)
	if !ok {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(queryData)
	if err != nil {
		return err
	}
	query := make(map[string]interface{})
	if err := json.Unmarshal(decoded, &query); err != nil {
		return err
	}

	sli, _ := query["applicationIDs"].([]interface{})
	var appIDs []string
	for i := range sli {
		id, _ := sli[i].(string)
		if id != "" {
			appIDs = append(appIDs, id)
		}
	}

	sli, _ = query["userIDs"].([]interface{})
	var userIDs []string
	for i := range sli {
		id, _ := sli[i].(string)
		if id != "" {
			userIDs = append(userIDs, id)
		}
	}

	sli, _ = query["createdAtStartEnd"].([]interface{})
	var timestamps []int64
	for i := range sli {
		id, _ := sli[i].(float64)
		if id > 0 {
			timestamps = append(timestamps, int64(id))
		}
	}
	f.State.Values.ApplicationIDs = appIDs
	f.State.Values.UserIDs = userIDs
	f.State.Values.CreatedAtStartEnd = timestamps
	f.State.Values.CommitID, _ = query["commitID"].(string)
	f.State.Values.BranchID, _ = query["branchID"].(string)
	f.State.Values.ReleaseID, _ = query["releaseID"].(string)
	return nil
}

func (f *ComponentReleaseFilter) EncodeURLQuery() error {
	urlQuery := make(map[string]interface{})
	urlQuery["applicationIDs"] = f.State.Values.ApplicationIDs
	urlQuery["userIDs"] = f.State.Values.UserIDs
	urlQuery["createdAtStartEnd"] = f.State.Values.CreatedAtStartEnd
	urlQuery["commitID"] = f.State.Values.CommitID
	urlQuery["branchID"] = f.State.Values.BranchID
	urlQuery["releaseID"] = f.State.Values.ReleaseID
	jsonData, err := json.Marshal(urlQuery)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	f.State.ReleaseFilterURLQuery = encoded
	return nil
}

func (f *ComponentReleaseFilter) RenderFilter() error {
	f.Data.HideSave = true
	userID := f.sdk.Identity.UserID
	orgIDStr := f.sdk.Identity.OrgID
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrInvoke.InvalidParameter(fmt.Errorf("invalid org id %s, %v", orgIDStr, err))
	}

	if !f.State.IsProjectRelease {
		appResp, err := f.bdl.GetAppsByProject(uint64(f.State.ProjectID), orgID, userID)
		if err != nil {
			return errors.Errorf("failed to list apps, %v", err)
		}
		appCondition := Condition{
			Key:         "applicationIDs",
			Label:       f.sdk.I18n("application"),
			Placeholder: f.sdk.I18n("selectApplication"),
			Type:        "select",
		}
		var appOptions []Option
		for i := range appResp.List {
			name := appResp.List[i].DisplayName
			id := appResp.List[i].ID
			appOptions = append(appOptions, Option{
				Label: name,
				Value: strconv.FormatInt(int64(id), 10),
			})
		}
		appCondition.Options = appOptions
		f.Data.Conditions = append(f.Data.Conditions, appCondition)
		f.Data.Conditions = append(f.Data.Conditions, Condition{
			Key:         "branchID",
			Label:       f.sdk.I18n("branch"),
			Placeholder: f.sdk.I18n("inputBranch"),
			Type:        "input",
		})
		f.Data.Conditions = append(f.Data.Conditions, Condition{
			Key:         "commitID",
			Label:       "commitID",
			Placeholder: f.sdk.I18n("inputCommitID"),
			Type:        "input",
		})
	}

	userCondition := Condition{
		Key:         "userIDs",
		Label:       f.sdk.I18n("creator"),
		Placeholder: f.sdk.I18n("selectCreator"),
		Type:        "select",
	}
	var userOptions []Option
	usersResp, err := f.bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   f.State.ProjectID,
		PageNo:    1,
		PageSize:  500,
	})
	if err != nil {
		return errors.Errorf("failed to list user, %v", err)
	}

	for i := range usersResp {
		userOptions = append(userOptions, Option{
			Label: usersResp[i].Nick,
			Value: usersResp[i].UserID,
		})
	}
	userCondition.Options = userOptions
	f.Data.Conditions = append(f.Data.Conditions, userCondition)
	f.Data.Conditions = append(f.Data.Conditions, Condition{
		Key:         "releaseID",
		Label:       f.sdk.I18n("releaseID"),
		Placeholder: f.sdk.I18n("inputReleaseID"),
		Type:        "input",
	})
	return nil
}

func (f *ComponentReleaseFilter) Transfer(component *cptype.Component) {
	component.Data = map[string]interface{}{
		"conditions": f.Data.Conditions,
		"hideSave":   f.Data.HideSave,
	}
	component.State = map[string]interface{}{
		"values":                  f.State.Values,
		"releaseFilter__urlQuery": f.State.ReleaseFilterURLQuery,
		"isProjectRelease":        f.State.IsProjectRelease,
		"projectID":               f.State.ProjectID,
	}
}
