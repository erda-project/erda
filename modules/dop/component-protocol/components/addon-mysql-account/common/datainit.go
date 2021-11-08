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

package common

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	addonmysqlpb "github.com/erda-project/erda-proto-go/orchestrator/addon/mysql/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/pkg/strutil"
)

type AccountData struct {
	Attachments     []*addonmysqlpb.Attachment
	AttachmentMap   map[uint64]*addonmysqlpb.Attachment
	Accounts        []*addonmysqlpb.MySQLAccount
	AccountMap      map[string]*addonmysqlpb.MySQLAccount
	AccountRefCount map[string]int
	Apps            []apistructs.ApplicationDTO
	AppMap          map[string]*apistructs.ApplicationDTO
}

func InitAccountData(ctx context.Context, instanceID string, projectID uint64) (*AccountData, error) {
	identity := cputil.GetIdentity(ctx)
	addonMySQLSvc := ctx.Value(types.AddonMySQLService).(addonmysqlpb.AddonMySQLServiceServer)
	r, err := addonMySQLSvc.ListAttachment(utils.NewContextWithHeader(ctx), &addonmysqlpb.ListAttachmentRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return nil, err
	}
	attachmentMap := make(map[uint64]*addonmysqlpb.Attachment)
	for _, attachment := range r.Attachments {
		attachmentMap[attachment.Id] = attachment
	}

	ra, err := addonMySQLSvc.ListMySQLAccount(utils.NewContextWithHeader(ctx), &addonmysqlpb.ListMySQLAccountRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return nil, err
	}
	accountMap := make(map[string]*addonmysqlpb.MySQLAccount)
	for _, a := range ra.Accounts {
		accountMap[a.Id] = a
	}

	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	orgID, err := strutil.Atoi64(identity.OrgID)
	if err != nil {
		return nil, err
	}
	rap, err := bdl.GetAppsByProject(projectID, uint64(orgID), identity.UserID)
	if err != nil {
		return nil, err
	}
	appMap := make(map[string]*apistructs.ApplicationDTO)
	for i := range rap.List {
		a := rap.List[i]
		appMap[strutil.String(a.ID)] = &a
	}

	counter := map[string]int{}
	for _, att := range r.Attachments {
		if att.AccountId != "" {
			counter[att.AccountId]++
		}
		if att.AccountState == "PRE" && att.PreAccountId != "" {
			counter[att.PreAccountId]++
		}
	}

	data := &AccountData{
		Attachments:     r.Attachments,
		AttachmentMap:   attachmentMap,
		Accounts:        ra.Accounts,
		AccountMap:      accountMap,
		AccountRefCount: counter,
		Apps:            rap.List,
		AppMap:          appMap,
	}

	ctx.Value(cptype.GlobalInnerKeyStateTemp).(map[string]interface{})["accountData"] = data

	return data, nil
}

func LoadAccountData(ctx context.Context) (*AccountData, error) {
	data, ok := ctx.Value(cptype.GlobalInnerKeyStateTemp).(map[string]interface{})["accountData"].(*AccountData)
	if !ok {
		return nil, fmt.Errorf("account data not found")
	}
	return data, nil
}

//func ClearAccountData(ctx context.Context, gs *cptype.GlobalStateData) {
//	delete(*gs, "accountData")
//}

func (d *AccountData) GetAccountName(accountID string) string {
	a, ok := d.AccountMap[accountID]
	if ok && a != nil {
		return a.Username
	}
	return accountID
}

func (d *AccountData) GetApp(appID string) *apistructs.ApplicationDTO {
	return d.AppMap[appID]
}

func (d *AccountData) GetAppName(appID string) string {
	app := d.GetApp(appID)
	if app == nil {
		return ""
	}
	return app.Name
}
